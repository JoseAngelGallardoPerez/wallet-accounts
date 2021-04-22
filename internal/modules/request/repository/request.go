package repository

import (
	currenciesService "github.com/Confialink/wallet-accounts/internal/modules/currency/service"
	"github.com/Confialink/wallet-accounts/internal/modules/request/model"
	usersService "github.com/Confialink/wallet-accounts/internal/modules/user/service"
	"github.com/Confialink/wallet-pkg-list_params"
	"github.com/Confialink/wallet-pkg-list_params/adapters"
	pb "github.com/Confialink/wallet-users/rpc/proto/users"
	"github.com/jinzhu/gorm"
)

type RequestRepositoryInterface interface {
	Create(request *model.Request) (err error)
	Delete(request *model.Request) (err error)
	WrapContext(db *gorm.DB) RequestRepositoryInterface
	Updates(request *model.Request) error
	FindById(id uint64) (*model.Request, error)
	GetList(*list_params.ListParams) ([]*model.Request, error)
	GetListCount(*list_params.ListParams) (uint64, error)
	FillUsers(requests []*model.Request) error
}

type requestRepository struct {
	db                *gorm.DB
	currenciesService currenciesService.CurrenciesServiceInterface
	usersService      *usersService.UserService
	owtDataRepository *DataOwt
}

func NewRequestRepository(
	db *gorm.DB,
	currenciesService currenciesService.CurrenciesServiceInterface,
	usersService *usersService.UserService,
	owtDataRepository *DataOwt,
) RequestRepositoryInterface {
	return &requestRepository{
		db:                db,
		currenciesService: currenciesService,
		usersService:      usersService,
		owtDataRepository: owtDataRepository,
	}
}

// TODO: add user id check
func (r *requestRepository) Create(request *model.Request) error {
	return r.db.Create(request).Error
}

func (r *requestRepository) Delete(request *model.Request) error {
	return r.db.Delete(request).Error
}

func (r *requestRepository) FindById(id uint64) (*model.Request, error) {
	request := model.Request{}
	if err := r.db.Where("id = ?", id).First(&request).Error; err != nil {
		return nil, err
	}
	return &request, nil
}

func (r requestRepository) WrapContext(db *gorm.DB) RequestRepositoryInterface {
	r.db = db
	return &r
}

func (r *requestRepository) Updates(request *model.Request) error {
	return r.db.Model(request).Updates(request).Error
}

func (r *requestRepository) GetList(params *list_params.ListParams) (
	[]*model.Request, error,
) {
	var requests []*model.Request
	adapter := adapters.NewGorm(r.db)
	err := adapter.LoadList(&requests, params, "requests")

	return requests, err
}

func (r *requestRepository) GetListCount(params *list_params.ListParams) (
	uint64, error,
) {
	var count uint64
	str, arguments := params.GetWhereCondition()
	query := r.db.Where(str, arguments...)
	query = query.Joins(params.GetJoinCondition())

	if err := query.Model(&model.Request{}).Select("count(distinct(requests.id))").Count(&count).Error; err != nil {
		return count, err
	}

	return count, nil
}

func (r *requestRepository) FillUsers(requests []*model.Request) error {
	userIds := make([]string, 0)
	for _, v := range requests {
		if !r.isExistString(userIds, *v.UserId) {
			userIds = append(userIds, *v.UserId)
		}
	}

	users, err := r.usersService.GetByUIDs(userIds)
	if err != nil {
		return err
	}

	for _, v := range requests {
		user := r.findUserById(users, *v.UserId)
		if user != nil {
			r.fillUser(v, user)
		}
	}

	return nil
}

func (r *requestRepository) findUserById(array []*pb.User, id string) *pb.User {
	for _, v := range array {
		if v.UID == id {
			return v
		}
	}
	return nil
}

func (r *requestRepository) isExistString(array []string, elem string) bool {
	for _, v := range array {
		if v == elem {
			return true
		}
	}
	return false
}

func (r *requestRepository) fillUser(request *model.Request, user *pb.User) {
	request.User = &model.RequestUser{
		Id:        &user.UID,
		Username:  &user.Username,
		Email:     &user.Email,
		FirstName: &user.FirstName,
		LastName:  &user.LastName,
	}
}
