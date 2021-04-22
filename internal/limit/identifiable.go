package limit

// Identifier represents properties that uniquely defines limit
type Identifier struct {
	// Name is limit name that have some semantic meaning e.g. max_per_transfer, max_per_day
	Name string
	// Entity is the name of the entity with which this limit is associated e.g. user, account etc.
	Entity string
	// EntityId is unique identifier of the entity (primary key)
	EntityId string
}

// Identifier returns Identifier
func (i Identifier) Identifier() Identifier {
	return i
}

// IsUnique indicates that all fields of the identifier contain non-empty values,
// this means that the identifier describes only 1 object, i.e. unique
func (i *Identifier) IsUnique() bool {
	return i.Name != "" && i.Entity != "" && i.EntityId != ""
}

// Identifiable is anything that could return Identifier
type Identifiable interface {
	Identifier() Identifier
}

// IdentifiableLimit is Limit that provides its Identifier
type IdentifiableLimit interface {
	Limit
	Identifiable
}

type identifiable struct {
	limit      Limit
	identifier Identifier
}

// NewIdentifiableLimit is used in order to attach identifier to the given limit
func NewIdentifiableLimit(limit Limit, identifier Identifier) IdentifiableLimit {
	return &identifiable{limit: limit, identifier: identifier}
}

// Identifier returns Identifier
func (i *identifiable) Identifier() Identifier {
	return i.identifier
}

// Value indicates how much can be spent
func (i *identifiable) Available() Value {
	return i.limit.Available()
}

// WithinLimit checks if the given value is within limit
func (i *identifiable) WithinLimit(value CurrencyAmount) error {
	return i.limit.WithinLimit(value)
}
