package service

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/Confialink/wallet-accounts/internal/errcodes"

	"github.com/Confialink/wallet-pkg-list_params"
	csvPkg "github.com/Confialink/wallet-pkg-utils/csv"
	"github.com/Confialink/wallet-pkg-utils/pointer"
	"github.com/Confialink/wallet-pkg-utils/timefmt"
	"github.com/shopspring/decimal"

	"github.com/Confialink/wallet-accounts/internal/modules/card/model"
	"github.com/Confialink/wallet-accounts/internal/modules/syssettings"
)

type Csv struct {
	service *CardService
}

func NewCsv(service *CardService) *Csv {
	return &Csv{service}
}

// Import imports accounts from csv
// File format: user id, number, card type id, expiration month, expiration year, status
func (s Csv) CsvToCards(b *bytes.Buffer) ([]*model.Card, error) {
	reader := csv.NewReader(bufio.NewReader(b))

	var cards []*model.Card

	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, errcodes.CreatePublicError(errcodes.CodeFileInvalid, "File is not valid")
		}

		if len(line) != 6 {
			return nil, errcodes.CreatePublicError(errcodes.CodeFileInvalid, "File is not valid")
		}

		user := line[0]
		number := line[1]
		cardTypeId, _ := strconv.ParseUint(line[2], 10, 32)
		expirationMonth, _ := strconv.ParseUint(line[3], 10, 64)
		expirationYear, _ := strconv.ParseUint(line[4], 10, 64)
		status := line[5]

		card := model.Card{
			UserId:          &user,
			Number:          &number,
			CardTypeId:      pointer.ToUint32(uint32(cardTypeId)),
			ExpirationMonth: pointer.ToInt32(int32(expirationMonth)),
			ExpirationYear:  pointer.ToInt32(int32(expirationYear)),
			Status:          &status,
			Balance:         &decimal.Zero,
		}
		cards = append(cards, &card)
	}

	return cards, nil
}

func (s *Csv) GetCardsFile(params *list_params.ListParams) (*csvPkg.File, error) {
	cards, err := s.service.GetList(params)
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
	file.Name = fmt.Sprintf("cards-%s.csv", formattedCurrentTime)

	header := []string{"Card id", "Card Creation Date", "Card Owner", "Card #", "Card Type", "Card type category", "Currency", "Expiration Date", "Status"}
	file.WriteRow(header)

	for _, v := range cards {
		formattedCreatedAt := timefmt.Format(*v.CreatedAt, timeSettings.DateTimeFormat, timeSettings.Timezone)

		record := []string{
			strconv.FormatUint(uint64(*v.Id), 10),
			formattedCreatedAt,
			*v.User.Username,
			s.service.GetNumber(v),
			*v.CardType.Name,
			*v.CardType.Category.Name,
			*v.CardType.CurrencyCode,
			s.getExpirationDate(v),
			strings.Title(*v.Status),
		}
		file.WriteRow(record)
	}

	return file, nil
}

func (s *Csv) getExpirationDate(card *model.Card) string {
	var month string
	if card.ExpirationMonth == nil {
		month = "not set"
	} else {
		monthObj := time.Month(*card.ExpirationMonth)
		month = monthObj.String()
	}
	return fmt.Sprintf("%s / %d", month, *card.ExpirationYear)
}
