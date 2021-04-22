package service

import (
	"testing"

	cardModel "github.com/Confialink/wallet-accounts/internal/modules/card/model"
	userpb "github.com/Confialink/wallet-users/rpc/proto/users"
	"github.com/stretchr/testify/assert"
)

func TestcanClientReadCardCorrectUserId(t *testing.T) {
	var userId string = "1"
	user := userpb.User{UID: userId}
	card := cardModel.Card{UserId: &userId}

	assert.True(t, canClientReadCard(&card, &user))
}

func TestcanClientReadCardInvalidUserId(t *testing.T) {
	var userId string = "1"
	var userIdReference string = "2"
	user := userpb.User{UID: userId}
	card := cardModel.Card{UserId: &userIdReference}

	assert.False(t, canClientReadCard(&card, &user))
}
