package report

import (
	"smegg.me/smeggtuner/backend/common/i18n"
	"smegg.me/smeggtuner/backend/core/target"
)

// InCents reports whether the sheet prints cents rather than reference units.
func (r *Report) InCents() bool { return r.Unit == target.UnitCents }

// UnitName is the sheet's deviation unit as a table heading names it.
func (r *Report) UnitName() string {
	if r.InCents() {
		return i18n.ReportUnitCentsName(i18n.Localizer())
	}
	return i18n.ReportUnitReferenceName(i18n.Localizer())
}

// UnitSymbol is the unit's short mark beside a number.
func (r *Report) UnitSymbol() string {
	if r.InCents() {
		return i18n.ReportUnitCentsSymbol(i18n.Localizer())
	}
	return i18n.ReportUnitReferenceSymbol(i18n.Localizer())
}
