package tuner

import (
	"context"
	"time"

	coreaudio "smegg.me/smeggtuner/core/audio"
)

const playheadHz = 30

// followPlayhead emits the playhead until its run ends, including while paused so a seek still moves the needle.
func followPlayhead(ctx context.Context, t coreaudio.Transport) {
	tick := time.NewTicker(time.Second / playheadHz)
	defer tick.Stop()

	last := PlaybackDTO{Position: -1}
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			now := PlaybackDTO{
				Position: t.Position().Seconds(),
				Paused:   t.Paused(),
				Moving:   t.Moving(),
			}
			if now == last {
				continue
			}
			last = now
			emitEvent(EventPlayback, now)
		}
	}
}
