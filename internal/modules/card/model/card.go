package model

import (
	"github.com/pkg/errors"
	"time"

	"github.com/shopspring/decimal"

	cardTypeModel "github.com/Confialink/wallet-accounts/internal/modules/card-type/model"
)

type Card struct {
	Id              *uint32                 `json:"id"`
	Number          *string                 `json:"number" binding:"required,validCardFormat,cardNumberUnique"`
	Balance         *decimal.Decimal        `json:"balance" binding:"required"`
	Status          *string                 `json:"status" binding:"required"`
	CardTypeId      *uint32                 `json:"cardTypeId" binding:"required"`
	UserId          *string                 `json:"userId" binding:"required,existUserId,userIsActive"`
	ExpirationYear  *int32                  `json:"expirationYear" binding:"required"`
	ExpirationMonth *int32                  `json:"expirationMonth" binding:"required,gte=1,lte=12"`
	CreatedAt       *time.Time              `json:"createdAt"`
	UpdatedAt       *time.Time              `json:"updatedAt"`
	CardType        *cardTypeModel.CardType `gorm:"foreignkey:CardTypeId;association_foreignkey:Id;association_autoupdate:false" json:"cardType"`
	User            *User                   `json:"user"`
}

type SerializedCard struct {
	Id              *uint32                           `json:"id"`
	Number          *string                           `json:"number"`
	Status          *string                           `json:"status"`
	CardTypeId      *uint32                           `json:"cardTypeId"`
	UserId          *string                           `json:"userId"`
	ExpirationYear  *int32                            `json:"expirationYear"`
	ExpirationMonth *int32                            `json:"expirationMonth"`
	CreatedAt       *string                           `json:"createdAt"`
	CardType        *cardTypeModel.SerializedCardType `json:"cardType"`
	User            *User                             `json:"user"`
	Balance         *decimal.Decimal                  `json:"balance"`
}

type User struct {
	Id       *string `json:"id"`
	Username *string `json:"username"`
	Email     *string `json:"email"`
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
}

// TableName sets Card's table name to be `cards`
func (*SerializedCard) TableName() string {
	return "cards"
}

func (c *Card) CurrentBalance() (decimal.Decimal, error) {
	return *c.Balance, nil
}

func (c *Card) AvailableBalance() (decimal.Decimal, error) {
	return *c.Balance, nil
}

func (c *Card) GetCurrencyCode() (string, error) {
	if c.CardType == nil {
		return "", errors.New("card type must be loaded to be able to access currency code")
	}
	return *c.CardType.CurrencyCode, nil
}

func (c *Card) TypeName() string {
	return "card"
}

func (c *Card) GetId() *uint64 {
	id := uint64(*c.Id)
	return &id
}

func (c *Card) GetUserId() *string {
	return c.UserId
}
