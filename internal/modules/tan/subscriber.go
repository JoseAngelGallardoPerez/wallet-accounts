package tan

import (
	"github.com/jinzhu/gorm"

	"github.com/Confialink/wallet-accounts/internal/modules/tan/model"
)

type SubscriberRepository struct {
	db *gorm.DB
}

func NewSubscriberRepository(db *gorm.DB) *SubscriberRepository {
	return &SubscriberRepository{db}
}

func (s *SubscriberRepository) AddSubscriber(uid string) {
	subscriber := &model.TanSubscriber{UID: uid}
	if s.db.Find(subscriber, "uid = ?", uid).RecordNotFound() {
		s.db.Create(subscriber)
	}
}

func (s *SubscriberRepository) UserIdsHavingZeroTans() ([]string, error) {
	subscribers := make([]*model.TanSubscriber, 0, 16)
	err := s.db.
		Model(&model.Tan{}).
		Where("uid NOT IN (SELECT uid FROM tans GROUP BY uid)").
		Limit(100).
		Scan(&subscribers).
		Error

	if nil != err {
		return nil, err
	}

	result := make([]string, 0, len(subscribers))
	for _, subscriber := range subscribers {
		result = append(result, subscriber.UID)
	}

	return result, nil
}
