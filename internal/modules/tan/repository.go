package tan

import (
	"github.com/Confialink/wallet-accounts/internal/modules/tan/model"
	"github.com/jinzhu/gorm"
)

type Repository struct {
	db *gorm.DB
}

type TansCountResult struct {
	Uid   string
	Count uint
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db}
}

func (r *Repository) SaveInTransaction(tans []*model.Tan) error {
	var err error
	tx := r.db.Begin()
	for _, tan := range tans {
		err = tx.Save(tan).Error
		if nil != err {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

func (r *Repository) Create(uid, tan string) (*model.Tan, error) {
	tanModel := &model.Tan{Tan: tan, UID: uid}
	return tanModel, r.db.Save(tanModel).Error
}

func (r *Repository) Delete(tan *model.Tan) error {
	return r.db.Delete(tan).Error
}

func (r *Repository) DeleteByUserId(userId string) error {
	return r.db.Delete(&model.Tan{}, "uid = ?", userId).Error
}

func (r *Repository) FindByUID(uid string) ([]*model.Tan, error) {
	result := make([]*model.Tan, 0, 20)
	return result, r.db.Find(&result, "uid = ?", uid).Error
}

func (r *Repository) CountByUserId(userId string) (uint, error) {
	var result uint = 0
	return result, r.db.Model(&model.Tan{}).Where("uid = ?", userId).Count(&result).Error
}

func (r *Repository) FindUserIdsHavingTansLessThan(qty uint) ([]*TansCountResult, error) {
	results := make([]*TansCountResult, 0, 16)
	return results, r.db.
		Model(&model.Tan{}).
		Select("uid, count(0) as count").
		Group("uid").Having("count < ?", qty).
		Scan(&results).
		Error
}
