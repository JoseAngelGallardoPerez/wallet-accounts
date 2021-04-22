package request

import (
	"github.com/Confialink/wallet-accounts/internal/modules/app/validator"
	"github.com/Confialink/wallet-accounts/internal/modules/request/form"
	"github.com/gin-gonic/gin/binding"
)

func init() {
	binding.Validator.Engine().(validator.Interface).RegisterStructValidation(form.OwtStructLevelValidation, form.OWT{})
}
