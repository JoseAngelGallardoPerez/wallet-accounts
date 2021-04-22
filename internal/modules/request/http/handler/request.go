package handler

import (
	"github.com/Confialink/wallet-accounts/internal/modules/request/event"
	"github.com/Confialink/wallet-accounts/internal/modules/request/transfers"
	"github.com/Confialink/wallet-accounts/internal/modules/request/view"
	"github.com/Confialink/wallet-accounts/internal/transfer"
	"github.com/olebedev/emitter"
	"net/http"

	"github.com/Confialink/wallet-pkg-errors"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	"github.com/Confialink/wallet-accounts/internal/modules/app/http/response"
	"github.com/Confialink/wallet-accounts/internal/modules/app/http/service"
	requestRepository "github.com/Confialink/wallet-accounts/internal/modules/request/repository"
	requestService "github.com/Confialink/wallet-accounts/internal/modules/request/service"
)

type RequestHandler struct {
	View              view.View
	contextService    service.ContextInterface
	requestRepository requestRepository.RequestRepositoryInterface
	executor          *requestService.Executor
	canceller         *requestService.Canceller
	db                *gorm.DB
	currencyProvider  transfer.CurrencyProvider
	emitter           *emitter.Emitter
	pf                transfers.PermissionFactory
}

func NewRequestHandler(
	contextService service.ContextInterface,
	requestRepository requestRepository.RequestRepositoryInterface,
	View view.View,
	db *gorm.DB,
	executor *requestService.Executor,
	canceller *requestService.Canceller,
	currencyProvider transfer.CurrencyProvider,
	emitter *emitter.Emitter,
	pf transfers.PermissionFactory,
) *RequestHandler {
	return &RequestHandler{
		View:              View,
		contextService:    contextService,
		requestRepository: requestRepository,
		executor:          executor,
		canceller:         canceller,
		db:                db,
		currencyProvider:  currencyProvider,
		emitter:           emitter,
		pf:                pf,
	}
}

func (r *RequestHandler) ModifyRequest(c *gin.Context) {
	req := r.contextService.GetRequestedRequest(c)
	if req == nil {
		return
	}
	//user := r.contextService.MustGetCurrentUser(c)

	rateForm := &struct {
		Rate string `json:"rate" binding:"required,decimal,decimalGT=0"`
	}{}

	if err := c.ShouldBind(rateForm); err != nil {
		errors.AddShouldBindError(c, err)
		return
	}

	rate, _ := decimal.NewFromString(rateForm.Rate)
	if !req.Rate.Equal(rate) {
		req.Rate = &rate

		tx := r.db.Begin()
		modifier, err := transfers.CreateModifier(tx, req, r.currencyProvider, r.pf)
		if err != nil {
			tx.Rollback()
			errors.AddErrors(c, errcodes.ConvertToTyped(err))
			return
		}
		details, err := modifier.Modify(req)
		if err != nil {
			tx.Rollback()
			errors.AddErrors(c, errcodes.ConvertToTyped(err))
			return
		}

		eventContext := &event.ContextRequestModified{
			Tx:      tx,
			Request: req,
			Details: details,
		}
		<-r.emitter.Emit(event.RequestModified, eventContext)
		tx.Commit()
	}

	c.JSON(http.StatusOK, response.New().SetData(req))
}

func (r *RequestHandler) CancelRequest(c *gin.Context) {
	req := r.contextService.GetRequestedRequest(c)
	if req == nil {
		return
	}
	user := r.contextService.MustGetCurrentUser(c)

	cancelForm := &struct {
		Reason string `json:"reason,omitempty"`
	}{}
	if err := c.ShouldBind(cancelForm); err != nil {
		errors.AddShouldBindError(c, err)
		return
	}

	if err := r.canceller.Call(req, cancelForm.Reason, user); err != nil {
		errors.AddErrors(c, errcodes.ConvertToTyped(err))
		return
	}

	c.JSON(http.StatusOK, response.New().SetData(req))
}

func (r *RequestHandler) ExecuteRequest(c *gin.Context) {
	req := r.contextService.GetRequestedRequest(c)
	if req == nil {
		return
	}

	user := r.contextService.MustGetCurrentUser(c)
	if err := r.executor.Call(req, user); err != nil {
		errors.AddErrors(c, errcodes.ConvertToTyped(err))
		return
	}

	c.JSON(http.StatusOK, response.New().SetData(req))
}

func (r *RequestHandler) ViewRequest(c *gin.Context) {
	user := r.contextService.MustGetCurrentUser(c)
	req := r.contextService.GetRequestedRequest(c)
	if req == nil {
		return
	}

	data, err := r.View.View(req, user)
	if err != nil {
		errors.AddErrors(c, errcodes.ConvertToTyped(err))
		return
	}

	c.JSON(http.StatusOK, response.New().SetData(data))
}
