package import_csv

import (
	"github.com/Confialink/wallet-accounts/internal/modules/request/model"
	"github.com/Confialink/wallet-pkg-errors"
	"github.com/shopspring/decimal"
)

type Context struct {
	errors  []errors.TypedError
	request *model.Request
	rate    *decimal.Decimal
	status  *string
}

func NewContext() *Context {
	return &Context{}
}

func (c *Context) AddError(err errors.TypedError) {
	c.errors = append(c.errors, err)
}

func (c *Context) GetErrors() []errors.TypedError {
	return c.errors
}

func (c *Context) GetRequest() *model.Request {
	return c.request
}

func (c *Context) SetRequest(r *model.Request) {
	c.request = r
}

func (c *Context) GetRate() *decimal.Decimal {
	return c.rate
}

func (c *Context) SetRate(r decimal.Decimal) {
	c.rate = &r
}

func (c *Context) GetStatus() *string {
	return c.status
}

func (c *Context) SetStatus(s string) {
	c.status = &s
}
