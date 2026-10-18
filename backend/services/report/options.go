package report

import (
	"errors"
	"strings"
	"time"

	"smegg.me/smeggtuner/backend/common/i18n"

	appconfig "smegg.me/smeggtuner/backend/common/config"
	"smegg.me/smeggtuner/backend/common/logger"
	corereport "smegg.me/smeggtuner/backend/core/report"
	"smegg.me/smeggtuner/backend/core/target"
	"smegg.me/smeggtuner/backend/core/tuning"
	sessionsvc "smegg.me/smeggtuner/backend/services/session"
)

// options maps the dialog answers and config into core/report's; tolerances come from the tuner screen, not recomputed here.
func (s *Service) options(opts OptionsDTO) (corereport.Options, error) {
	cfg := appconfig.Get().Tuner
	tol, beatTol := target.Tolerances(cfg.Tolerance, cfg.BeatTolerance)
	out := corereport.Options{
		Naming:        tuning.ParseNaming(cfg.ScaleNaming),
		Tolerance:     tol,
		BeatTolerance: beatTol,
		Unit:          sheetUnit(cfg.DeviationUnit),
	}

	if d := strings.TrimSpace(opts.Date); d != "" {
		when, err := time.Parse(dateLayout, d)
		if err != nil {
			logger.Warn(i18n.KeyLogReportRejected, logger.Str("date", d))
			return out, ErrInvalidDate
		}
		out.Date = when
	}

	// The letterhead is a config setting, used whenever any field is filled; the dialog doesn't control it.
	if opts.Format == FormatCSV {
		return out, nil
	}

	head := appconfig.Get().Report
	if head.CompanyName == "" && head.CompanyAddress == "" && head.CompanyWebsite == "" {
		return out, nil
	}

	out.Letterhead = &corereport.Letterhead{
		CompanyName:    head.CompanyName,
		CompanyAddress: head.CompanyAddress,
		CompanyWebsite: head.CompanyWebsite,
	}
	return out, nil
}

// sheetUnit is the report's deviation unit for the tuner.deviation_unit setting.
func sheetUnit(setting string) target.Unit {
	if setting == "cents" {
		return target.UnitCents
	}
	return target.UnitReference
}

// buildError translates core/report's refusals into the session service's keys.
func buildError(err error) error {
	switch {
	case errors.Is(err, corereport.ErrNoSession):
		return sessionsvc.ErrNoSession
	case errors.Is(err, corereport.ErrNoReadings):
		return sessionsvc.ErrNoReadings
	default:
		return ErrRenderFailed
	}
}
