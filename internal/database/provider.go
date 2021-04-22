package database

func Providers() []interface{} {
	return []interface{}{
		//*gorm.DB
		NewConnection,
	}
}
