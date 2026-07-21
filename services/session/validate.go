package session

import (
	"errors"
	"math"

	"smegg.me/smeggtuner/common/logger"
	"smegg.me/smeggtuner/core/repositories"
	coresession "smegg.me/smeggtuner/core/session"
	"smegg.me/smeggtuner/core/target"
)

// A4 range; outside 430..450 is a typo. defaultA4 is inherited when an instrument names none.
const (
	minA4     = 430.0
	maxA4     = 450.0
	defaultA4 = 440.0
)

func validName(name string) error {
	if name == "" {
		return ErrInvalidName
	}
	return nil
}

func validA4(hz float64) error {
	if math.IsNaN(hz) || hz < minA4 || hz > maxA4 {
		logger.Warn(logger.MsgSessionRejected, logger.Str("setting", "a4"), logger.Any("value", hz))
		return ErrInvalidA4
	}
	return nil
}

// validReeds accepts what an instrument may sound, not the engine's musette cap (that is services/tuner's).
func validReeds(n int) error {
	if n < coresession.MinReeds || n > coresession.MaxReeds {
		logger.Warn(logger.MsgSessionRejected, logger.Str("setting", "reed_count"), logger.Int("value", n))
		return ErrInvalidReedCount
	}
	return nil
}

func parseUnit(unit string) (target.Unit, error) {
	switch target.Unit(unit) {
	case target.UnitCents:
		return target.UnitCents, nil
	case target.UnitHz:
		return target.UnitHz, nil
	default:
		return "", ErrInvalidUnit
	}
}

func loadError(err error) error {
	if errors.Is(err, repositories.ErrNotFound) || errors.Is(err, coresession.ErrBadID) || isNotExist(err) {
		return ErrNotFound
	}
	return ErrLoadFailed
}
