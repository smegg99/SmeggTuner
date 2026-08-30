package report

import (
	"fmt"

	"smegg.me/smeggtuner/backend/common/i18n"
)

// BassPart is the bass side of the sheet. It exposes the same surface the table templates read off
// the Report, so "wide" and "grouped" render either without knowing which keyboard they hold.
type BassPart struct {
	// Feet names the columns (32', 16'...), largest first; empty when the machine was never
	// declared and the columns are numbered instead.
	Feet []int
	// Reeds is the columns: 1..the machine's voice count.
	Reeds         []int
	MultiRegister bool
	Pairs         []Pair
	Rows          []Row
}

// Head is a rank column's heading: the foot, or the translated reed number.
func (p *BassPart) Head(i int) string {
	if i >= 0 && i < len(p.Feet) {
		return fmt.Sprintf("%d'", p.Feet[i])
	}
	return i18n.ReportHeadReed(i18n.Localizer(), i18n.ReportHeadReedParams{Index: i + 1})
}

// VoiceColumns is the columns needed to label a reading, zero unless multi-register.
func (p *BassPart) VoiceColumns() int {
	if p.MultiRegister {
		return 1
	}
	return 0
}
