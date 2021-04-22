package handler

import (
	"fmt"
	"net/http"

	"github.com/Confialink/wallet-accounts/internal/modules/request/constants"
	"github.com/Confialink/wallet-pkg-list_params"

	"github.com/Confialink/wallet-pkg-errors"
	"github.com/Confialink/wallet-pkg-model_serializer"
	"github.com/Confialink/wallet-users/rpc/proto/users"
	"github.com/gin-gonic/gin"
	log "github.com/inconshreveable/log15"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	"github.com/Confialink/wallet-accounts/internal/modules/app/http/response"
	"github.com/Confialink/wallet-accounts/internal/modules/app/http/service"
	"github.com/Confialink/wallet-accounts/internal/modules/request/filters"
	"github.com/Confialink/wallet-accounts/internal/modules/request/model"
	requestRepository "github.com/Confialink/wallet-accounts/internal/modules/request/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/request/serializer"
)

type ListHandler struct {
	contextService    service.ContextInterface
	logger            log.Logger
	requestRepository requestRepository.RequestRepositoryInterface
	params            *HandlerPrams
}

func NewListHandler(
	contextService service.ContextInterface,
	logger log.Logger,
	requestRepository requestRepository.RequestRepositoryInterface,
	params *HandlerPrams,
) *ListHandler {
	return &ListHandler{
		contextService:    contextService,
		logger:            logger,
		requestRepository: requestRepository,
		params:            params,
	}
}

func (l *ListHandler) ListAdmin(c *gin.Context) {
	user := l.contextService.MustGetCurrentUser(c)

	listParams := l.params.forAdmin(c.Request.URL.RawQuery)
	// TODO: refactor is_visible field is actually used in order to filter requests which are initiated by admin
	// or has been approved without confirmation, we should remove is_visible field and add instantly_approved field instead
	listParams.AddFilter("isVisible", []string{"true"})
	listParams.AddFilter("status", []string{constants.StatusNew}, list_params.OperatorNin)
	if ok, errorsList := listParams.Validate(); !ok {
		errcodes.AddErrorMeta(c, errcodes.CodeInvalidQueryParameters, errorsList)
	}

	items, err := l.requestRepository.GetList(listParams)
	if err != nil {
		errors.AddErrors(c, &errors.PrivateError{Message: fmt.Sprintf("list handler. Unable to find requests. Error: %s", err.Error())})
		return
	}

	count, err := l.requestRepository.GetListCount(listParams)
	if err != nil {
		errors.AddErrors(c, &errors.PrivateError{Message: fmt.Sprintf("list handler. Unable to get count of requests. Error: %s", err.Error())})
		return
	}

	serialized := serializeRequests(items, listParams.GetOutputFields(), user, l.logger)
	resp := response.NewWithListAndPageLinks(serialized, count,
		c.Request.URL.RequestURI(), listParams)
	c.JSON(http.StatusOK, resp)
}

func (l *ListHandler) ListUser(c *gin.Context) {
	user := l.contextService.MustGetCurrentUser(c)

	listParams := l.params.forClient(c.Request.URL.RawQuery)

	if ok, errorsList := listParams.Validate(); !ok {
		errcodes.AddErrorMeta(c, errcodes.CodeInvalidQueryParameters, errorsList)
		return
	}

	// this is crucial
	listParams.AddCustomFilter("userId", filters.UserIdEq)
	listParams.AddCustomFilter("incomingByStatus", filters.IncomingByStatus)
	listParams.AddFilter("userId", []string{user.UID})
	listParams.AddFilter("incomingByStatus", []string{constants.StatusExecuted})

	items, err := l.requestRepository.GetList(listParams)
	if err != nil {
		privateError := errors.PrivateError{Message: "list handler. Unable to find requests."}
		privateError.AddLogPair("error", err.Error())
		errors.AddErrors(c, &privateError)
		return
	}

	count, err := l.requestRepository.GetListCount(listParams)
	if err != nil {
		privateError := errors.PrivateError{Message: "list handler. Unable to get count of requests."}
		privateError.AddLogPair("error", err.Error())
		errors.AddErrors(c, &privateError)
		return
	}

	serialized := serializeRequests(items, listParams.GetOutputFields(), user, l.logger)
	resp := response.NewWithListAndPageLinks(serialized, count,
		c.Request.URL.RequestURI(), listParams)
	c.JSON(http.StatusOK, resp)
}

func serializeRequests(requests []*model.Request, fields []interface{}, user *users.User, logger log.Logger) []interface{} {
	result := make([]interface{}, len(requests))
	fields = append(fields, serializer.ProvideSnapshotsSerializer(user, logger))
	fields = append(fields, serializer.ProvideAmountSerializer())
	fields = append(fields, serializer.ProvideBalanceDifferenceSerializer())
	for i, v := range requests {
		result[i] = model_serializer.Serialize(v, fields)
	}
	return result
}
