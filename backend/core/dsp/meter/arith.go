// Package meter holds the reference-compatible measurement stages: pure
// functions and explicit state, kept in native operation order.
package meter

import "math"

// Deviation unit: 1 = 0.06 percent of the reference frequency.
const (
	unitFraction = 0.06
	Missing      = -99.0
	InvalidBeat  = -1.0
)

// Conversions such as float64(a*b) stop the compiler fusing a*b+c into FMA.

func ReferenceFrequency(table, offset float64) float64 { return table - offset }

func FrequencyHz(reference, deviation float64) float64 {
	shift := float64(float64(deviation*reference)*unitFraction) / 100
	return shift + reference
}

func DeviationUnits(bin, referenceBin float64) float64 {
	return ((bin - referenceBin) * 100) / float64(referenceBin*unitFraction)
}

// Interpolate is the three-bin peak offset: binary32 differences, binary64 ratio.
func Interpolate(left, center, right float32) float64 {
	rise, fall := float64(center-left), float64(center-right)
	return rise/(fall+rise) - 0.5
}

func Residual(deviation, target float64) float64 {
	if deviation == Missing || target == Missing {
		return deviation
	}
	return deviation - target
}

func TargetPair(first, second, reference float64, coverage, hertz bool) float64 {
	if !coverage || first <= Missing || second <= Missing {
		return Missing
	}
	difference := second - first
	if !hertz {
		return difference
	}
	return float64(float64(difference*reference)*unitFraction) / 100
}

func SubtractPairTarget(value, target float64, scaleRelative, present bool) float64 {
	if !present {
		return Missing
	}
	if scaleRelative || target <= Missing {
		return value
	}
	return value - target
}

// HarmonicBeat keeps the reciprocal-period form; zero input or a non-unison cents pair is InvalidBeat.
func HarmonicBeat(a, b, reference float64, hertz bool) float64 {
	if a == 0 || b == 0 {
		return InvalidBeat
	}
	low, high := a, b
	if low > high {
		low, high = high, low
	}
	k := int32(math.Trunc(high/low + 0.5))
	period := (1 / low) / float64(k)
	other := 1 / high
	beat := 1 / float64((period/math.Abs(period-other))*other)
	if hertz {
		return beat
	}
	if k != 1 {
		return InvalidBeat
	}
	return (beat * 100) / float64(reference*unitFraction)
}

// QuantizeReadout rounds half away from zero at the panel precision, through binary32.
func QuantizeReadout(value float64, decimals int, signed bool) (float32, bool) {
	scale := [3]float64{1, 10, 100}[decimals]
	scaled := value
	if decimals > 0 {
		scaled = float64(value * scale)
	}
	// The native path converts through int32, so a negative fraction rounding to zero prints as +0.
	var integer int32
	if scaled >= 0 {
		integer = int32(scaled + 0.5)
	} else {
		integer = int32(scaled - 0.5)
	}
	quantized := float32(integer)
	if decimals > 0 {
		quantized = float32(float64(quantized) / scale)
	}
	return quantized, signed && math.Abs(value) >= [3]float64{0.5, 0.05, 0.005}[decimals]
}
