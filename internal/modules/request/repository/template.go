package repository

import (
	"github.com/Confialink/wallet-accounts/internal/errcodes"
	"github.com/Confialink/wallet-accounts/internal/modules/request/constants"
	"github.com/Confialink/wallet-accounts/internal/modules/request/model"
	"github.com/jinzhu/gorm"
)

type Template struct {
	db *gorm.DB
}

func NewTemplate(db *gorm.DB) *Template {
	return &Template{db: db}
}

func (t *Template) Create(template *model.Template) error {

	duplicate := &model.Template{}
	err := t.
		db.
		Model(duplicate).
		Find(duplicate, "user_id = ? and request_subject = ? and name = ?", template.UserId, template.RequestSubject, template.Name).
		Error

	if err == nil {
		return errcodes.CreatePublicError(errcodes.CodeDuplicateTransferTemplate, "template with the same name and request subject is already exist")
	} else if err != gorm.ErrRecordNotFound {
		return err
	}

	return t.db.Model(template).Create(template).Error
}

func (t *Template) FindById(id uint64) (*model.Template, error) {
	result := &model.Template{}
	err := t.db.Model(result).Find(result, id).Error
	return result, err
}

func (t *Template) Delete(template *model.Template) error {
	return t.db.Delete(template).Error
}

func (t *Template) FindByUserId(userId string) (model.Templates, error) {
	var result model.Templates
	return result, t.db.Model(result).Find(&result, "user_id = ?", userId).Error
}

func (t *Template) FindByUserIdAndRequestSubject(userId string, subject constants.Subject) (model.Templates, error) {
	var result model.Templates
	return result, t.db.Model(result).Find(&result, "user_id = ? and request_subject = ?", userId, subject).Error
}

func (t Template) WrapContext(db *gorm.DB) *Template {
	t.db = db
	return &t
}
