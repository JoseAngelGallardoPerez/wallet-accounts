package model

import (
	"time"

	cardTypeCategoryModel "github.com/Confialink/wallet-accounts/internal/modules/card-type-category/model"
	cardTypeFormatModel "github.com/Confialink/wallet-accounts/internal/modules/card-type-format/model"
)

type CardType struct {
	Id                 *uint32                                 `json:"id"`
	Name               *string                                 `json:"name" binding:"required,min=1,max=250"`
	CurrencyCode       *string                                 `json:"currencyCode" binding:"required,activeCurrencyCode"`
	IconId             *uint64                                 `json:"iconId"`
	CreatedAt          *time.Time                              `json:"createdAt"`
	UpdatedAt          *time.Time                              `json:"updatedAt"`
	CardTypeCategoryId *uint32                                 `json:"cardTypeCategoryId" binding:"required"`
	CardTypeFormatId   *uint32                                 `json:"cardTypeFormatId" binding:"required"`
	Category           *cardTypeCategoryModel.CardTypeCategory `json:"category" gorm:"foreignkey:CardTypeCategoryId"`
	Format             *cardTypeFormatModel.CardTypeFormat     `json:"format" gorm:"foreignkey:CardTypeFormatId"`
}

type SerializedCardType struct {
	Id                 *uint32                                 `json:"id"`
	Name               *string                                 `json:"name"`
	CurrencyCode       *string                                 `json:"currencyCode"`
	IconId             *uint64                                 `json:"iconId"`
	CardTypeCategoryId *uint32                                 `json:"cardTypeCategoryId"`
	CardTypeFormatId   *uint32                                 `json:"cardTypeFormatId"`
	Category           *cardTypeCategoryModel.CardTypeCategory `json:"category"`
	Format             *cardTypeFormatModel.CardTypeFormat     `json:"format"`
}
