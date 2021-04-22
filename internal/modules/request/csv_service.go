package request

import (
	"github.com/Confialink/wallet-accounts/internal/modules/request/event"
	"github.com/Confialink/wallet-accounts/internal/modules/request/transfers"
	"github.com/Confialink/wallet-accounts/internal/modules/transaction/types"
	"github.com/Confialink/wallet-accounts/internal/transfer"
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"github.com/olebedev/emitter"
	"io"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	accRepo "github.com/Confialink/wallet-accounts/internal/modules/account/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/request/form"
	importCsv "github.com/Confialink/wallet-accounts/internal/modules/request/service/import-csv"
	updateCsv "github.com/Confialink/wallet-accounts/internal/modules/request/service/update-csv"
	"github.com/Confialink/wallet-pkg-errors"
	"github.com/Confialink/wallet-users/rpc/proto/users"

	"github.com/Confialink/wallet-pkg-list_params"
	csvPkg "github.com/Confialink/wallet-pkg-utils/csv"
	"github.com/Confialink/wallet-pkg-utils/timefmt"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"

	"github.com/Confialink/wallet-accounts/internal/modules/request/constants"
	"github.com/Confialink/wallet-accounts/internal/modules/request/model"
	"github.com/Confialink/wallet-accounts/internal/modules/request/repository"
	csvServices "github.com/Confialink/wallet-accounts/internal/modules/request/service/csv"
	"github.com/Confialink/wallet-accounts/internal/modules/syssettings"
)

const ActionCredit = "credit"
const ActionDebit = "debit"

type updateRequestContainer struct {
	request *model.Request
	status  *string
	rate    *decimal.Decimal
}

type CsvService struct {
	requestRepository repository.RequestRepositoryInterface
	accountRepository *accRepo.AccountRepository
	currencyProvider  transfer.CurrencyProvider
	creator           *Creator
	db                *gorm.DB
	logger            log15.Logger
	emitter           *emitter.Emitter
	pf                transfers.PermissionFactory
}

func NewCsvService(
	requestRepository repository.RequestRepositoryInterface,
	accountRepository *accRepo.AccountRepository,
	currencyProvider transfer.CurrencyProvider,
	creator *Creator,
	db *gorm.DB,
	logger log15.Logger,
	emitter *emitter.Emitter,
	pf transfers.PermissionFactory,
) *CsvService {
	return &CsvService{
		requestRepository: requestRepository,
		accountRepository: accountRepository,
		currencyProvider:  currencyProvider,
		creator:           creator,
		db:                db,
		logger:            logger.New("service", "RequestCsvService"),
		emitter:           emitter,
		pf:                pf,
	}
}

func (s *CsvService) GetCsvFile(params *list_params.ListParams, roleName string) (*csvPkg.File, error) {
	logger := s.logger.New("GetCsvFile")

	requests, err := s.requestRepository.GetList(params)
	if err != nil {
		logger.Error("Can't retrieve list of requests", "err", err)
		return nil, err
	}
	currentTime := time.Now()
	timeSettings, err := syssettings.GetTimeSettings()
	if err != nil {
		logger.Error("Can't get time settings", "err", err)
		return nil, err
	}

	file := csvPkg.NewFile()
	formattedCurrentTime := timefmt.FormatFilenameWithTime(currentTime, timeSettings.Timezone)
	file.Name = fmt.Sprintf("transfer-requests-%s.csv", formattedCurrentTime)

	header, err := csvServices.GetHeader(roleName)
	if err != nil {
		logger.Error("Can't build rows", "err", err)
		return nil, err
	}

	file.WriteRow(header)

	for _, request := range requests {
		rowBuilder := csvServices.NewRowBuilder(request, timeSettings, roleName)
		row, err := rowBuilder.Call()
		if err != nil {
			logger.Error("Can't build rows", "err", err)
			return file, err
		}
		file.WriteRow(row)
	}

	return file, nil
}

