package audio

import (
	"context"
	"math"
	"time"

	"smegg.me/smeggtuner/backend/common/i18n"

	appconfig "smegg.me/smeggtuner/backend/common/config"
	"smegg.me/smeggtuner/backend/common/logger"
	coreaudio "smegg.me/smeggtuner/backend/core/audio"
	"smegg.me/smeggtuner/backend/core/dsp/meter"
)

// calibrate runs one session: every block is a clock point stamped in the capture callback, a lost
// block is a glitch, and the estimator steps once per nominal half second of device frames.
func (s *Service) calibrate(ctx context.Context, c *calibration, src coreaudio.Source, blocks <-chan coreaudio.Block, start time.Time) {
	defer close(c.done)
	defer func() { _ = src.Stop() }()
	rate := src.Info().SampleRate
	est := meter.NewClockCalibration(rate)
	bar := meter.NewClockProgress()
	half := int64(max(1, rate/2))
	next := half
	dropouts := 0
	for {
		select {
		case <-ctx.Done():
			s.abandon(c)
			return
		case b, ok := <-blocks:
			if !ok {
				s.abandon(c)
				return
			}
			at := b.Time.Sub(start).Seconds()
			if b.Lost > 0 {
				dropouts++
			}
			est.Packet(b.Frames, at, b.Lost > 0)
			for b.Frames >= next {
				next += half
				r := est.Update(at)
				st := CalibrationDTO{DeviceID: c.key, Status: CalibrationPreparing, Elapsed: int(math.Round(at)),
					Progress: bar.Next(r, at), Settled: bar.Settled(), Spread: r.Spread, Rate: rate, Dropouts: dropouts, Running: true}
				if r.Status != meter.ClockPreparing {
					st.Status, st.Percent = CalibrationCalibrating, meter.StoredCalibration(r.Live)
				}
				if r.Done() {
					s.finish(c, r, st)
					return
				}
				s.publish(c, st)
			}
		}
	}
}

// finish stores the outcome: the accepted correction, or 0 for an unreliable clock.
func (s *Service) finish(c *calibration, r meter.ClockReport, st CalibrationDTO) {
	st.Status, st.Percent, st.Running = CalibrationCalibrated, meter.StoredCalibration(r.Value), false
	if r.Status == meter.ClockUnreliable {
		st.Status, st.Percent = CalibrationUnreliable, 0
	}
	if err := s.storeDevice(c.key, func(d *appconfig.Device) {
		d.Calibration, d.CalibrationPercent = string(st.Status), st.Percent
	}); err != nil {
		s.retire(c, s.calibrationSaveFailed(c.key, err))
		return
	}
	logger.Info(i18n.KeyLogCalibrationFinished, logger.Str("device_id", c.key), logger.Str("status", string(st.Status)),
		logger.Any("percent", st.Percent), logger.Bool("timeout", r.Timeout), logger.Int("seconds", st.Elapsed))
	s.retire(c, st)
}

// abandon ends a session that was cancelled or lost its device; the device keeps what is stored.
func (s *Service) abandon(c *calibration) {
	logger.Info(i18n.KeyLogCalibrationAbandoned, logger.Str("device_id", c.key))
	s.retire(c, stored(c.key))
}

func (s *Service) publish(c *calibration, st CalibrationDTO) {
	s.calMu.Lock()
	c.state = st
	s.calMu.Unlock()
	emitEvent(EventCalibration, st)
}

func (s *Service) retire(c *calibration, st CalibrationDTO) {
	s.calMu.Lock()
	if s.cal == c {
		s.cal = nil
	}
	s.calMu.Unlock()
	emitEvent(EventCalibration, st)
}
