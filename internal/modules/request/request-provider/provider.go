package request_provider

import (
	"github.com/Confialink/wallet-accounts/internal/modules/request"
	"github.com/Confialink/wallet-accounts/internal/modules/request/http/handler"
	"github.com/Confialink/wallet-accounts/internal/modules/request/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/request/service"
	"github.com/Confialink/wallet-accounts/internal/modules/request/transfers"
	"github.com/Confialink/wallet-accounts/internal/modules/request/view"
)

func Providers() []interface{} {
	return []interface{}{
		repository.NewRequestRepository,
		repository.NewDataOwt,
		repository.NewTemplate,
		request.NewCreator,
		request.NewCsvService,
		service.NewRequestsService,
		service.NewIncludes,
		service.NewExecutor,
		service.NewCanceller,

		//request.View
		view.NewDefaultView,
		request.NewDataPresenter,

		handler.NewHandlerPrams,
		handler.NewListHandler,
		handler.NewCaHandler,
		handler.NewCftHandler,
		handler.NewDaHandler,
		handler.NewTbaHandler,
		handler.NewTbuHandler,
		handler.NewMoneyRequestTbuHandler,
		handler.NewOwtHandler,
		handler.NewDraHandler,
		handler.NewRequestHandler,
		handler.NewTemplateHandler,
		handler.NewCsvHandler,

		transfers.NewDefaultPermissionFactory,
	}
}
