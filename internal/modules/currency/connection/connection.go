package connection

import (
	"github.com/Confialink/wallet-accounts/internal/srvdiscovery"
	"net/http"

	currencyPb "github.com/Confialink/wallet-currencies/rpc/currencies"
	pb "github.com/Confialink/wallet-currencies/rpc/currencies"
)

type CurrencyConnectionInterface interface {
	GetClient() (pb.CurrencyFetcher, error)
}

type currencyConnection struct {
}

func NewCurrencyConnection() CurrencyConnectionInterface {
	return &currencyConnection{}
}

func (c *currencyConnection) GetClient() (pb.CurrencyFetcher, error) {
	currenciesUrl, err := srvdiscovery.ResolveRPC(srvdiscovery.ServiceNameCurrencies)
	if err != nil {
		return nil, err
	}

	return currencyPb.NewCurrencyFetcherProtobufClient(currenciesUrl.String(), http.DefaultClient), nil
}
