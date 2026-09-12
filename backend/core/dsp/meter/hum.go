package meter

// HumRegion is the attenuated neighbourhood: bins Center-d-1 and Center+d for every d in 0..Width.
type HumRegion struct {
	Center, Width int
}

// NewHumRegion places a hum frequency on count magnitude bins in binary32; false leaves nothing to attenuate.
func NewHumRegion(count, sampleRate int, rateFactor float64, frequency float32) (HumRegion, bool) {
	rate := float32(float64(sampleRate) * rateFactor)
	product := float32(float32(count) * frequency)
	coordinate := (product + product) / rate
	rounded := coordinate + 0.5
	if !(rounded >= 1 && rounded < float32(count)) {
		return HumRegion{}, false
	}
	center := int(rounded)
	width := max(1, int(float32(coordinate*float32(0.1))+0.5))
	width = min(width, center, count-center-1)
	// The native left index would reach -1 here; that write is outside the admitted domain.
	if center-width-1 < 0 {
		return HumRegion{}, false
	}
	return HumRegion{Center: center, Width: width}, true
}

// AttenuateHum divides the hum region in place by 20/(1+d/8)+1; with both filters apply 50 then 60.
func AttenuateHum(magnitudes []float32, sampleRate int, rateFactor float64, frequency float32) {
	r, ok := NewHumRegion(len(magnitudes), sampleRate, rateFactor, frequency)
	if !ok {
		return
	}
	for d := range r.Width + 1 {
		scaled := float32(float32(d)*0.125) + 1
		divisor := 20/scaled + 1
		magnitudes[r.Center-d-1] /= divisor
		magnitudes[r.Center+d] /= divisor
	}
}
