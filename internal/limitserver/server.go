package limitserver

import (
	"github.com/Confialink/wallet-accounts/internal/limit"
	"github.com/Confialink/wallet-accounts/internal/modules/request/transfers"
	rpcLimit "github.com/Confialink/wallet-accounts/rpc/limit"
	"context"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

var limitNameEnumToStrMap = map[rpcLimit.LimitName]string{
	rpcLimit.LimitName_EMPTY:                     "",
	rpcLimit.LimitName_MAX_TOTAL_BALANCE:         transfers.LimitMaxTotalBalance,
	rpcLimit.LimitName_MAX_CREDIT_PER_TRANSFER:   transfers.LimitMaxCreditPerTransfer,
	rpcLimit.LimitName_MAX_DEBIT_PER_TRANSFER:    transfers.LimitMaxDebitPerTransfer,
	rpcLimit.LimitName_MAX_TOTAL_DEBIT_PER_DAY:   transfers.LimitMaxTotalDebitPerDay,
	rpcLimit.LimitName_MAX_TOTAL_DEBIT_PER_MONTH: transfers.LimitMaxTotalDebitPerMonth,
}

var limitNameStrToEnumMap = map[string]rpcLimit.LimitName{
	"":                                   rpcLimit.LimitName_EMPTY,
	transfers.LimitMaxTotalBalance:       rpcLimit.LimitName_MAX_TOTAL_BALANCE,
	transfers.LimitMaxCreditPerTransfer:  rpcLimit.LimitName_MAX_CREDIT_PER_TRANSFER,
	transfers.LimitMaxDebitPerTransfer:   rpcLimit.LimitName_MAX_DEBIT_PER_TRANSFER,
	transfers.LimitMaxTotalDebitPerDay:   rpcLimit.LimitName_MAX_TOTAL_DEBIT_PER_DAY,
	transfers.LimitMaxTotalDebitPerMonth: rpcLimit.LimitName_MAX_TOTAL_DEBIT_PER_MONTH,
}

type Server struct {
	limitService *limit.Service
	db           *gorm.DB
}

// NewServer is limit RPC server constructor
func NewServer(limitService *limit.Service, db *gorm.DB) *Server {
	return &Server{limitService: limitService, db: db}
}

// Set creates new limits and updates existing ones
func (s *Server) Set(_ context.Context, request *rpcLimit.SetLimitsRequest) (*rpcLimit.SetLimitsResponse, error) {
	tx := s.db.Begin()
	srv := s.limitService.WrapContext(tx)

	for _, data := range request.Limits {
		val, err := requestLimitToValue(data.Limit)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		id := requestIdToLimitId(data.LimitId)
		if !id.IsUnique() {
			tx.Rollback()
			return nil, limit.ErrIdIncomplete
		}
		_, err = srv.FindOne(id)
		// create new if not found
		if errors.Cause(err) == limit.ErrNotFound {
			err = srv.Create(val, id)
			if err != nil {
				tx.Rollback()
				return nil, err
			}
			continue
		}
		err = srv.UpdateOne(val, id)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	tx.Commit()
	return &rpcLimit.SetLimitsResponse{}, nil
}

// Get retrieves a list of limits
func (s *Server) Get(_ context.Context, request *rpcLimit.GetLimitsRequest) (*rpcLimit.GetLimitsResponse, error) {
	result := make([]*rpcLimit.LimitWithId, 0, len(request.Identifiers))
	srv := s.limitService

	for _, reqId := range request.Identifiers {
		found, err := srv.Find(requestIdToLimitId(reqId))

		if errors.Cause(err) == limit.ErrNotFound || len(found) == 0 {
			result = append(result, &rpcLimit.LimitWithId{
				Limit:   &rpcLimit.Limit{Exists: false},
				LimitId: reqId,
			})
			continue
		}
		if err != nil {
			return nil, err
		}
		for _, l := range found {
			result = append(result, &rpcLimit.LimitWithId{
				Limit:   limitValueToRequestLimit(l.Available()),
				LimitId: limitIdToRequestId(l.Identifier()),
			})
		}
	}
	return &rpcLimit.GetLimitsResponse{
		Limits: result,
	}, nil
}

// ResetToDefault deletes specified limits so that the default values are applied
func (s *Server) ResetToDefault(_ context.Context, request *rpcLimit.ResetLimitsRequest) (*rpcLimit.ResetLimitsResponse, error) {
	tx := s.db.Begin()
	srv := s.limitService.WrapContext(tx)

	for _, reqId := range request.Identifiers {
		err := srv.DeleteOne(requestIdToLimitId(reqId))
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	tx.Commit()
	return &rpcLimit.ResetLimitsResponse{}, nil
}

func requestIdToLimitId(id *rpcLimit.LimitId) limit.Identifier {
	return limit.Identifier{
		Name:     limitNameEnumToStrMap[id.GetName()],
		Entity:   id.GetEntity(),
		EntityId: id.GetEntityId(),
	}
}

func limitIdToRequestId(limitId limit.Identifier) *rpcLimit.LimitId {
	return &rpcLimit.LimitId{
		Name:     limitNameStrToEnumMap[limitId.Name],
		Entity:   limitId.Entity,
		EntityId: limitId.EntityId,
	}
}

func requestLimitToValue(lim *rpcLimit.Limit) (limit.Value, error) {
	if lim.NoLimit {
		return limit.NoLimit(), nil
	}
	amount, err := decimal.NewFromString(lim.Amount)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to convert received limit amount value '%s' to decimal",
			lim.Amount,
		)
	}
	return limit.Val(amount, lim.CurrencyCode), nil
}

func limitValueToRequestLimit(val limit.Value) *rpcLimit.Limit {
	result := &rpcLimit.Limit{
		Exists: true,
	}
	if val.NoLimit() {
		result.NoLimit = true
		return result
	}
	result.Amount = val.CurrencyAmount().Amount().String()
	result.CurrencyCode = val.CurrencyAmount().CurrencyCode()

	return result
}
