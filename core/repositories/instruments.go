package repositories

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"smegg.me/smeggtuner/core/session"
)

// InstrumentRepository owns the templates table; every read except Image omits the image blob, which is served separately (services/session/images.go).
type InstrumentRepository struct{ *Repository[session.Template] }

func GetInstrumentRepository() *InstrumentRepository {
	return &InstrumentRepository{Repository: New[session.Template]()}
}

// descriptionColumns is every column except the image blob.
var descriptionColumns = []string{"id", "name", "instrument", "has_image", "image_rev"}

// List returns every instrument sorted by name.
func (r *InstrumentRepository) List() ([]session.Template, error) {
	var all []session.Template
	if err := db().Select(descriptionColumns).Find(&all).Error; err != nil {
		return nil, err
	}
	sort.Slice(all, func(i, j int) bool { return all[i].Name < all[j].Name })
	return all, nil
}

// Get returns one instrument by id, without the image blob.
func (r *InstrumentRepository) Get(id string) (*session.Template, error) {
	if !session.ValidID(id) {
		return nil, fmt.Errorf("%w: %q", session.ErrNoTemplate, id)
	}
	var t session.Template
	err := db().Select(descriptionColumns).First(&t, "id = ?", id).Error
	if errors.Is(translate(err), ErrNotFound) {
		return nil, fmt.Errorf("%w: %q", session.ErrNoTemplate, id)
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// Save writes only the description (not the image); a template with no ID gets one and is refreshed from the row.
func (r *InstrumentRepository) Save(t *session.Template) error {
	if err := t.Validate(); err != nil {
		return err
	}
	if t.ID == "" {
		t.ID = session.NewID()
	}
	if !session.ValidID(t.ID) {
		return fmt.Errorf("%w: %q", session.ErrBadID, t.ID)
	}

	res := db().Model(&session.Template{}).Where("id = ?", t.ID).
		Select("name", "instrument").Updates(t)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		fresh := session.Template{ID: t.ID, Name: t.Name, Instrument: t.Instrument}
		if err := r.Create(&fresh); err != nil {
			return err
		}
	}

	saved, err := r.Get(t.ID)
	if err != nil {
		return err
	}
	*t = *saved
	return nil
}

// Delete removes an instrument and its image.
func (r *InstrumentRepository) Delete(id string) error {
	if !session.ValidID(id) {
		return fmt.Errorf("%w: %q", session.ErrNoTemplate, id)
	}
	err := r.Repository.Delete(id)
	if errors.Is(err, ErrNotFound) {
		return fmt.Errorf("%w: %q", session.ErrNoTemplate, id)
	}
	return err
}

// SetImage sets an instrument's image (nil removes it); jpg must already be PrepareImage'd, and the bumped revision is what makes a swapped image appear via its URL (session.Template.ImageRev).
func (r *InstrumentRepository) SetImage(id string, jpg []byte) error {
	if _, err := r.Get(id); err != nil {
		return err
	}

	values := map[string]any{"image": nil, "has_image": false, "image_rev": 0}
	if jpg != nil {
		values = map[string]any{"image": jpg, "has_image": true, "image_rev": time.Now().UnixNano()}
	}
	return db().Model(&session.Template{}).Where("id = ?", id).Updates(values).Error
}

// Image returns an instrument's image and its revision.
func (r *InstrumentRepository) Image(id string) ([]byte, int64, error) {
	if !session.ValidID(id) {
		return nil, 0, fmt.Errorf("%w: %q", session.ErrNoTemplate, id)
	}
	var t session.Template
	err := db().Select("id", "has_image", "image_rev", "image").First(&t, "id = ?", id).Error
	if errors.Is(translate(err), ErrNotFound) {
		return nil, 0, fmt.Errorf("%w: %q", session.ErrNoTemplate, id)
	}
	if err != nil {
		return nil, 0, err
	}
	if !t.HasImage || len(t.Image) == 0 {
		return nil, 0, fmt.Errorf("%w: %q", session.ErrNoImage, id)
	}
	return t.Image, t.ImageRev, nil
}

// ImportFile reads a .stif and saves it under a NEW id, so an import never overwrites an existing instrument.
func (r *InstrumentRepository) ImportFile(path string) (*session.Template, error) {
	t, jpg, err := session.ReadInstrumentFile(path)
	if err != nil {
		return nil, err
	}

	t.ID = ""
	if err := r.Save(t); err != nil {
		return nil, err
	}
	if jpg != nil {
		if err := r.SetImage(t.ID, jpg); err != nil {
			return nil, err
		}
		saved, err := r.Get(t.ID)
		if err != nil {
			return nil, err
		}
		*t = *saved
	}
	return t, nil
}

// ExportFile writes an instrument to a .stif, image included.
func (r *InstrumentRepository) ExportFile(id, path string) error {
	t, err := r.Get(id)
	if err != nil {
		return err
	}

	// A missing image is not an error here.
	jpg, _, err := r.Image(id)
	if err != nil && !errors.Is(err, session.ErrNoImage) {
		return err
	}
	return session.WriteInstrumentFile(path, t, jpg)
}
