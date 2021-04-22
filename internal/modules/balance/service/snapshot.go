package service

import (
	"github.com/Confialink/wallet-accounts/internal/modules/balance"
	balanceModel "github.com/Confialink/wallet-accounts/internal/modules/balance/model"
	"github.com/Confialink/wallet-accounts/internal/modules/balance/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/request/model"
	"github.com/Confialink/wallet-pkg-utils/value"
	"github.com/inconshreveable/log15"
	"github.com/jinzhu/gorm"
)

type Snapshot struct {
	typeRepository     *repository.Type
	snapshotRepository *repository.Snapshot
	logger             log15.Logger
}

func NewSnapshot(
	typeRepository *repository.Type,
	snapshotRepository *repository.Snapshot,
	logger log15.Logger,
) *Snapshot {
	return &Snapshot{
		typeRepository:     typeRepository,
		snapshotRepository: snapshotRepository,
		logger:             logger.New("service", "BalanceSnapshot"),
	}
}

func (s *Snapshot) MakeSnapshot(request *model.Request, balance balance.Balance) (*balanceModel.Snapshot, error) {
	logger := s.logger.New("method", "NewSnapshot")
	typeModel, err := s.typeRepository.FindByName(balance.TypeName())
	if err != nil {
		logger.Error("failed to find balance type by name", "error", err)
		return nil, err
	}

	snapshot, err := s.snapshotRepository.FindOrInitByRequestIdAndBalanceId(value.FromUint64(request.Id), value.FromUint64(balance.GetId()))
	if err != nil {
		logger.Error("failed to find or init balance snapshot", "error", err)
		return nil, err
	}
	if snapshot.Id == 0 {
		snapshot.RequestId = *request.Id
		snapshot.BalanceTypeId = typeModel.Id
		snapshot.BalanceId = balance.GetId()
		snapshot.UserId = balance.GetUserId()
	}

	err = snapshot.SetValue(balance)
	if err != nil {
		logger.Error("failed to set snapshot value", "error", err)
		return nil, err
	}

	err = s.snapshotRepository.Save(snapshot)
	if err != nil {
		logger.Error("failed to save balance snapshot", "error", err)
		return nil, err
	}

	return snapshot, nil
}

func (s Snapshot) WrapContext(db *gorm.DB) *Snapshot {
	s.snapshotRepository = s.snapshotRepository.WrapContext(db)
	s.typeRepository = s.typeRepository.WrapContext(db)
	return &s
}
