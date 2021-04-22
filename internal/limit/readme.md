[TOC]

# Limit Package

The package provides functionality to work with limits.

## What is a limit?

A limit is a certain amount that should not be violated in one way or another. 
This package specializes in currency values(because of the accounts service specific), 
but can be easily adapted to work with any other values.

## Usage examples

**Check if limit exceeded**
```go
	maxLimit := limit.Max(decimal.NewFromInt(100), "BTC")
	myAmount := limit.Amount(decimal.NewFromInt(101), "BTC")

	err := maxLimit.WithinLimit(myAmount)
	// check if limit is exceeded (101 BTC is greater than 100 BTC)
	fmt.Println(errors.Cause(err) == limit.ErrLimitExceeded) // true
	fmt.Println(err) // the requested value 101 exceeds the available limit 100: LIMIT_EXCEEDED

	myAmount = limit.Amount(decimal.NewFromInt(100), "BTC")

	err = maxLimit.WithinLimit(myAmount)
	// limit is not exceeded and no other errors occurred
	fmt.Println(err == nil) // true
```

**No limit**
```go
	unlimited := limit.New(limit.NoLimit())
	anyAmount := limit.Amount(decimal.NewFromInt(9999999999), "ANY")

	fmt.Println(unlimited.WithinLimit(anyAmount)) // <nil>
	// It is possible to check whether the limit is "unlimited"
	fmt.Println(unlimited.Available().NoLimit()) // true
	// unlimited values have no currency amount
	fmt.Println(unlimited.Available().CurrencyAmount()) // <nil>
```

## Storing limits

The package provides convenient way to define, store and retrieve limits. 
In order to uniquely define a limit you should specify 3 properties:

* **Name** of the limit
* **Entity** with which this limit is associated
* **EntityId** which identifies the entity

These properties are defined by the "Identifier" struct.
```go
// Identifier represents properties that uniquely defines limit
type Identifier struct {
	// Name is limit name that have some semantic meaning e.g. max_per_transfer, max_per_day
	Name string
	// Entity is the name of the entity with which this limit is associated e.g. user, account etc.
	Entity string
	// EntityId is unique identifier of the entity (primary key)
	EntityId string
}
```

The second thing we need is limit "Value". 
Basically the "Value" defines limit parameters i.e. "no limit" or "limited by an amount". 
```go
// Value defines limit value
type Value interface {
	NoLimit() bool
	CurrencyAmount() CurrencyAmount
}
```
There are 2 functions that help to create limit values:
* limit.Val(amount decimal.Decimal, currencyCode string) Value
* limit.NoLimit() Value

The below examples demonstrates limit.Storage usage.

Save / Update
```go
	// 100 USD
	val := limit.Val(decimal.NewFromInt(100), "USD")
	// NOTE that this is not "real" limit, this just illustrates how the package could be used
	id := limit.Identifier{
		Name:     "max_balance",
		Entity:   "account",
		EntityId: "123-456-789",
	}
	// create "max_balance" limit of 100 USD for the account "123-456-789"
	err := storage.Save(val, id)
	if err != nil {
		// ...
	}
	
	// update the limit	
	err = storage.Update(limit.NoLimit(), id)
	if err != nil {
		// ...
	}
```

Take a look at the limit.Storage interface:
```go
// Storage defines how to store and fetch limits
type Storage interface {
	// Save saves new limit with the given identifier
	Save(value Value, identifier Identifier) error
	// Update updates available amount on existing limit by its identifier
	Update(value Value, identifier Identifier) error
	// Find retrieves all limits that much the given identifier parameters
	// must return ErrNotFound if no records is found
	Find(identifier Identifier) ([]Model, error)
	// Delete deletes all limits that much the given identifier parameters
	Delete(identifier Identifier) error
}
```

There is storage implementation in the package which uses "GORM" as storage backend.

* NewStorageGORM(db *gorm.DB) Storage

## Service

Although storage interface is simple, it is not very convenient to use it for searching as it works with models.
Models only cary out the data, but do not implement an interface for checking the limit.

The service requires "storage" and "factory". 
limit.Factory is responsible for creating instances of limit.Limit using a model data.
```go
// Factory defines minimum requirements for creating limits
type Factory interface {
	// CreateLimit creates Limit by the given model
	CreateLimit(model Model) Limit
}
```

It could be useful when you need to define your own limit implementation.
There is a default factory which creates a "max" limits, and an "unlimited" limits.

```go
	service := limit.NewService(storage, limit.NewFactory())
	// Find all limits by name
	id := limit.Identifier{
		Name:     "some limit name",
	}	
	limits, err := service.Find(id)
	if err != nil {
		//...
	}
	// iterate over the found limits
	for _, lim := range limits {
		limId := lim.Identifier()
		available := lim.Available()
		fmt.Printf("Entity: %s, EntityId: %s", limId.Entity, limId.EntityId)
		if available.NoLimit() {
			fmt.Println("Available: unlimited")
		} else {
			amount := available.CurrencyAmount()
			fmt.Printf("Available: %s %s", amount.Amount(), amount.CurrencyCode())
		}
		// lim.WithinLimit(...)
	}
    // Could output something like this:
    //  Entity: user, EntityId: 123
    //      Available: 100 EUR
    //  Entity: user, EntityId: 456
    //      Available: 120 BTC
    //  Entity: user, EntityId: 789
    //      Available: unlimited
```

Note how `lim.Identifier()` returns limit identifier. 
This is because the service returns instance of "limit.IdentifiableLimit". 
It leverages "limit.NewIdentifiableLimit" which wraps any limit together with the given identifier.

* NewIdentifiableLimit(limit Limit, identifier Identifier) IdentifiableLimit 

```go
// Identifiable is anything that could return Identifier
type Identifiable interface {
	Identifier() Identifier
}

// IdentifiableLimit is Limit that provides its Identifier
type IdentifiableLimit interface {
	Limit
	Identifiable
}
```

In the above example, we used only the name of the limit for the search, i.e. searched by name only.
You can use any other identifier property to search for this property only.
By adding additional properties, you add "AND" search criteria and thereby narrow the selection result. 

In order to find unique records you can specify all 3 properties:
```go
	service := limit.NewService(storage, limit.NewFactory())
	// Find one limit
	id := limit.Identifier{
		Name:     "some limit name",
		Entity:   "user",
		EntityId: "123",
	}
	// method "FindOne" uses "Find" but always returns first element from the result slice 
	lim, err := service.FindOne(id)
	if err != nil {
		//...
	}
```

> For more details please check source code and test files.

> This package is used to implement transfer limitations, see [transfers.](../modules/request/transfers/readme.md)