// UpdateFromCsv updates list of requests from csv
// File format: request id, status, rate
func (s *CsvService) UpdateFromCsv(b *bytes.Buffer, currentUser *users.User) (tErrs []errors.TypedError, failed uint64, success uint64) {
	reader := csv.NewReader(bufio.NewReader(b))

	var updContainers []updateRequestContainer
	var c uint64
	headers := []string{"Request ID", "Status", "Exchange rate"}

	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			tErrs = append(tErrs, errcodes.CreatePublicError(errcodes.CodeFileInvalid, "File is invalid"))
			continue
		}

		if s.shouldHeadersBeSkipped(headers, line) {
			// skip headers
			// set count of line to 1
			c = 1
			continue
		}

		// start count lines
		c++

		row := updateCsv.NewRow(line, c)

		v := updateCsv.NewLineNengthValidator()
		v.SetNext(updateCsv.NewRequestExistValidator(s.requestRepository)).
			SetNext(updateCsv.NewRequestStatusValidator()).
			SetNext(updateCsv.NewRequestRateValidator())

		ctx := updateCsv.NewContext()
		errs := v.Validate(row, ctx)

		// skip row if is not valid
		if len(errs) > 0 {
			tErrs = append(tErrs, errs...)
			failed++
			continue
		}

		updContainer := updateRequestContainer{
			request: ctx.GetRequest(),
			status:  ctx.GetStatus(),
			rate:    ctx.GetRate(),
		}

		updContainers = append(updContainers, updContainer)
	}

	if c == 0 {
		tErrs = append(tErrs, errcodes.CreatePublicError(errcodes.CodeFileEmpty, "File is empty"))
		return
	}

	requestTErrs, updFailed, updSuccess := s.updateRequests(updContainers, s.db, currentUser)
	if len(requestTErrs) > 0 {
		tErrs = append(tErrs, requestTErrs...)
		return
	}

	failed += updFailed
	success += updSuccess

	return
}

// ImportFromCsv imports list of requests from csv
// File format: account number, debit or credit, amount, description, revenue, apply IWT fee
func (s *CsvService) ImportFromCsv(b *bytes.Buffer,
	user *users.User) (tErrs []errors.TypedError, failed uint64, success uint64) {
	logger := s.logger.New("ImportFromCsv")
	reader := csv.NewReader(bufio.NewReader(b))
	var c uint64
	headers := []string{"Account number", "Debit or Credit", "amount", "Description", "Revenue", "Apply IWT Fee"}

	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			tErrs = append(tErrs, errcodes.CreatePublicError(errcodes.CodeFileInvalid, "File is invalid"))
			continue
		}

		if s.shouldHeadersBeSkipped(headers, line) {
			// skip headers
			// set count of line to 1
			c = 1
			continue
		}

		// start count lines
		c++
		row := importCsv.NewRow(line, c)

		v := importCsv.NewLineLengthValidator()
		v.SetNext(importCsv.NewRequestAccountNumberValidator(s.accountRepository)).
			SetNext(importCsv.NewRequestActionValidator()).
			SetNext(importCsv.NewRequestAmountValidator()).
			SetNext(importCsv.NewRequestDescriptionValidator()).
			SetNext(importCsv.NewRequestRevenueValidator()).
			SetNext(importCsv.NewRequestApplyIwtFeeValidator())

		ctx := importCsv.NewContext()
		errs := v.Validate(row, ctx)

		// skip row if is not valid
		if len(errs) > 0 {
			tErrs = append(tErrs, errs...)
			failed++
			continue
		}

		var revenue bool
		if row.Revenue == "yes" {
			revenue = true
		}

		var applyIwtFee bool
		if row.ApplyIwtFee == "yes" {
			applyIwtFee = true
		}

		tx := s.db.Begin()
		switch row.Action {
		case ActionCredit:
			err = s.createCARequest(row.AccountNumber, row.Description, row.Amount, revenue, applyIwtFee, user, tx)
		case ActionDebit:
			err = s.createDARequest(row.AccountNumber, row.Description, row.Amount, revenue, user, tx)
		}

		if nil != err {
			logger.Error("Can't create request", "err", err)

			message := "Can't create request"
			pubErr, ok := err.(*errors.PublicError)
			if ok {
				message = message + ": " + pubErr.Title
			}
			tErrs = append(tErrs, errcodes.CreatePublicError(errcodes.CodeCsvFileInvalidRow,
				fmt.Sprintf("Row number %d of the CSV file failed. Reason - %s", c, message)))
			failed++
			tx.Rollback()
			continue
		}

		success++
		tx.Commit()
	}

	if c == 0 {
		tErrs = append(tErrs, errcodes.CreatePublicError(errcodes.CodeFileEmpty, "file is empty"))
		return
	}

	if len(tErrs) > 0 {
		return
	}

	return
}

func (s CsvService) WrapContext(db *gorm.DB) *CsvService {
	s.db = db
	s.accountRepository = s.accountRepository.WrapContext(db)
	s.requestRepository = s.requestRepository.WrapContext(db)
	return &s
}

