//go:build linux && !android

// backend/core/audio/endpoint/pulse_linux.go
package endpoint

import (
	"bytes"
	"fmt"
	"math"
	"net"

	"github.com/jfreymuth/pulse/proto"
)

const supported = true

// pulseSource is one PulseAudio source; malgo's pulse device id is the source name.
type pulseSource struct {
	client *proto.Client
	conn   net.Conn
	name   string
}

// open connects to the session's pulse server (PULSE_SERVER, else the runtime directory's socket).
func open(raw []byte) (Endpoint, error) {
	if i := bytes.IndexByte(raw, 0); i >= 0 {
		raw = raw[:i]
	}
	client, conn, err := proto.Connect("")
	if err != nil {
		return nil, fmt.Errorf("pulse connect: %w", err)
	}
	p := &pulseSource{client: client, conn: conn, name: string(raw)}
	props := proto.PropList{"application.name": proto.PropListString("SmeggTuner")}
	if err := client.Request(&proto.SetClientName{Props: props}, &proto.SetClientNameReply{}); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("pulse client name: %w", err)
	}
	if p.name == "" {
		var info proto.GetServerInfoReply
		if err := client.Request(&proto.GetServerInfo{}, &info); err != nil {
			_ = conn.Close()
			return nil, fmt.Errorf("pulse server info: %w", err)
		}
		p.name = info.DefaultSourceName
	}
	if _, err := p.info(); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return p, nil
}

func (p *pulseSource) info() (proto.GetSourceInfoReply, error) {
	var info proto.GetSourceInfoReply
	if err := p.client.Request(&proto.GetSourceInfo{SourceIndex: proto.Undefined, SourceName: p.name}, &info); err != nil {
		return info, fmt.Errorf("pulse source %q: %w", p.name, err)
	}
	return info, nil
}

// Level is the loudest channel, 1 at the server's normal volume.
func (p *pulseSource) Level() (float64, error) {
	info, err := p.info()
	if err != nil {
		return 0, err
	}
	return float64(loudest(info.ChannelVolumes)) / float64(proto.VolumeNorm), nil
}

// SetLevel moves the loudest channel to level and the others with it, so the balance stays.
func (p *pulseSource) SetLevel(level float64) error {
	info, err := p.info()
	if err != nil {
		return err
	}
	target, top := float64(proto.NormVolume(level)), float64(loudest(info.ChannelVolumes))
	vols := make(proto.ChannelVolumes, len(info.ChannelVolumes))
	for i, v := range info.ChannelVolumes {
		if top == 0 {
			vols[i] = proto.Volume(target)
			continue
		}
		vols[i] = proto.Volume(math.Round(float64(v) * target / top))
	}
	req := &proto.SetSourceVolume{SourceIndex: proto.Undefined, SourceName: p.name, ChannelVolumes: vols}
	if err := p.client.Request(req, nil); err != nil {
		return fmt.Errorf("pulse set volume %q: %w", p.name, err)
	}
	return nil
}

func (p *pulseSource) Muted() (bool, error) {
	info, err := p.info()
	return info.Mute, err
}

func (p *pulseSource) SetMuted(muted bool) error {
	req := &proto.SetSourceMute{SourceIndex: proto.Undefined, SourceName: p.name, Mute: muted}
	if err := p.client.Request(req, nil); err != nil {
		return fmt.Errorf("pulse set mute %q: %w", p.name, err)
	}
	return nil
}

func (p *pulseSource) Close() error { return p.conn.Close() }

func loudest(vols proto.ChannelVolumes) proto.Volume {
	var top proto.Volume
	for _, v := range vols {
		top = max(top, v)
	}
	return top
}
