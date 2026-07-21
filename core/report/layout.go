package report

const (
	// maxPortraitColumns fits across 190 mm of A4 portrait at 8 pt. Sixteen is the musette three.
	maxPortraitColumns = 16
	// maxLandscapeColumns is the same across 277 mm at 7 pt. Twenty-eight reaches five reeds.
	maxLandscapeColumns = 28
)

// Layout is how the sheet is put on the paper.
type Layout struct {
	// Landscape turns the page.
	Landscape bool
	// Grouped uses one narrow table per reed and beat pair, fitting any reed count.
	Grouped bool
}

func layoutFor(columns int) Layout {
	switch {
	case columns <= maxPortraitColumns:
		return Layout{}
	case columns <= maxLandscapeColumns:
		return Layout{Landscape: true}
	default:
		return Layout{Grouped: true}
	}
}

// Class is the layout name the stylesheet keys off; the @page rule stays a template literal so html/template can vet the CSS.
func (l Layout) Class() string {
	switch {
	case l.Grouped:
		return "grouped"
	case l.Landscape:
		return "wide landscape"
	default:
		return "wide portrait"
	}
}
