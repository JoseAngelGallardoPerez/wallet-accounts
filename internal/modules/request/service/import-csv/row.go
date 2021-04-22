package import_csv

import "strings"

type Row struct {
	AccountNumber string
	Action        string
	Amount        string
	Description   string
	Revenue       string
	ApplyIwtFee   string

	Len    int
	Number uint64
}

func NewRow(line []string, n uint64) Row {
	if len(line) != 6 {
		return Row{
			Len:    len(line),
			Number: n,
		}
	}

	return Row{
		AccountNumber: line[0],
		Action:        strings.ToLower(line[1]),
		Amount:        line[2],
		Description:   line[3],
		Revenue:       strings.ToLower(line[4]),
		ApplyIwtFee:   strings.ToLower(line[5]),
		Len:           len(line),
		Number:        n,
	}
}
