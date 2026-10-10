// backend/core/dsp/detection.go
package dsp

import "smegg.me/smeggtuner/backend/core/dsp/meter"

// Detection trades how soon a new note reads and locks against how sure the engine is before it locks.
type Detection string

const (
	DetectionFast     Detection = "fast"
	DetectionBalanced Detection = "balanced"
	DetectionSteady   Detection = "steady" // the reference: Dirk's pace and nothing added
)

// detection is what a Detection sets.
type detection struct {
	perSecond int  // analyses a second; native events stay at meter.UpdatesPerSecond
	settle    int  // native picks before a lock (History.Settle)
	watch     int  // quiet short-window passes before a note is let go; 0 never watches
	extended  bool // Release, QuietSeed, Follow and Fresh
}

var detections = map[Detection]detection{
	DetectionFast:     {perSecond: 8, settle: 2, watch: 2, extended: true},
	DetectionBalanced: {perSecond: 8, settle: 4, watch: 3, extended: true},
	DetectionSteady:   {perSecond: meter.UpdatesPerSecond},
}

// detection resolves the preset; a file has no learned background, and a mains line its selection names
// for the first two seconds would lock under fewer than four picks, so files never settle faster.
func (c *EngineConfig) detection() detection {
	d := detections[c.Detection]
	if d.extended && !c.LearnOnStart {
		d.settle = max(d.settle, 4)
	}
	return d
}
