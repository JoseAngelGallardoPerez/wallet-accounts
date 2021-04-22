package transfer

import "github.com/pkg/errors"

// performerGroup is used in order to group number of performers
type performerGroup struct {
	performers []Performer
	performed  bool
}

// NewPerformerGroup groups a number of performers
func NewPerformerGroup(performers ...Performer) Performer {
	return &performerGroup{performers: performers, performed: false}
}

// Perform executes all performers
func (p *performerGroup) Perform() error {
	if p.performed {
		return errors.Wrapf(
			ErrAlreadyPerformed,
			"the group has already been performed",
		)
	}
	p.performed = true
	for _, performer := range p.performers {
		if err := performer.Perform(); err != nil {
			return err
		}
	}
	return nil
}

// IsPerformed indicates whether the group is performed
func (p *performerGroup) IsPerformed() bool {
	return p.performed
}
