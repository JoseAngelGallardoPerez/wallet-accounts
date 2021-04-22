package update_csv

import "strings"

type Row struct {
	RequestId string
	Status    string
	Rate      string

	Len    int
	Number uint64
}

func NewRow(line []string, n uint64) Row {
	if len(line) != 3 && len(line) != 2 {
		return Row{
			Len:    len(line),
			Number: n,
		}
	}

	status := strings.ToLower(line[1])
	if status == StatusCanceled {
		status = StatusCancelled
	}

	rate := ""
	if len(line) == 3 {
		rate = line[2]
	}

	return Row{
		RequestId: line[0],
		Status:    status,
		Rate:      rate,
		Len:       len(line),
		Number:    n,
	}
}
