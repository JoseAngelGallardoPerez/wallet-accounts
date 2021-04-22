package service

import (
	"strconv"

	"github.com/Confialink/wallet-pkg-errors"
	userpb "github.com/Confialink/wallet-users/rpc/proto/users"
	"github.com/gin-gonic/gin"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	accountModel "github.com/Confialink/wallet-accounts/internal/modules/account/model"
	cardModel "github.com/Confialink/wallet-accounts/internal/modules/card/model"
	requestModel "github.com/Confialink/wallet-accounts/internal/modules/request/model"
	transactionModel "github.com/Confialink/wallet-accounts/internal/modules/transaction/model"
)

type ContextInterface interface {
	GetCurrentUser(c *gin.Context) *userpb.User
	MustGetCurrentUser(c *gin.Context) *userpb.User
	GetIdParam(c *gin.Context) (uint64, errors.TypedError)
	GetRequestedAccount(c *gin.Context) *accountModel.Account
	GetRequestedCard(c *gin.Context) *cardModel.Card
	GetRequestedRequest(c *gin.Context) *requestModel.Request
	GetRequestedTransaction(c *gin.Context) *transactionModel.Transaction
}

type Context struct{}

func NewContext() ContextInterface {
	context := Context{}
	return &context
}

// getCurrentUser returns current user or nil
func (s *Context) GetCurrentUser(c *gin.Context) *userpb.User {
	account, exist := c.Get("_user")
	if !exist {
		return nil
	}
	return account.(*userpb.User)
}

// mustGetCurrentUser returns current user or throw error
func (s *Context) MustGetCurrentUser(c *gin.Context) *userpb.User {
	user := s.GetCurrentUser(c)
	if nil == user {
		panic("user must be set")
	}
	return user
}

// GetIdParam returns id or nil
func (s *Context) GetIdParam(c *gin.Context) (uint64, errors.TypedError) {
	id := c.Params.ByName("id")

	// convert string to uint
	id64, err := strconv.ParseUint(id, 10, 64)

	if err != nil {
		return 0, errcodes.CreatePublicError(errcodes.CodeNumeric, "id param must be an integer")
	}

	return uint64(id64), nil
}

// GetRequestedAccount returns requested account or nil
func (s *Context) GetRequestedAccount(c *gin.Context) *accountModel.Account {
	account, exist := c.Get("_requested_account")
	if !exist {
		return nil
	}
	return account.(*accountModel.Account)
}

// GetRequestedCard returns requested card or nil
func (s *Context) GetRequestedCard(c *gin.Context) *cardModel.Card {
	account, exist := c.Get("_requested_card")
	if !exist {
		return nil
	}
	return account.(*cardModel.Card)
}

// GetRequestedRequest returns requested request or nil
func (s *Context) GetRequestedRequest(c *gin.Context) *requestModel.Request {
	request, exist := c.Get("_requested_request")
	if !exist {
		return nil
	}
	return request.(*requestModel.Request)
}

// GetRequestedTransaction returns requested transaction or nil
func (s *Context) GetRequestedTransaction(c *gin.Context) *transactionModel.Transaction {
	transaction, exist := c.Get("_requested_transaction")
	if !exist {
		return nil
	}
	return transaction.(*transactionModel.Transaction)
}
