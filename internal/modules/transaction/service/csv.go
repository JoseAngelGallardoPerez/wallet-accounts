package service

import (
	"fmt"
	"strconv"
	"time"

	requestModel "github.com/Confialink/wallet-accounts/internal/modules/request/model"
	requestRepository "github.com/Confialink/wallet-accounts/internal/modules/request/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/syssettings"
	"github.com/Confialink/wallet-accounts/internal/modules/transaction/model"
	"github.com/Confialink/wallet-pkg-list_params"
	"github.com/Confialink/wallet-pkg-utils/csv"
	"github.com/Confialink/wallet-pkg-utils/timefmt"
)

type Csv struct {
	requestsRepository requestRepository.RequestRepositoryInterface
}

func NewCsvService(requestsRepository requestRepository.RequestRepositoryInterface) *Csv {
	return &Csv{requestsRepository}
}

func (service *Csv) FileWithTransactions(transactions []*model.Transaction,
) (*csv.File, error) {
	currentTime := time.Now()
	timeSettings, err := syssettings.GetTimeSettings()
	if err != nil {
		return nil, err
	}

	file := csv.NewFile()
	formattedCurrentTime := timefmt.FormatFilenameWithTime(currentTime, timeSettings.Timezone)
	file.Name = fmt.Sprintf("transactions-%s.csv", formattedCurrentTime)

	file.WriteRow([]string{})
	if err := service.writeTransactions(transactions, timeSettings, file); err != nil {
		return nil, err
	}
	return file, nil
}

func (service *Csv) writeTransactions(transactions []*model.Transaction, timeSettings *syssettings.TimeSettings, file *csv.File) error {
	transactionsHeader := []string{"Date", "Transaction ID", "Description", "Debit/CreditFromAlias", "Value balance", "Status"}
	file.WriteRow(transactionsHeader)

	requestsById, err := service.requestsMap(transactions)
	if err != nil {
		return err
	}

	for _, v := range transactions {
		var balance string
		if v.AvailableBalanceSnapshot != nil {
			balance = v.AvailableBalanceSnapshot.String()
		}

		if v.ShowAvailableBalanceSnapshot != nil {
			balance = v.ShowAvailableBalanceSnapshot.String()
		}

		amount := *v.Amount
		if v.ShowAmount != nil {
			amount = *v.ShowAmount
		}

		request := requestsById[*v.RequestId]
		formattedDate := timefmt.Format(*request.StatusChangedAt, timeSettings.DateTimeFormat, timeSettings.Timezone)

		record := []string{
			formattedDate,
			strconv.FormatUint(*v.Id, 10),
			*v.Description,
			amount.String(),
			balance,
			*v.Status,
		}
		file.WriteRow(record)
	}

	return nil
}

// return map of request by id
func (service *Csv) requestsMap(transactions []*model.Transaction) (requestsById map[uint64]*requestModel.Request, err error) {
	requestIds := make([]string, 0)
	for _, t := range transactions {
		requestIds = append(requestIds, strconv.FormatUint(uint64(*t.RequestId), 10))
	}

	listParams := list_params.NewListParamsFromQuery("", requestModel.Request{})
	listParams.AddFilter("id", requestIds, list_params.OperatorIn)
	listParams.Pagination.PageSize = 0

	requests, err := service.requestsRepository.GetList(listParams)
	if err != nil {
		return requestsById, err
	}

	requestsById = make(map[uint64]*requestModel.Request)
	_ = requests
	for _, r := range requests {
		requestsById[*r.Id] = r
	}

	return requestsById, nil
}
