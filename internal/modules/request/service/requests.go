package service

import (
	accountModel "github.com/Confialink/wallet-accounts/internal/modules/account/model"
	accountsRepository "github.com/Confialink/wallet-accounts/internal/modules/account/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/request/model"
	usersService "github.com/Confialink/wallet-accounts/internal/modules/user/service"
)

type RequestsService struct {
	usersService       *usersService.UserService
	accountsRepository *accountsRepository.AccountRepository
}

func NewRequestsService(
	usersService *usersService.UserService,
	accountsRepository *accountsRepository.AccountRepository,
) *RequestsService {
	return &RequestsService{
		usersService:       usersService,
		accountsRepository: accountsRepository,
	}
}

// LoadFullUsers assigns users by source account of request
// Works only for requests having source account
func (s *RequestsService) LoadFullUsers(requests []*model.Request, userFields []string) error {
	accountIds := make([]uint64, 0, 2)
	accountRequest := make(map[uint64][]*model.Request)
	for _, r := range requests {
		if accId, exists := r.SourceAccountId(); exists {
			accIdU64 := uint64(accId)
			if accountRequest[accIdU64] == nil {
				accountRequest[accIdU64] = make([]*model.Request, 0, 2)
				accountIds = append(accountIds, accIdU64)
			}
			accountRequest[accIdU64] = append(accountRequest[accIdU64], r)
		}
	}

	sourceAccounts, err := s.accountsRepository.FindManyByIds(accountIds)
	if err != nil {
		return err
	}
	userIds := make([]string, 0, len(sourceAccounts))
	userAccount := make(map[string][]*accountModel.Account)
	for _, acc := range sourceAccounts {
		if userAccount[acc.UserId] == nil {
			userAccount[acc.UserId] = make([]*accountModel.Account, 0, 2)
			userIds = append(userIds, acc.UserId)
		}
		userAccount[acc.UserId] = append(userAccount[acc.UserId], acc)
	}

	fullUsers, err := s.usersService.GetFullByUIDs(userIds, userFields)
	if err != nil {
		return err
	}

	for _, u := range fullUsers {
		if userAccounts, ok := userAccount[u.UID]; ok {
			for _, acc := range userAccounts {
				if requests, ok := accountRequest[acc.ID]; ok {
					for _, request := range requests {
						request.FullUser = u
					}
				}
			}
		}
	}
	return nil
}
