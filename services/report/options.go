package report

import (
	"errors"
	"strings"
	"time"

	appconfig "smegg.me/smeggtuner/common/config"
	"smegg.me/smeggtuner/common/logger"
	corereport "smegg.me/smeggtuner/core/report"
	"smegg.me/smeggtuner/core/target"
	"smegg.me/smeggtuner/core/tuning"
	sessionsvc "smegg.me/smeggtuner/services/session"
)

// options maps the dialog answers and config into core/report's; tolerances come from the tuner screen, not recomputed here.
func (s *Service) options(opts OptionsDTO) (corereport.Options, error) {
	cfg := appconfig.Get().Tuner
	tol, beatTol := target.Tolerances(cfg.Tolerance, cfg.BeatTolerance)
	out := corereport.Options{
		Naming:        tuning.ParseNaming(cfg.ScaleNaming),
		Tolerance:     tol,
		BeatTolerance: beatTol,
	}

	if d := strings.TrimSpace(opts.Date); d != "" {
		when, err := time.Parse(dateLayout, d)
		if err != nil {
			logger.Warn(logger.MsgReportRejected, logger.Str("date", d))
			return out, ErrInvalidDate
		}
		out.Date = when
	}

	// The letterhead is a config setting, used whenever any field is filled; the dialog doesn't control it.
	if opts.Format == FormatCSV {
		return out, nil
	}

	head := appconfig.Get().Report
	if head.CompanyName == "" && head.CompanyAddress == "" && head.CompanyWebsite == "" && head.LogoPath == "" {
		return out, nil
	}

	// A logo that will not load is logged and left out, not a reason to refuse the card.
	logo, err := corereport.LoadLogo(head.LogoPath)
	if err != nil {
		logger.Warn(logger.MsgReportRejected, logger.Str("logo", head.LogoPath), logger.Err(err))
		logo = ""
	}
	out.Letterhead = &corereport.Letterhead{
		CompanyName:    head.CompanyName,
		CompanyAddress: head.CompanyAddress,
		CompanyWebsite: head.CompanyWebsite,
		Logo:           logo,
	}
	return out, nil
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
