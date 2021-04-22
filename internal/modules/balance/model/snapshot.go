package model

import (
	"encoding/json"
	"time"

	"github.com/Confialink/wallet-accounts/internal/modules/balance"
	"github.com/shopspring/decimal"
)

type SnapshotValue struct {
	AvailableAmount decimal.Decimal `json:"availableAmount"`
	Balance         decimal.Decimal `json:"balance"`
}

type Snapshot struct {
	Id            uint64     `json:"id"`
	RequestId     uint64     `json:"requestId"`
	BalanceTypeId uint64     `json:"balanceTypeId"`
	BalanceType   *Type      `json:"balanceType" gorm:"foreignkey:BalanceTypeId"`
	UserId        *string    `json:"userId"`
	BalanceId     *uint64    `json:"balanceId"`
	Snapshot      string     `json:"snapshot"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     *time.Time `json:"updatedAt"`
}

func (s *Snapshot) SetValue(balance balance.Balance) error {
	currentBalance, err := balance.CurrentBalance()
	if err != nil {
		return err
	}

	availableAmount, err := balance.AvailableBalance()
	if err != nil {
		return err
	}
	serialized, err := json.Marshal(SnapshotValue{
		Balance:         currentBalance,
		AvailableAmount: availableAmount,
	})
	if err != nil {
		return err
	}

	s.Snapshot = string(serialized)
	return nil
}

func (s *Snapshot) GetValue() (*SnapshotValue, error) {
	value := &SnapshotValue{}
	if s.Snapshot == "" {
		return value, nil
	}

	err := json.Unmarshal([]byte(s.Snapshot), value)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (s *Snapshot) MarshalJSON() ([]byte, error) {
	val, err := s.GetValue()
	if err != nil {
		return nil, err
	}
	data := map[string]interface{}{
		"id":            s.Id,
		"snapshotValue": val,
	}
	if s.BalanceType != nil {
		data["type"] = s.BalanceType
	}

	return json.Marshal(data)
}

func (s *Snapshot) TableName() string {
	return "balance_snapshots"
}
