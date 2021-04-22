package service

import (
	"github.com/Confialink/wallet-accounts/internal/modules/card/model"
	userpb "github.com/Confialink/wallet-users/rpc/proto/users"
)

// canClientReadCard checks if client can read a card
func canClientReadCard(card interface{}, user *userpb.User) bool {
	cardPtr := card.(*model.Card)
	return *(cardPtr.UserId) == user.UID
}
