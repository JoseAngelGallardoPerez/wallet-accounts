package handler

import (
	"github.com/Confialink/wallet-pkg-list_params"
	userpb "github.com/Confialink/wallet-users/rpc/proto/users"

	"github.com/Confialink/wallet-accounts/internal/modules/moneyrequest/model"
	"github.com/Confialink/wallet-accounts/internal/modules/user/service"
)

type Params struct {
	userService *service.UserService
}

func NewParams(userService *service.UserService) *Params {
	return &Params{userService}
}

func (p *Params) forIncoming(query string) *list_params.ListParams {
	params := list_params.NewListParamsFromQuery(query, model.MoneyRequest{})
	params.AllowPagination()
	params.AllowSortings([]string{"status", "isNew", "createdAt"})
	params.AllowFilters([]string{"status", "isNew"})
	params.AddCustomFilter("isNew", list_params.BoolFilter("money_requests.is_new"))
	params.AllowIncludes([]string{"recipient"})
	params.AddCustomIncludes("recipient", p.RecipientsIncludes)

	return params
}

func (p *Params) forOutgoing(query string) *list_params.ListParams {
	params := list_params.NewListParamsFromQuery(query, model.MoneyRequest{})
	params.AllowPagination()
	params.AllowSortings([]string{"status", "createdAt"})
	params.AllowFilters([]string{"status"})
	params.AllowIncludes([]string{"sender"})
	params.AddCustomIncludes("sender", p.SendersIncludes)

	return params
}

func (p *Params) RecipientsIncludes(records []interface{}) error {
	requests := make([]*model.MoneyRequest, len(records))
	uids := make([]string, 0, len(records))
	for i, v := range records {
		r := v.(*model.MoneyRequest)
		requests[i] = r
		uids = append(uids, r.InitiatorUserID)
	}

	users, err := p.userService.GetByUIDs(uids)
	if err != nil {
		return err
	}

	for _, v := range requests {
		user := p.findUserByUID(users, v.InitiatorUserID)
		if user != nil {
			v.Recipient = p.fillUser(user)
		}
	}

	return nil
}

func (p *Params) SendersIncludes(records []interface{}) error {
	requests := make([]*model.MoneyRequest, len(records))
	uids := make([]string, 0, len(records))
	for i, v := range records {
		r := v.(*model.MoneyRequest)
		requests[i] = r
		uids = append(uids, r.TargetUserID)
	}

	users, err := p.userService.GetByUIDs(uids)
	if err != nil {
		return err
	}

	for _, v := range requests {
		user := p.findUserByUID(users, v.TargetUserID)
		if user != nil {
			v.Sender = p.fillUser(user)
		}
	}

	return nil
}

func (p *Params) findUserByUID(array []*userpb.User, uid string) *userpb.User {
	for _, v := range array {
		if v.UID == uid {
			return v
		}
	}
	return nil
}

func (p *Params) fillUser(user *userpb.User) *model.User {
	return &model.User{
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		PhoneNumber: user.PhoneNumber,
	}
}
