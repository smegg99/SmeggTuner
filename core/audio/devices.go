// core/audio/devices.go
package audio

import (
	"encoding/hex"

	"github.com/gen2brain/malgo"
)

// DeviceInfo describes one capture device for the UI layer.
type DeviceInfo struct {
	ID      string `json:"id"` // hex of malgo device ID bytes; stable across runs
	Name    string `json:"name"`
	Default bool   `json:"default"`
}

func newMalgoContext() (*malgo.AllocatedContext, error) {
	return malgo.InitContext(nil, malgo.ContextConfig{}, nil)
}

// Devices enumerates the available capture devices.
func Devices() ([]DeviceInfo, error) {
	ctx, err := newMalgoContext()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = ctx.Uninit()
		ctx.Free()
	}()
	infos, err := ctx.Devices(malgo.Capture)
	if err != nil {
		return nil, err
	}
	out := make([]DeviceInfo, 0, len(infos))
	for i := range infos {
		di := &infos[i]
		out = append(out, DeviceInfo{
			ID:      hex.EncodeToString(di.ID[:]),
			Name:    di.Name(),
			Default: di.IsDefault != 0,
		})
	}
	return out, nil
}
