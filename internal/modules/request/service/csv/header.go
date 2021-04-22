package csv

import "github.com/pkg/errors"

func GetHeader(roleName string) ([]string, error) {
	switch roleName {
	case "root", "admin":
		return getAdminHeader(), nil
	default:
		return nil, errors.New("undefined role name " + roleName)
	}
}

func getAdminHeader() []string {
	return []string{
		"Request ID",
		"Request Date",
		"Request Updated",
		"User name",
		"Profile type",
		"Company name",
		"First Name",
		"Last Name",
		//"Country",
		//"Zip code",
		//"State",
		//"City",
		//"Address",
		//"Address 2nd line",
		"Group",
		"Pay from Account number",
		"Account currency",
		"Description",
		"Subject",
		"Status",
		"OWT fee type",
		"Payment amount",
		"Payment currency",
		"Beneficiary Bank: SWIFT / BIC",
		"Beneficiary Bank: Name",
		"Beneficiary Bank: Address",
		"Beneficiary Bank: Location",
		"Beneficiary Bank: Country",
		"Beneficiary Bank: ABA/RTN",
		"Beneficiary: Name",
		"Beneficiary: Address",
		"Beneficiary: Acc#/IBAN",
		"Ref message",
		"Intermediary Bank SWIFT / BIC",
		"Intermediary Bank Name",
		"Intermediary Bank Address",
		"Intermediary Bank Location",
		"Intermediary Bank Country",
		"Intermediary Bank ABA/RTN",
		"Intermediary Bank Acc#/IBAN",
	}
}
