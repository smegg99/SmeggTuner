package audio

import (
	"context"
	"time"

	"smegg.me/smeggtuner/backend/common/i18n"

	appconfig "smegg.me/smeggtuner/backend/common/config"
	"smegg.me/smeggtuner/backend/common/logger"
	"smegg.me/smeggtuner/backend/core/dsp/meter"
)

// CalibrationStatus is where a capture device's sound-card clock calibration stands.
type CalibrationStatus string

const (
	CalibrationUncalibrated CalibrationStatus = "uncalibrated" // never measured or abandoned: no correction
	CalibrationPreparing    CalibrationStatus = "preparing"    // a session is warming the device up
	CalibrationCalibrating  CalibrationStatus = "calibrating"  // a session is measuring
	CalibrationCalibrated   CalibrationStatus = "calibrated"
	CalibrationUnreliable   CalibrationStatus = "unreliable" // the clock never settled or audio was lost: 0 stored, tuning uncorrected
)

// CalibrationDTO is one device's calibration as the audio settings show it.
type CalibrationDTO struct {
	DeviceID string            `json:"deviceId"`
	Status   CalibrationStatus `json:"status"`
	Percent  float64           `json:"percent"`  // the stored correction; while calibrating, the live rate difference
	Elapsed  int               `json:"elapsed"`  // whole seconds the running session has taken
	Progress int               `json:"progress"` // the progress bar, 0..100
	Settled  bool              `json:"settled"`  // the second half of the bar: the reading holds still and is being confirmed
	Spread   float64           `json:"spread"`   // range of the clock estimates over the last 5 s, percent
	Rate     int               `json:"rate"`     // the device's nominal sample rate, Hz
	Dropouts int               `json:"dropouts"` // capture blocks lost so far
	Running  bool              `json:"running"`
	Error    string            `json:"error"` // i18n key for a failed persistence attempt
}

// CorrectionDTO is the clock correction the engine runs with for the current input.
type CorrectionDTO struct {
	Percent float64 `json:"percent"`
	Missing bool    `json:"missing"` // the mic's device was never calibrated: measured uncorrected
}

// calibration is one running session.
type calibration struct {
	key    string
	cancel context.CancelFunc
	done   chan struct{}
	state  CalibrationDTO // guarded by Service.calMu
}

// StartCalibration measures a capture device's sample clock against the computer's ("" is the system
// default); no sound is needed. Like the reference it forgets the device's stored correction first,
// so a session that does not finish leaves it uncalibrated. A running session, on any device, ends.
// Progress arrives as EventCalibration.
func (s *Service) StartCalibration(deviceID string) error {
	s.calOpMu.Lock()
	defer s.calOpMu.Unlock()
	s.cancelCalibration()

	key := s.keyOf(deviceID)
	src, err := s.openCapture(key)
	if err != nil {
		logger.Warn(i18n.KeyLogCalibrationNoDevice, logger.Str("device_id", key), logger.Err(err))
		return ErrDeviceGone
	}
	start := time.Now()
	ctx, cancel := context.WithCancel(context.Background())
	blocks, err := src.Start(ctx)
	if err != nil {
		cancel()
		logger.Warn(i18n.KeyLogCalibrationNoDevice, logger.Str("device_id", key), logger.Err(err))
		emitEvent(EventCalibration, stored(key))
		return ErrDeviceGone
	}
	if err := s.storeDevice(key, func(d *appconfig.Device) {
		d.Calibration, d.CalibrationPercent = string(CalibrationUncalibrated), meter.CalibrationMissing
	}); err != nil {
		cancel()
		_ = src.Stop()
		emitEvent(EventCalibration, s.calibrationSaveFailed(key, err))
		return ErrCalibrationSave
	}
	c := &calibration{key: key, cancel: cancel, done: make(chan struct{})}
	c.state = CalibrationDTO{DeviceID: key, Status: CalibrationPreparing, Progress: meter.NewClockProgress().Shown(),
		Rate: src.Info().SampleRate, Running: true}
	s.calMu.Lock()
	delete(s.calErrors, key)
	s.cal = c
	s.calMu.Unlock()
	emitEvent(EventCalibration, c.state)
	logger.Info(i18n.KeyLogCalibrationStarted, logger.Str("device_id", key), logger.Int("rate", src.Info().SampleRate))
	go s.calibrate(ctx, c, src, blocks, start)
	return nil
}

// CancelCalibration ends a running session and waits for it; its device stays uncalibrated.
func (s *Service) CancelCalibration() {
	s.calOpMu.Lock()
	defer s.calOpMu.Unlock()
	s.cancelCalibration()
}

func (s *Service) cancelCalibration() {
	s.calMu.Lock()
	c := s.cal
	s.calMu.Unlock()
	if c != nil {
		c.cancel()
		<-c.done
	}
}

// CalibrationState is a device's calibration ("" is the system default): a running session's live
// state, otherwise what is stored.
func (s *Service) CalibrationState(deviceID string) CalibrationDTO {
	key := s.keyOf(deviceID)
	s.calMu.Lock()
	defer s.calMu.Unlock()
	if s.cal != nil && s.cal.key == key {
		return s.cal.state
	}
	st := stored(key)
	st.Error = s.calErrors[key]
	return st
}

// Correction is the clock correction for the current input: a file needs none, a mic its device's
// stored value. services/tuner maps it into the engine.
func (s *Service) Correction() CorrectionDTO {
	if s.Current().Kind != SourceMic {
		return CorrectionDTO{}
	}
	p, ok := meter.LoadCalibration(device(s.micKey()).CalibrationPercent)
	if !ok {
		return CorrectionDTO{Missing: true}
	}
	return CorrectionDTO{Percent: p}
}

// stored is a device's calibration as the config keeps it.
func stored(key string) CalibrationDTO {
	d := device(key)
	p, ok := meter.LoadCalibration(d.CalibrationPercent)
	if !ok {
		p = 0
	}
	return CalibrationDTO{DeviceID: key, Status: CalibrationStatus(d.Calibration), Percent: p}
}

// calibrationSaveFailed keeps the persisted correction visible and remembers the failed attempt.
func (s *Service) calibrationSaveFailed(key string, err error) CalibrationDTO {
	logger.Warn(i18n.KeyLogDeviceStoreFailed, logger.Str("device_id", key), logger.Err(err))
	s.calMu.Lock()
	defer s.calMu.Unlock()
	if s.calErrors == nil {
		s.calErrors = make(map[string]string)
	}
	s.calErrors[key] = ErrCalibrationSave.Key
	st := stored(key)
	st.Error = ErrCalibrationSave.Key
	return st
}
