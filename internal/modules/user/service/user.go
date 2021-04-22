package service

import (
	"github.com/Confialink/wallet-accounts/internal/srvdiscovery"
	"context"
	"net/http"

	"github.com/Confialink/wallet-accounts/internal/modules/user/model"
	pb "github.com/Confialink/wallet-users/rpc/proto/users"
)

type UserService struct {
}

func NewUserService() *UserService {
	return &UserService{}
}

// GetByUsername returns User by passed username
func (u *UserService) GetByUsername(username string) (*pb.User, error) {
	req := pb.Request{Username: username}
	client, err := u.getClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.GetByUsername(context.Background(), &req)
	if err != nil {
		return nil, err
	}
	return resp.User, nil
}

// find users by first_name, last_name, email, uid, username
func (u *UserService) GetByProfileData(searchQuery string) ([]*pb.User, error) {
	req := pb.Request{
		Username:      searchQuery,
		SearchColumns: []string{"first_name", "last_name"},
	}
	client, err := u.getClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.GetByProfileData(context.Background(), &req)
	if err != nil {
		return nil, err
	}
	return resp.Users, nil
}

func (u *UserService) GetByUID(uid string) (*pb.User, error) {
	req := pb.Request{UID: uid}
	client, err := u.getClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.GetByUID(context.Background(), &req)
	if err != nil {
		return nil, err
	}
	return resp.User, nil
}

func (u *UserService) GetByUIDs(uids []string) ([]*pb.User, error) {
	req := pb.Request{UIDs: uids}
	client, err := u.getClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.GetByUIDs(context.Background(), &req)
	if err != nil {
		return nil, err
	}
	return resp.Users, nil
}

func (u *UserService) GetFullByUIDs(uids, fields []string) ([]*model.FullUser, error) {
	req := pb.RequestFullUsersByUIDs{UIDs: uids, Fields: fields}
	client, err := u.getClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.GetFullUsersByUIDs(context.Background(), &req)
	if err != nil {
		return nil, err
	}

	result := make([]*model.FullUser, len(resp.FullUsers))
	for i, v := range resp.FullUsers {
		result[i] = pbFullUserToFullUser(v)
	}

	return result, nil
}

func (u *UserService) GetUserStaff(currentUser *pb.User) ([]*pb.User, error) {
	parentUID := currentUser.ParentUID
	if parentUID == "" {
		parentUID = currentUser.UID
	}
	req := pb.Request{ParentUID: parentUID}
	client, err := u.getClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.GetStaffUsers(context.Background(), &req)
	if err != nil {
		return nil, err
	}
	return resp.Users, nil
}

func (u *UserService) GetAndSaveCompaniesByNames(names []string) ([]*pb.Company, error) {
	req := pb.CompaniesNameRequest{Names: names}
	client, err := u.getClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.SaveCompaniesByName(context.Background(), &req)
	if err != nil {
		return nil, err
	}
	return resp.Companies, nil
}

func (u *UserService) GetCompaniesByIDs(ids []uint64) ([]*pb.Company, error) {
	req := pb.CompaniesIDsRequest{IDs: ids}
	client, err := u.getClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.GetCompaniesByIDs(context.Background(), &req)
	if err != nil {
		return nil, err
	}
	return resp.Companies, nil
}

func (u *UserService) getClient() (pb.UserHandler, error) {
	usersUrl, err := srvdiscovery.ResolveRPC(srvdiscovery.ServiceNameUsers)
	if nil != err {
		return nil, err
	}

	return pb.NewUserHandlerProtobufClient(usersUrl.String(), http.DefaultClient), nil
}

func pbFullUserToFullUser(pbUser *pb.FullUser) *model.FullUser {
	return &model.FullUser{
		UID:             pbUser.Uid,
		Email:           pbUser.Email,
		Username:        pbUser.Username,
		FirstName:       pbUser.FirstName,
		LastName:        pbUser.LastName,
		PhoneNumber:     pbUser.PhoneNumber,
		IsCorporate:     pbUser.IsCorporate,
		RoleName:        pbUser.RoleName,
		Status:          pbUser.Status,
		PhysicalAddress: protoPhysicalAddressToPhysicalAddress(pbUser.PhysicalAdress),
		Group:           protoGroupToGroup(pbUser.UserGroup),
		Company:         protoCompanyToCompany(pbUser.CompanyDetails),
	}
}

func protoGroupToGroup(group *pb.UserGroup) *model.Group {
	if group == nil {
		return nil
	}

	return &model.Group{
		Name: group.Name,
	}
}

func protoPhysicalAddressToPhysicalAddress(address *pb.PhysicalAdress) *model.PhysicalAddress {
	if address == nil {
		return nil
	}

	return &model.PhysicalAddress{
		PaZipPostalCode:   address.PaZipPostalCode,
		PaAddress:         address.PaAddress,
		PaAddress2NdLine:  address.PaAddress_2NdLine,
		PaCity:            address.PaCity,
		PaCountryIso2:     address.PaCountryIso2,
		PaStateProvRegion: address.PaStateProvRegion,
	}
}

func protoCompanyToCompany(company *pb.Company) *model.Company {
	if company == nil {
		return nil
	}

	return &model.Company{
		CompanyName: company.CompanyName,
	}
}
