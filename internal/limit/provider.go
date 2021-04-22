package limit

func Providers() []interface{} {
	return []interface{}{
		NewFactory,
		NewStorageGORM,
		NewService,
	}
}
