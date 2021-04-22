package service

import (
	"context"

	"github.com/Confialink/wallet-accounts/internal/modules/currency/connection"
	"github.com/Confialink/wallet-accounts/internal/modules/currency/model"
	"github.com/Confialink/wallet-accounts/internal/modules/currency/serializer"
	pb "github.com/Confialink/wallet-currencies/rpc/currencies"
	"github.com/shopspring/decimal"
)

type CurrenciesServiceInterface interface {
	Convert(amount decimal.Decimal, currencyCodeFrom, currencyCodeTo string) (decimal.Decimal, error)
	GetByCode(string) (*model.Currency, error)
	GetCurrenciesRateValueByCodes(currencyCodeFrom, currencyCodeTo string) (decimal.Decimal, error)
	GetCurrenciesRateByCodes(currencyCodeFrom, currencyCodeTo string) (*Rate, error)
}

type currenciesService struct {
	connection connection.CurrencyConnectionInterface
	serializer serializer.CurrencySerializerInterface
}

func (s *currenciesService) Convert(amount decimal.Decimal, currencyCodeFrom, currencyCodeTo string) (decimal.Decimal, error) {
	rate, err := s.GetCurrenciesRateValueByCodes(currencyCodeFrom, currencyCodeTo)
	if err != nil {
		return decimal.Zero, err
	}
	return amount.Mul(rate), nil
}

func (s *currenciesService) GetCurrenciesRateValueByCodes(currencyCodeFrom, currencyCodeTo string) (decimal.Decimal, error) {
	request := pb.CurrenciesRateValueRequest{
		CurrencyCodeFrom: currencyCodeFrom,
		CurrencyCodeTo:   currencyCodeTo,
	}
	client, err := s.connection.GetClient()
	if err != nil {
		return decimal.Zero, err
	}
	if response, err := client.GetCurrenciesRateValueByCodes(context.Background(), &request); err != nil {
		return decimal.Zero, err
	} else {
		return decimal.NewFromString(response.Value)
	}
}

func (s *currenciesService) GetCurrenciesRateByCodes(currencyCodeFrom, currencyCodeTo string) (*Rate, error) {
	request := pb.CurrenciesRateValueRequest{CurrencyCodeFrom: currencyCodeFrom, CurrencyCodeTo: currencyCodeTo}
	client, err := s.connection.GetClient()
	if err != nil {
		return nil, err
	}

	response, err := client.GetCurrenciesRateByCodes(context.Background(), &request)
	if err != nil {
		return nil, err
	}

	decimalRate, err := decimal.NewFromString(response.Value)
	if err != nil {
		return nil, err
	}
	decimalMargin, err := decimal.NewFromString(response.ExchangeMargin)
	if err != nil {
		return nil, err
	}

	rate := &Rate{
		Rate:           decimalRate,
		ExchangeMargin: decimalMargin,
	}

	return rate, nil
}

func NewCurrenciesService(
	connection connection.CurrencyConnectionInterface,
	serializer serializer.CurrencySerializerInterface,
) CurrenciesServiceInterface {
	return &currenciesService{connection, serializer}
}

func (s *currenciesService) GetByCode(code string) (*model.Currency, error) {
	currencyRequest := pb.CurrencyReq{Code: code}
	client, err := s.connection.GetClient()
	if err != nil {
		return nil, err
	}
	if currencyResponse, err := client.GetCurrency(context.Background(), &currencyRequest); err != nil {
		return nil, err
	} else {
		return s.serializer.Deserialize(currencyResponse), nil
	}
}

func (s *currenciesService) GetMain() (*model.Currency, error) {
	client, err := s.connection.GetClient()
	if err != nil {
		return nil, err
	}
	if currencyResponse, err := client.GetMain(context.Background(), &pb.CurrencyReq{}); err != nil {
		return nil, err
	} else {
		return s.serializer.Deserialize(currencyResponse), nil
	}
}
