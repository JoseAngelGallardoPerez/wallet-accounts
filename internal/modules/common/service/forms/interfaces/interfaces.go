package interfaces

import (
	"github.com/Confialink/wallet-pkg-custom_form"
)

// contains form providers for a model
type FormProviderAgregator interface {
	// name of model
	Name() string

	// map of form providers
	FormProviders() map[string]FormProvider
}

// provides a form for a model
type FormProvider interface {
	// make custom form
	MakeForm() (*custom_form.Form, error)
}
