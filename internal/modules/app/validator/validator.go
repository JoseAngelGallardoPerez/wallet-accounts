package validator

import validatorPkg "github.com/go-playground/validator/v10"

// Interface is a validator interface
type Interface interface {
	Struct(current interface{}) error
	StructExcept(current interface{}, fields ...string) error
	RegisterValidation(key string, fn validatorPkg.Func, callValidationEvenIfNull ...bool) error
	RegisterStructValidation(fn validatorPkg.StructLevelFunc, types ...interface{})
}

func NewValidator() Interface {
	return validatorPkg.New()
}
