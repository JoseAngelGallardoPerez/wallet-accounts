package csv

import (
	"strconv"

	requestModel "github.com/Confialink/wallet-accounts/internal/modules/request/model"
	requestRepository "github.com/Confialink/wallet-accounts/internal/modules/request/repository"
	transactionModel "github.com/Confialink/wallet-accounts/internal/modules/transaction/model"
	transactionRepository "github.com/Confialink/wallet-accounts/internal/modules/transaction/repository"

	"github.com/Confialink/wallet-pkg-list_params"
)

type RequestsMapper struct {
	requestsRepository    requestRepository.RequestRepositoryInterface
	transactionRepository *transactionRepository.TransactionRepository
}

func NewRequestsMapper(
	requestsRepository requestRepository.RequestRepositoryInterface,
	transactionRepository *transactionRepository.TransactionRepository,
) *RequestsMapper {
	return &RequestsMapper{requestsRepository, transactionRepository}
}

// returns map of request by id
func (m *RequestsMapper) RequestsMap(transactions []*transactionModel.Transaction) (requestsMap map[uint64]interface{}, err error) {
	requestIds := m.getRequestIds(transactions)

	transactionsMap, err := m.transactionsMap(requestIds)
	if err != nil {
		return requestsMap, err
	}

	listParams := list_params.NewListParamsFromQuery("", requestModel.Request{})
	listParams.AddFilter("id", requestIds, list_params.OperatorIn)
	listParams.Pagination.PageSize = 0

	requests, err := m.requestsRepository.GetList(listParams)
	if err != nil {
		return requestsMap, err
	}

	requestsMap = make(map[uint64]interface{})
	_ = requests
	for _, r := range requests {
		transactions := transactionsMap[*r.Id]

		requestsMap[*r.Id] = map[string]interface{}{
			"request":      r,
			"transactions": transactions,
		}
	}

	return requestsMap, nil
}

// returns request ids
func (m *RequestsMapper) getRequestIds(transactions []*transactionModel.Transaction) []string {
	requestIds := make([]string, 0)
	for _, t := range transactions {
		requestIds = append(requestIds, strconv.FormatUint(uint64(*t.RequestId), 10))
	}
	return requestIds
}

// returns map of transactions by request id
func (m *RequestsMapper) transactionsMap(
	requestIds []string,
) (transactionsMap map[uint64][]*transactionModel.Transaction, err error) {
	listParams := list_params.NewListParamsFromQuery("", transactionModel.Transaction{})
	listParams.AddFilter("request_id", requestIds, list_params.OperatorIn)
	listParams.Pagination.PageSize = 0

	collection, err := m.transactionRepository.GetList(listParams)
	if err != nil {
		return transactionsMap, err
	}

	transactionsMap = make(map[uint64][]*transactionModel.Transaction)
	for _, t := range collection {
		transactionsMap[*t.RequestId] = append(transactionsMap[*t.RequestId], t)
	}

	return transactionsMap, nil
}
