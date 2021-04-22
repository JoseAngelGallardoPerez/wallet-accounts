package balance

import (
	accountsRepository "github.com/Confialink/wallet-accounts/internal/modules/account/repository"
	cardsRepository "github.com/Confialink/wallet-accounts/internal/modules/card/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/transaction/model"
	list_params "github.com/Confialink/wallet-pkg-list_params"
	"github.com/pkg/errors"

	"github.com/jinzhu/gorm"
)

const (
	NotResolved       = Error("balance is not resolved")
	AlreadyRegistered = Error("resolver is already registered")
)

type Resolver interface {
	Resolve(source interface{}) (Balance, error)
}

type ResolverWithDbContext interface {
	Resolver
	WrapContext(db *gorm.DB) ResolverWithDbContext
}

type ResolverCallable func(source interface{}) (Balance, error)

func (r ResolverCallable) Resolve(source interface{}) (Balance, error) {
	return r(source)
}

type Resolvers map[string]Resolver

func (r Resolvers) Resolve(source interface{}) (Balance, error) {
	for _, resolver := range r {
		balance, err := resolver.Resolve(source)
		if err != nil && errors.Cause(err) != NotResolved {
			return nil, err
		}
		if balance != nil {
			return balance, nil
		}
	}
	return nil, NotResolved
}

func (r Resolvers) WrapContext(db *gorm.DB) ResolverWithDbContext {
	clone := Resolvers{}
	for name, resolver := range r {
		if resolver, ok := resolver.(ResolverWithDbContext); ok {
			clone[name] = resolver.WrapContext(db)
			continue
		}
		clone[name] = resolver
	}
	return clone
}

func (r Resolvers) RegisterResolver(name string, resolver ResolverCallable) error {
	if _, registered := r[name]; registered {
		return AlreadyRegistered
	}
	r[name] = resolver
	return nil
}

type defaultResolver struct {
	accountsRepository        *accountsRepository.AccountRepository
	cardsRepository           cardsRepository.CardRepositoryInterface
	revenueAccountsRepository *accountsRepository.RevenueAccountRepository
}

func NewDefaultResolver(
	accountsRepository *accountsRepository.AccountRepository,
	cardsRepository cardsRepository.CardRepositoryInterface,
	revenueAccountsRepository *accountsRepository.RevenueAccountRepository,
) Resolver {
	return &defaultResolver{
		accountsRepository:        accountsRepository,
		cardsRepository:           cardsRepository,
		revenueAccountsRepository: revenueAccountsRepository,
	}
}

func (t *defaultResolver) Resolve(source interface{}) (Balance, error) {
	var tx *model.Transaction
	switch source := source.(type) {
	case Balance:
		return source, nil
	case *model.Transaction:
		tx = source
	case model.Transaction:
		tx = &source
	default:
		return nil, errors.Wrapf(
			NotResolved,
			"defaultResolver: expected given source to be type of *model.Transaction or model.Transaction but got %T",
			source,
		)
	}

	if tx.AccountId != nil {
		return t.accountsRepository.FindByID(*tx.AccountId)
	}
	if tx.CardId != nil {
		return t.cardsRepository.Get(*tx.CardId, list_params.NewIncludes("include=CardType"))
	}
	if tx.RevenueAccountId != nil {
		return t.revenueAccountsRepository.FindByID(*tx.RevenueAccountId)
	}
	return nil, errors.Wrapf(
		NotResolved,
		"defaultResolver: given transaction(#%d) fields AccountId, CardId and RevenueAccountId are all nil",
		*tx.Id,
	)
}

func (t defaultResolver) WrapContext(db *gorm.DB) ResolverWithDbContext {
	t.accountsRepository = t.accountsRepository.WrapContext(db)
	t.cardsRepository = t.cardsRepository.WrapContext(db)
	t.revenueAccountsRepository = t.revenueAccountsRepository.WrapContext(db)
	return &t
}
