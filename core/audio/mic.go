// core/audio/mic.go
package audio

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"sync/atomic"
	"time"

	"github.com/gen2brain/malgo"

	"smegg.me/smeggtuner/common/logger"
)

const (
	msgMicBlocksDropped logger.MessageID = "mic consumer stalled, dropping blocks"
	msgMicDeviceLost    logger.MessageID = "capture device stopped unexpectedly"
)

var (
	_ Source      = (*MicSource)(nil)
	_ ErrorSource = (*MicSource)(nil)
)

// MicSource captures mono float32 audio from a capture device. Blocks are dropped (never
// queued unbounded) when the consumer stalls. A device that stops on its own closes the block
// channel with Err reporting ErrDeviceLost, so the service can retry.
type MicSource struct {
	deviceID string
	// sampleRate is written by Start and read by the malgo callback on the audio thread, so
	// it is atomic.
	sampleRate atomic.Int64
	dropped    atomic.Int64
	lost       atomic.Bool
	closing    atomic.Bool
	cancel     context.CancelFunc
}

// NewMicSource prepares a mic source for the given device ID from Devices(). An empty ID
// selects the default capture device.
func NewMicSource(deviceID string) (*MicSource, error) {
	return &MicSource{deviceID: deviceID}, nil
}

func (s *MicSource) Info() SourceInfo {
	name := "mic:default"
	if s.deviceID != "" {
		name = "mic:" + s.deviceID
	}
	return SourceInfo{Name: name, SampleRate: int(s.sampleRate.Load()), Realtime: true}
}

// Err reports why the block channel closed: ErrDeviceLost if the device stopped on its own,
// nil if we stopped it. Only meaningful once the channel has closed.
func (s *MicSource) Err() error {
	if s.lost.Load() {
		return ErrDeviceLost
	}
	return nil
}

func (s *MicSource) Start(ctx context.Context) (<-chan Block, error) {
	mctx, err := newMalgoContext()
	if err != nil {
		return nil, err
	}

	cfg := malgo.DefaultDeviceConfig(malgo.Capture)
	cfg.Capture.Format = malgo.FormatF32
	cfg.Capture.Channels = 1
	cfg.SampleRate = 0 // native rate
	if s.deviceID != "" {
		raw, err := hex.DecodeString(s.deviceID)
		if err != nil {
			_ = mctx.Uninit()
			mctx.Free()
			return nil, fmt.Errorf("bad device id: %w", err)
		}
		var id malgo.DeviceID
		copy(id[:], raw)
		cfg.Capture.DeviceID = id.Pointer()
	}

	ch := make(chan Block, 8)
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.lost.Store(false)
	s.closing.Store(false)

	callbacks := malgo.DeviceCallbacks{
		Data: func(_, input []byte, frameCount uint32) {
			samples := make([]float32, frameCount)
			for i := range samples {
				samples[i] = math.Float32frombits(binary.LittleEndian.Uint32(input[i*4:]))
			}
			select {
			case ch <- Block{
				Samples:    samples,
				SampleRate: int(s.sampleRate.Load()),
				Time:       time.Now(),
			}:
			default:
				if n := s.dropped.Add(1); n%100 == 1 {
					logger.Debug(msgMicBlocksDropped, logger.Int64("dropped_blocks", n))
				}
			}
		},
		// Stop fires when the device stops for any reason. malgo drops the callback
		// before uninit, so a stop we asked for never reaches here; anything that does
		// is the device going away. Cancelling hands teardown to the watcher goroutine,
		// the only place that may close ch (Uninit from this callback would deadlock).
		Stop: func() {
			if s.closing.Load() {
				return
			}
			s.lost.Store(true)
			logger.Warn(msgMicDeviceLost, logger.Str("device", s.Info().Name))
			cancel()
		},
	}

	dev, err := malgo.InitDevice(mctx.Context, cfg, callbacks)
	if err != nil {
		s.cancel()
		_ = mctx.Uninit()
		mctx.Free()
		return nil, err
	}
	s.sampleRate.Store(int64(dev.SampleRate()))
	if err := dev.Start(); err != nil {
		s.cancel()
		dev.Uninit()
		_ = mctx.Uninit()
		mctx.Free()
		return nil, err
	}
	go func() {
		<-ctx.Done()
		s.closing.Store(true)
		// Uninit joins the audio thread, so no Data callback runs after it returns; only
		// then is ch safe to close.
		dev.Uninit()
		_ = mctx.Uninit()
		mctx.Free()
		close(ch)
	}()
	return ch, nil
}

func (s *MicSource) Stop() error {
	if s.cancel != nil {
		s.cancel()
	}
	return nil
}