// update requests updates requests and return count of success, failed operations and errors
func (s *CsvService) updateRequests(
	updateContainers []updateRequestContainer,
	db *gorm.DB, currentUser *users.User,
) (tErrs []errors.TypedError, failed uint64, success uint64) {
	c := 0
	for _, container := range updateContainers {
		tx := db.Begin()
		c++
		request := container.request
		var (
			details types.Details
			topErr  error
		)

		if request.Rate != nil && container.rate != nil && !request.Rate.Equal(*container.rate) {
			request.Rate = container.rate
			modifier, err := transfers.CreateModifier(tx, request, s.currencyProvider, s.pf)
			if err != nil {
				tErr := errors.PrivateError{
					Message: "failed to create modifier",
				}

				tErr.AddLogPair("error", err.Error())
				tErrs = append(tErrs, &tErr)
				failed++
				tx.Rollback()
				continue
			}
			details, err = modifier.Modify(request)
			if err != nil {
				tErr := errors.PrivateError{
					Message: "failed to modify request rate",
				}

				tErr.AddLogPair("error", err.Error())
				tErrs = append(tErrs, &tErr)
				failed++
				tx.Rollback()
				continue
			}
			eventContext := &event.ContextRequestModified{
				Tx:      tx,
				Request: request,
				Details: details,
			}
			<-s.emitter.Emit(event.RequestModified, eventContext)
		}

		if request.Status != container.status {
			switch *container.status {
			case constants.StatusExecuted:
				executor, err := transfers.CreateExecutor(tx, request, s.currencyProvider, s.pf)
				if err != nil {
					tErr := errors.PrivateError{
						Message: "failed to create executor",
					}

					tErr.AddLogPair("error", err.Error())
					tErrs = append(tErrs, &tErr)
					failed++
					tx.Rollback()
					continue
				}
				details, topErr = executor.Execute(request)
				if topErr == nil {
					eventContext := &event.ContextRequestExecuted{
						Tx:      tx,
						Request: request,
						Details: details,
					}
					<-s.emitter.Emit(event.RequestExecuted, eventContext)
				}
			case constants.StatusCancelled:
				canceller, err := transfers.CreateCanceller(tx, request, s.currencyProvider, s.pf)
				if err != nil {
					tErr := errors.PrivateError{
						Message: "failed to create canceller",
					}

					tErr.AddLogPair("error", err.Error())
					tErrs = append(tErrs, &tErr)
					failed++
					tx.Rollback()
					continue
				}
				topErr = canceller.Cancel(request, "")
				if topErr == nil {
					eventContext := &event.ContextPendingRequestCancelled{
						Tx:        tx,
						UserID:    *request.UserId,
						RequestID: *request.Id,
						Reason:    "",
					}

					<-s.emitter.Emit(event.PendingRequestCancelled, eventContext)
				}
			default:
				tErrs = append(tErrs, errcodes.CreatePublicError(errcodes.CodeCsvFileInvalidRow,
					fmt.Sprintf("Row number %d of the CSV file failed. Reason - Status is not valid.", c)))
				failed++
				tx.Rollback()
				continue
			}
		}
		if topErr != nil {
			tErr := errors.PrivateError{
				Message: "failed to process action",
			}

			tErr.AddLogPair("error", topErr.Error())
			tErrs = append(tErrs, &tErr)
			failed++
			tx.Rollback()
			continue
		}

		tx.Commit()
		success++
	}

	return
}

func (s *CsvService) createCARequest(
	accountNumber string,
	description string,
	amount string,
	revenue bool,
	applyIwtFee bool,
	user *users.User,
	db *gorm.DB,
) error {
	acc, err := s.accountRepository.FindByNumber(accountNumber)
	if err != nil {
		return err
	}
	if acc == nil {
		return errcodes.CreatePublicError(errcodes.CodeAccountNotFound, "account "+accountNumber+" not found")
	}
	f := form.CA{
		AccountId:               acc.ID,
		Amount:                  amount,
		Description:             description,
		DebitFromRevenueAccount: &revenue,
		ApplyIwtFee:             &applyIwtFee,
	}
	_, err = s.creator.CreateCARequest(&f, user, db)
	if err != nil {
		return err
	}
	return nil
}

func (s *CsvService) createDARequest(
	accountNumber string,
	description string,
	amount string,
	revenue bool,
	user *users.User,
	db *gorm.DB,
) error {
	acc, err := s.accountRepository.FindByNumber(accountNumber)
	if err != nil {
		return err
	}
	if acc == nil {
		return errcodes.CreatePublicError(errcodes.CodeAccountNotFound, "account "+accountNumber+" not found")
	}
	f := form.DA{
		AccountId:              acc.ID,
		Amount:                 amount,
		Description:            description,
		CreditToRevenueAccount: &revenue,
	}
	_, err = s.creator.CreateDARequest(&f, user, db)
	if err != nil {
		return err
	}
	return nil
}

func (s *CsvService) shouldHeadersBeSkipped(headers []string, line []string) bool {
	if len(headers) != len(line) {
		return false
	}

	eqCount := 0
	for i, header := range headers {
		if line[i] == header {
			eqCount++
		}
	}
	return eqCount == len(headers)
}
