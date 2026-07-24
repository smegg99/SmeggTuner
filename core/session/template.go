package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
)

// Template is a saved, reusable instrument description; it travels as a .stif.
type Template struct {
	ID   string `json:"id" gorm:"primaryKey"`
	Name string `json:"name"`

	Instrument Instrument `json:"instrument" gorm:"serializer:json"`

	// HasImage says there is a photograph without carrying the bytes; those travel over the asset server (/instruments/<id>/image) or the .stif image entry.
	HasImage bool `json:"hasImage"`

	// ImageRev bumps when the photo changes and goes in the image URL (.../image?v=<rev>) to bust the webview cache; stamped by the repository on write.
	ImageRev int64 `json:"imageRev,omitempty"`

	// Image is the photograph itself, already through PrepareImage. Kept out of the JSON on purpose.
	Image []byte `json:"-" gorm:"type:blob"`
}

// TemplateVersion is stamped into every instrument file, on the same terms as a session's.
const TemplateVersion = 1

var (
	ErrTemplateVersion = errors.New("session: unsupported instrument file version")
	ErrTemplateName    = errors.New("session: an instrument needs a name")
	ErrNoTemplate      = errors.New("session: no such instrument")
	ErrNoImage         = errors.New("session: that instrument has no photograph")
)

type templateFile struct {
	V int `json:"v"`
	Template
}

// Validate requires a name, a sound instrument, and a valid reed count.
func (t *Template) Validate() error {
	if strings.TrimSpace(t.Name) == "" {
		return ErrTemplateName
	}
	if err := t.Instrument.validate(); err != nil {
		return err
	}
	return validReeds(t.Instrument.ReedCount)
}

// ReadTemplate loads an instrument description from a legacy JSON file.
func ReadTemplate(path string) (*Template, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	t, err := decodeTemplate(data)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return t, nil
}

func decodeTemplate(data []byte) (*Template, error) {
	var f templateFile
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, err
	}
	if f.V != TemplateVersion {
		return nil, fmt.Errorf("%w: %d", ErrTemplateVersion, f.V)
	}
	if err := f.Template.Validate(); err != nil {
		return nil, err
	}

	t := f.Template
	return &t, nil
}

// FromSession makes a template out of the instrument a session describes: the model, not the one accordion.
func FromSession(s *Session, name string) *Template {
	if strings.TrimSpace(name) == "" {
		name = strings.TrimSpace(s.Instrument.Name)
	}

	i := s.Instrument
	i.Serial = "" // the serial is this accordion; the template is the model
	i.Banks = slices.Clone(i.Banks)
	i.Registers = slices.Clone(i.Registers)
	for n := range i.Registers {
		i.Registers[n].Banks = slices.Clone(i.Registers[n].Banks)
	}

	// No NameKey: this is the technician's own instrument.
	return &Template{Name: name, Instrument: i}
}
