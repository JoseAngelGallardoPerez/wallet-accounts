package database

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"

	"github.com/Confialink/wallet-accounts/internal/config"
)

// CreateConnection creates a new db connection
func NewConnection(config *config.Config) (*gorm.DB, error) {
	db := config.Db
	// initialize a new db connection
	connection, err := gorm.Open(
		config.Db.Driver,
		fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/%s?parseTime=true", // username:password@protocol(host)/dbname?param=value
			db.User, db.Password, db.Host, db.Port, db.Schema,
		),
	)
	if err != nil {
		return nil, err
	}

	if db.IsDebugMode {
		connection.LogMode(true)
	}
	return connection, nil
}
