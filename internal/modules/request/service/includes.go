package service

import (
	"fmt"

	"github.com/Confialink/wallet-accounts/internal/modules/balance"
	"github.com/Confialink/wallet-accounts/internal/modules/request/model"
	"github.com/Confialink/wallet-pkg-utils/value"
)

type Includes struct {
	balanceResolver balance.Resolver
}

func NewIncludes(
	balanceResolver balance.Resolver,
) *Includes {
	return &Includes{balanceResolver: balanceResolver}
}

func (i *Includes) BalanceDifference(records []interface{}) error {
	requests := i.toRequestSlice(records)
	for _, request := range requests {
		request.BalanceDifference = make([]*balance.Difference, 0, len(request.Transactions))
		diffMap := map[string]*balance.Difference{}
		for _, transaction := range request.Transactions {
			if transaction.RevenueAccountId != nil {
				continue
			}

			resolvedBalance, err := i.balanceResolver.Resolve(transaction)
			if err != nil {
				return err
			}

			diffKey := fmt.Sprintf("%s_%d", resolvedBalance.TypeName(), value.FromUint64(resolvedBalance.GetId()))
			diff, set := diffMap[diffKey]
			if !set {
				currencyCode, err := resolvedBalance.GetCurrencyCode()
				if err != nil {
					return err
				}
				diff = &balance.Difference{
					CurrencyCode: currencyCode,
					BalanceType:  resolvedBalance.TypeName(),
					BalanceId:    resolvedBalance.GetId(),
				}
				diffMap[diffKey] = diff
				request.BalanceDifference = append(request.BalanceDifference, diff)
			}

			diff.Difference = diff.Difference.Add(*transaction.Amount)
		}
	}
	return nil
}

func (i *Includes) toRequestSlice(records []interface{}) []*model.Request {
	requests := make([]*model.Request, len(records))
	for i, request := range records {
		requests[i] = request.(*model.Request)
	}
	return requests
}
