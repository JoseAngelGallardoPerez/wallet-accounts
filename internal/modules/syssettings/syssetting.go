package syssettings

import (
	"context"
	"fmt"

	pb "github.com/Confialink/wallet-settings/rpc/proto/settings"

	"github.com/Confialink/wallet-accounts/internal/modules/syssettings/connection"
)

// TimeSettings struct has timezone and date format
type TimeSettings struct {
	Timezone       string
	DateFormat     string
	TimeFormat     string
	DateTimeFormat string
}

// GetTimeSettings returns new TimeSettings from settings service or err if can not get it
func GetTimeSettings() (*TimeSettings, error) {
	timeSettings := TimeSettings{}
	client, err := connection.GetSystemSettingsClient()
	if err != nil {
		return &timeSettings, err
	}

	response, err := client.List(context.Background(), &pb.Request{Path: "regional/general/%"})
	if err != nil {
		return &timeSettings, err
	}

	timeSettings.Timezone = getSettingValue(response.Settings, "regional/general/default_timezone")
	timeSettings.DateFormat = getSettingValue(response.Settings, "regional/general/default_date_format")
	timeSettings.TimeFormat = getSettingValue(response.Settings, "regional/general/default_time_format")
	timeSettings.DateTimeFormat = fmt.Sprintf("%s %s", timeSettings.DateFormat, timeSettings.TimeFormat)

	return &timeSettings, nil
}

func getSettingValue(settings []*pb.Setting, path string) string {
	for _, v := range settings {
		if v.Path == path {
			return v.Value
		}
	}
	return ""
}
