package service

import (
	"regexp"

	userpb "github.com/Confialink/wallet-users/rpc/proto/users"
)

var tanSettingRegEx = regexp.MustCompile(`\w{2,3}_tan_required`)

//  check if setting is part of tan setting
func canClientReadSetting(settingName interface{}, _ *userpb.User) bool {
	return tanSettingRegEx.MatchString(settingName.(string))
}
