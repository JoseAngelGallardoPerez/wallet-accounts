package model

type User struct {
	UID       *string `json:"uid"`
	Email     *string `json:"email"`
	Username  *string `json:"username"`
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
	RoleName  *string `json:"roleName"`
	GroupId   *uint64 `json:"groupId"`
}

type FullUser struct {
	UID             string
	Email           string
	Username        string
	FirstName       string
	LastName        string
	PhoneNumber     string
	IsCorporate     bool
	RoleName        string
	Status          string
	PhysicalAddress *PhysicalAddress
	Group           *Group
	Company         *Company
}

type PhysicalAddress struct {
	PaZipPostalCode   string
	PaAddress         string
	PaAddress2NdLine  string
	PaCity            string
	PaCountryIso2     string
	PaStateProvRegion string
}

type Group struct {
	Name string
}

type Company struct {
	CompanyName string
}
