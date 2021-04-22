package scheduled_transaction

import (
	"fmt"
	"time"

	"github.com/jinzhu/now"
)

func findNextDate(period Period, day int) time.Time {
	if day <= 0 || day > 31 {
		panic("invalid day is provided")
	}
	month := findMonth(period)
	year := findYear(month)

	layout := "2006-1-2"
	date, _ := time.Parse(layout, fmt.Sprintf("%d-%d-1", year, month))
	lastDayOfTheMonth := now.New(date).EndOfMonth()
	if day > lastDayOfTheMonth.Day() {
		day = lastDayOfTheMonth.Day()
	}

	result, _ := time.Parse(layout, fmt.Sprintf("%d-%d-%d", year, month, day))
	return result
}

func findYear(scheduledMonth time.Month) int {
	currentTime := time.Now()
	currentMonth := currentTime.Month()

	if currentMonth < scheduledMonth {
		return currentTime.Year()
	}
	return currentTime.Year() + 1
}

func findMonth(period Period) time.Month {
	currentMonth := time.Now().Month()
	switch period {
	case PeriodMonthly:
		if currentMonth == time.December {
			return time.January
		}
		return time.Month(currentMonth + 1)

	case PeriodQuarterly:
		for _, qmonth := range [3]time.Month{time.April, time.July, time.October} {
			if currentMonth < qmonth {
				return qmonth
			}
		}
		return time.January

	case PeriodBiAnnually:
		if currentMonth < time.June {
			return time.June
		}
		return time.January

	case PeriodAnnually:
		return time.January
	}

	panic("unknown period " + period)
}
