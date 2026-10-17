package audio

import (
	"context"
	"sync/atomic"

	"smegg.me/smeggtuner/backend/common/i18n"

	appconfig "smegg.me/smeggtuner/backend/common/config"
	"smegg.me/smeggtuner/backend/common/logger"
	coreaudio "smegg.me/smeggtuner/backend/core/audio"
	"smegg.me/smeggtuner/backend/core/audio/endpoint"
	"smegg.me/smeggtuner/backend/core/audio/volume"
	"smegg.me/smeggtuner/backend/core/dsp"
)

// AutoVolumeDTO is a device's automatic volume. On is what the "signal too strong" warning follows:
// it is shown only while On is false.
type AutoVolumeDTO struct {
	DeviceID  string  `json:"deviceId"`
	On        bool    `json:"on"`        // switched on and this system has a backend
	Supported bool    `json:"supported"` // this system can reach a device's input level at all
	Active    bool    `json:"active"`    // steering a running capture now
	Level     float64 `json:"level"`     // where the device starts on the next capture, 0..1
}

// steering is the automatic volume on one running capture.
type steering struct {
	in   chan volume.Interval
	stop chan struct{}
	done chan struct{}
	live atomic.Bool // the endpoint opened and the stored level is applied
}

// steeredMic is a mic whose OS input level the automatic volume may steer while it captures.
type steeredMic struct {
	*coreaudio.MicSource
	svc *Service
	key string
}

func (m *steeredMic) Start(ctx context.Context) (<-chan coreaudio.Block, error) {
	blocks, err := m.MicSource.Start(ctx)
	if err != nil {
		return nil, err
	}
	m.svc.captureStarted(m.key)
	return blocks, nil
}

func (m *steeredMic) Stop() error {
	err := m.MicSource.Stop()
	m.svc.captureStopped(m.key)
	return err
}

// SetAutoVolume switches a device's automatic volume ("" is the system default). Switching it on
// during a capture starts steering; off gives the OS its level back at once.
func (s *Service) SetAutoVolume(deviceID string, on bool) (AutoVolumeDTO, error) {
	s.volOpMu.Lock()
	defer s.volOpMu.Unlock()
	key := s.keyOf(deviceID)
	if err := s.storeDevice(key, func(d *appconfig.Device) { d.AutoVolume = on }); err != nil {
		logger.Warn(i18n.KeyLogDeviceStoreFailed, logger.Str("device_id", key), logger.Err(err))
		return s.AutoVolume(key), ErrAutoVolumeSave
	}
	s.volMu.Lock()
	var ended *steering
	if s.capturing && s.capKey == key {
		switch {
		case on && s.steer == nil:
			s.beginLocked(key)
		case !on && s.steer != nil:
			ended, s.steer = s.steer, nil
		}
	}
	s.volMu.Unlock()
	ended.end()
	st := s.AutoVolume(key)
	emitEvent(EventVolume, st)
	return st, nil
}

// AutoVolume is a device's automatic volume ("" is the system default).
func (s *Service) AutoVolume(deviceID string) AutoVolumeDTO {
	key := s.keyOf(deviceID)
	d := device(key)
	s.volMu.Lock()
	active := s.capturing && s.capKey == key && s.steer != nil && s.steer.live.Load()
	s.volMu.Unlock()
	return AutoVolumeDTO{DeviceID: key, On: d.AutoVolume && endpoint.Supported(), Supported: endpoint.Supported(),
		Active: active, Level: d.Volume}
}

// OnMeasurement takes the engine's level intervals for the automatic volume; services/tuner calls
// it, not the frontend. Every event repeats the last interval, so a new one is told by its Seq.
func (s *Service) OnMeasurement(m dsp.Measurement) {
	iv := m.Input
	s.volMu.Lock()
	st := s.steer
	fresh := st != nil && iv.Seq != 0 && iv.Seq != s.lastSeq
	if fresh {
		s.lastSeq = iv.Seq
	}
	s.volMu.Unlock()
	if !fresh {
		return
	}
	select {
	case st.in <- volume.Interval{RMS: iv.RMS, Peak: iv.Peak, Clipped: iv.Clipped, Restart: iv.Restart}:
	default:
	}
}

func (s *Service) captureStarted(key string) {
	s.volOpMu.Lock()
	defer s.volOpMu.Unlock()
	s.volMu.Lock()
	s.capKey, s.capturing, s.lastSeq = key, true, 0
	if device(key).AutoVolume && endpoint.Supported() {
		s.beginLocked(key)
	}
	s.volMu.Unlock()
}

// captureStopped gives the OS its level back before the capture counts as stopped.
func (s *Service) captureStopped(key string) {
	s.volOpMu.Lock()
	defer s.volOpMu.Unlock()
	s.volMu.Lock()
	var ended *steering
	if s.capturing && s.capKey == key {
		s.capturing = false
		ended, s.steer = s.steer, nil
	}
	s.volMu.Unlock()
	ended.end()
}

// beginLocked starts steering the capture; callers hold volMu.
func (s *Service) beginLocked(key string) {
	st := &steering{in: make(chan volume.Interval, 8), stop: make(chan struct{}), done: make(chan struct{})}
	s.steer, s.lastSeq = st, 0
	go s.steerCapture(key, st)
}

// steerCapture owns the endpoint for one capture: apply the stored level, follow the intervals,
// store where it left the device and restore the OS level and mute.
func (s *Service) steerCapture(key string, st *steering) {
	defer close(st.done)
	ep, err := s.openEndpoint(key)
	if err != nil {
		logger.Warn(i18n.KeyLogVolumeUnavailable, logger.Str("device_id", key), logger.Err(err))
		return
	}
	defer func() { _ = ep.Close() }()
	sess, err := volume.Begin(ep, device(key).Volume)
	if err != nil {
		logger.Warn(i18n.KeyLogVolumeUnavailable, logger.Str("device_id", key), logger.Err(err))
		return
	}
	st.live.Store(true)
	emitEvent(EventVolume, s.AutoVolume(key))
	for {
		select {
		case iv := <-st.in:
			if err := sess.Interval(iv); err != nil {
				logger.Debug(i18n.KeyLogVolumeStepFailed, logger.Str("device_id", key), logger.Err(err))
			}
		case <-st.stop:
			st.live.Store(false)
			if left, err := sess.Left(); err == nil {
				if err := s.storeDevice(key, func(d *appconfig.Device) { d.Volume = left }); err != nil {
					logger.Warn(i18n.KeyLogDeviceStoreFailed, logger.Str("device_id", key), logger.Err(err))
				}
			}
			if err := sess.Restore(); err != nil {
				logger.Warn(i18n.KeyLogVolumeRestore, logger.Str("device_id", key), logger.Err(err))
			}
			emitEvent(EventVolume, s.AutoVolume(key))
			return
		}
	}
}

// end stops the steering and waits until the OS has its level back; nil is a no-op.
func (st *steering) end() {
	if st == nil {
		return
	}
	close(st.stop)
	<-st.done
}
