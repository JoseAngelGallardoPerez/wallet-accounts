package limit

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"strings"
)

// Service allows to easily work with limits
type Service struct {
	storage Storage
	factory Factory
}

// NewService is Service constructor
func NewService(storage Storage, factory Factory) *Service {
	return &Service{
		storage: storage,
		factory: factory,
	}
}

// Create is used in order to create new limit
func (s *Service) Create(value Value, identifiable Identifiable) error {
	id := identifiable.Identifier()
	if err := s.ensureIdIsComplete(id); err != nil {
		return err
	}
	exist, err := s.storage.Find(id)
	if errors.Cause(err) == ErrNotFound {
		err = nil
	}
	if err != nil {
		return errors.Wrap(err, "failed to create new limit")
	}
	if len(exist) != 0 {
		return errors.Wrapf(
			ErrAlreadyExist,
			"failed to create new limit: identifiable with the same properties is already exist (name = %s, entity = %s, entityId = %s)",
			id.Name,
			id.Entity,
			id.EntityId,
		)
	}
	return s.storage.Save(value, identifiable.Identifier())
}

// UpdateOne is used in order to update single limit, it requires that identifiable to be completely filled
func (s *Service) UpdateOne(value Value, identifiable Identifiable) error {
	id := identifiable.Identifier()
	if err := s.ensureIdIsComplete(id); err != nil {
		return err
	}
	return s.storage.Update(value, identifiable.Identifier())
}

// FindOne retrieves first found identifiable
func (s *Service) FindOne(identifier Identifier) (IdentifiableLimit, error) {
	result, err := s.Find(identifier)
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, ErrNotFound
	}
	return result[0], nil
}

// Find retrieves all limits by the given identifier parameters
func (s *Service) Find(identifier Identifier) ([]IdentifiableLimit, error) {
	found, err := s.storage.Find(identifier)
	if err != nil {
		return nil, err
	}

	result := make([]IdentifiableLimit, len(found))
	for i, r := range found {
		result[i] = NewIdentifiableLimit(s.factory.CreateLimit(r), r.Identifier)
	}
	return result, nil
}

// DeleteOne is used in order to delete single limit, it requires that identifiable to be completely filled
func (s *Service) DeleteOne(identifier Identifier) error {
	if err := s.ensureIdIsComplete(identifier); err != nil {
		return err
	}
	return s.storage.Delete(identifier)
}

// DeleteByEntity deletes all limits related to the given entity
func (s *Service) DeleteByEntity(entity, id string) error {
	return s.storage.Delete(Identifier{
		Entity:   entity,
		EntityId: id,
	})
}

// DeleteByName deletes all limits by the given name
func (s *Service) DeleteByName(name string) error {
	return s.storage.Delete(Identifier{
		Name: name,
	})
}

// WrapContext makes a copy of the service with passed db
func (s Service) WrapContext(db *gorm.DB) *Service {
	if storage, ok := s.storage.(TransactionalStorage); ok {
		s.storage = storage.WrapContext(db)
	}
	return &s
}

func (s *Service) ensureIdIsComplete(id Identifier) error {
	missed := make([]string, 0, 3)
	if id.Name == "" {
		missed = append(missed, "Name")
	}
	if id.Entity == "" {
		missed = append(missed, "Entity")
	}
	if id.EntityId == "" {
		missed = append(missed, "EntityId")
	}
	if len(missed) != 0 {
		return errors.Wrapf(
			ErrIdIncomplete,
			"the following identifier properties are required: %s",
			strings.Join(missed, ", "),
		)
	}
	return nil
}
