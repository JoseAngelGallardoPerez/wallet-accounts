package service

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/Confialink/wallet-accounts/internal/errcodes"

	"github.com/Confialink/wallet-pkg-list_params"
	csvPkg "github.com/Confialink/wallet-pkg-utils/csv"
	"github.com/Confialink/wallet-pkg-utils/timefmt"
	"github.com/shopspring/decimal"

	"github.com/Confialink/wallet-accounts/internal/modules/account/form"
	"github.com/Confialink/wallet-accounts/internal/modules/account/model"
	"github.com/Confialink/wallet-accounts/internal/modules/account/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/syssettings"
)

const (
	statusActive   = "Active"
	statusInactive = "Inactive"
)

// Csv
type Csv struct {
	repository *repository.AccountRepository
}

func NewCsv(repository *repository.AccountRepository) *Csv {
	return &Csv{
		repository,
	}
}

func (s *Csv) GetAccountsFile(params *list_params.ListParams) (*csvPkg.File, error) {
	accounts, err := s.repository.GetList(params)
	if err != nil {
		return nil, err
	}
	currentTime := time.Now()
	timeSettings, err := syssettings.GetTimeSettings()
	if err != nil {
		return nil, err
	}

	file := csvPkg.NewFile()
	formattedCurrentTime := timefmt.FormatFilenameWithTime(currentTime, timeSettings.Timezone)
	file.Name = fmt.Sprintf("accounts-%s.csv", formattedCurrentTime)

	header := []string{"Creation Date", "Owner", "Account #", "Type", "Currency", "CurrentBalance", "Status"}
	file.WriteRow(header)

	for _, v := range accounts {
		var status string

		if v.IsActive != nil && *v.IsActive {
			status = statusActive
		} else {
			status = statusInactive
		}

		formattedCreatedAt := timefmt.Format(v.CreatedAt, timeSettings.DateTimeFormat, timeSettings.Timezone)
		//TODO: some accounts has userID which doesn't exists in users' table. We need to clean DB from old records.
		//TODO: links DOM-74
		userName := ""
		if v.User != nil {
			userName = *v.User.Username
		}

		record := []string{
			formattedCreatedAt,
			userName,
			v.Number,
			v.Type.Name,
			v.Type.CurrencyCode,
			v.Balance.String(),
			status,
		}
		file.WriteRow(record)
	}

	return file, nil
}

// CsvToAccounts imports accounts from csv
// File format: user id, number (string), type id, balance,
// maturity date(string, format: YYYY-MM-DD), interest account id, payout day
// status (1, active, Active to set status active, everything else will be parsed as inactive),
// allow withdrawals (1, true, 0, false), allow deposits (1, true, 0, false)
func (s *Csv) CsvToAccounts(b *bytes.Buffer) ([]*model.Account, error) {
	reader := csv.NewReader(bufio.NewReader(b))

	var accounts []*model.Account
	var c uint64
	headers := []string{"User ID", "Number", "Type ID", "CurrentBalance", "Maturity date",
		"Interest account id", "Payout day", "Status", "Allow withdrawals", "Allow deposits"}

	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, errcodes.CreatePublicError(errcodes.CodeFileInvalid, "File is not valid")
		}

		// start count lines
		c++

		if len(line) != 7 {
			return nil, errcodes.CreatePublicError(errcodes.CodeFileInvalid,
				fmt.Sprintf("Row number %d of the CSV file failed. Reason: Invalid number of columns. Expected 7, got %d", c, len(line)))
		}

		if s.shouldHeadersBeSkipped(headers, line) {
			// skip headers
			// set count of line to 0
			c = 0
			continue
		}

		user := line[0]
		number := line[1]
		accountType, _ := strconv.ParseUint(line[2], 10, 64)
		balance, _ := decimal.NewFromString(line[3])
		maturityDate, _ := time.Parse("2006-01-02", line[4])
		interestAccountId, _ := strconv.ParseUint(line[5], 10, 64)
		payoutDay, _ := strconv.ParseUint(line[6], 10, 64)

		var isActive bool
		if line[7] == "active" || line[7] == "Active" || line[7] == "1" {
			isActive = true
		}

		var allowWithdrawals bool
		if line[8] == "true" || line[8] == "1" {
			allowWithdrawals = true
		}

		var allowDeposit bool
		if line[9] == "true" || line[9] == "1" {
			allowDeposit = true
		}

		public := model.AccountPublic{
			UserId:           user,
			Number:           number,
			TypeID:           accountType,
			InitialBalance:   &balance,
			MaturityDate:     &maturityDate,
			IsActive:         &isActive,
			AllowWithdrawals: &allowWithdrawals,
			AllowDeposits:    &allowDeposit,
		}

		if interestAccountId != 0 {
			public.InterestAccountId = &interestAccountId
		}

		if payoutDay != 0 {
			public.PayoutDay = &payoutDay
		}

		account := model.Account{AccountPublic: public}
		accounts = append(accounts, &account)
	}

	if c == 0 {
		return nil, errcodes.CreatePublicError(errcodes.CodeFileEmpty, "File is empty")
	}

	return accounts, nil
}

// Import imports accounts from csv
// File format: user id, number (string), type id, balance,
// maturity date(string, format: YYYY-MM-DD), interest account id, payout day
// status (1, active, Active to set status active, everything else will be parsed as inactive),
// allow withdrawals (1, true, 0, false), allow deposits (1, true, 0, false)
func (s *Csv) CsvToAccountForms(b *bytes.Buffer) ([]*form.Account, error) {
	reader := csv.NewReader(bufio.NewReader(b))

	var forms []*form.Account
	var c uint64

	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, errcodes.CreatePublicError(errcodes.CodeFileInvalid, "File is not valid")
		}

		c++

		if len(line) != 10 {
			return nil, errcodes.CreatePublicError(errcodes.CodeFileInvalid,
				fmt.Sprintf("Row number %d of the CSV file failed. Reason: Invalid number of columns. Expected 10, got %d", c, len(line)))
		}

		user := line[0]
		number := line[1]
		accountType, _ := strconv.ParseUint(line[2], 10, 64)
		balance, _ := decimal.NewFromString(line[3])
		maturityDate, _ := time.Parse("2006-01-02", line[4])
		interestAccountId, _ := strconv.ParseUint(line[5], 10, 64)
		payoutDay, _ := strconv.ParseUint(line[6], 10, 64)

		var isActive bool
		if line[7] == "active" || line[7] == "Active" || line[7] == "1" {
			isActive = true
		}

		var allowWithdrawals bool
		if line[8] == "true" || line[8] == "1" {
			allowWithdrawals = true
		}

		var allowDeposit bool
		if line[9] == "true" || line[9] == "1" {
			allowDeposit = true
		}

		f := form.Account{
			UserId:           user,
			Number:           number,
			TypeId:           accountType,
			InitialBalance:   balance,
			MaturityDate:     &maturityDate,
			IsActive:         &isActive,
			AllowDeposits:    &allowDeposit,
			AllowWithdrawals: &allowWithdrawals,
		}

		if interestAccountId != 0 {
			f.InterestAccountId = &interestAccountId
		}

		if payoutDay != 0 {
			f.PayoutDay = &payoutDay
		}

		forms = append(forms, &f)
	}

	if c == 0 {
		return nil, errcodes.CreatePublicError(errcodes.CodeFileEmpty, "File is empty")
	}

	return forms, nil
}

func (s *Csv) shouldHeadersBeSkipped(headers []string, line []string) bool {
	eqCount := 0
	for i, header := range headers {
		if line[i] == header {
			eqCount++
		}
	}
	return eqCount == len(headers)
}
