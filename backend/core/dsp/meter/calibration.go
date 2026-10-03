package meter

// CalibrationMissing is the stored no-calibration sentinel and the read default.
const CalibrationMissing = -999999.0

const calibrationLimit = 500.0

// calibrationClamp bounds to +/-500 in native maxsd/minsd operand order, so NaN passes through.
func calibrationClamp(x float64) float64 {
	m := x
	if -calibrationLimit > x {
		m = -calibrationLimit
	}
	c := m
	if calibrationLimit < m {
		c = calibrationLimit
	}
	return c
}

// StoredCalibration clamps a clock-correction percentage and rounds it to 0.001, half away from zero.
func StoredCalibration(percent float64) float64 {
	return round3(calibrationClamp(percent))
}

// LoadCalibration validates a parsed stored percentage without requantizing; -999998 and below are missing.
func LoadCalibration(parsed float64) (float64, bool) {
	if !(parsed > -999998) {
		return 0, false
	}
	return calibrationClamp(parsed), true
}

// CalibrationFactor scales the nominal sample rate: 1 + percent/100, or 0 when calibration is missing.
func CalibrationFactor(percent float64, valid bool) float64 {
	if !valid {
		return 0
	}
	return percent/100 + 1
}
