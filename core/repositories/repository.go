// Package repositories provides data access over the datastore; repositories hold no state.
package repositories

import (
	"errors"

	"gorm.io/gorm"

	"smegg.me/smeggtuner/core/datastore"
)

// ErrNotFound is returned when an id names nothing; callers match on this, not GORM's sentinel.
var ErrNotFound = errors.New("repositories: entity not found")

func translate(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}

func db() *gorm.DB { return datastore.Get() }

// Repository is the generic CRUD base an entity repository embeds.
type Repository[T any] struct{}

func New[T any]() *Repository[T] { return &Repository[T]{} }

func (r *Repository[T]) Create(entity *T) error {
	return db().Create(entity).Error
}

func (r *Repository[T]) GetByID(id string) (*T, error) {
	var entity T
	if err := db().First(&entity, "id = ?", id).Error; err != nil {
		return nil, translate(err)
	}
	return &entity, nil
}

func (r *Repository[T]) List() ([]*T, error) {
	var entities []*T
	if err := db().Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func (r *Repository[T]) Update(entity *T) error {
	return db().Save(entity).Error
}

func (r *Repository[T]) Delete(id string) error {
	res := db().Delete(new(T), "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
