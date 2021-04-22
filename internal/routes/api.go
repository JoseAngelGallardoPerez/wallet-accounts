package routes

import (
	"net/http"

	errors "github.com/Confialink/wallet-pkg-errors"
	service_names "github.com/Confialink/wallet-pkg-service_names"
	"github.com/gin-gonic/gin"
	"github.com/inconshreveable/log15"

	"github.com/Confialink/wallet-accounts/internal/authentication"
	"github.com/Confialink/wallet-accounts/internal/config"
	accountTypeHandler "github.com/Confialink/wallet-accounts/internal/modules/account-type/http/handler"
	accountHandler "github.com/Confialink/wallet-accounts/internal/modules/account/http/handler"
	accountMw "github.com/Confialink/wallet-accounts/internal/modules/account/http/middleware"
	accountRepo "github.com/Confialink/wallet-accounts/internal/modules/account/repository"
	appHandler "github.com/Confialink/wallet-accounts/internal/modules/app/http/handler"
	appMiddleware "github.com/Confialink/wallet-accounts/internal/modules/app/http/middleware"
	"github.com/Confialink/wallet-accounts/internal/modules/app/http/service"
	authMiddleware "github.com/Confialink/wallet-accounts/internal/modules/auth/middleware"
	authS "github.com/Confialink/wallet-accounts/internal/modules/auth/service"
	bankDetailsHandler "github.com/Confialink/wallet-accounts/internal/modules/bank-details/http/handler"
	cardTypeCategoryHandler "github.com/Confialink/wallet-accounts/internal/modules/card-type-category/http/handler"
	cardTypeFormatHandler "github.com/Confialink/wallet-accounts/internal/modules/card-type-format/http/handler"
	cardTypeHandler "github.com/Confialink/wallet-accounts/internal/modules/card-type/http/handler"
	cardHandlers "github.com/Confialink/wallet-accounts/internal/modules/card/http/handlers"
	cardMw "github.com/Confialink/wallet-accounts/internal/modules/card/http/middleware"
	cardRepo "github.com/Confialink/wallet-accounts/internal/modules/card/repository"
	commonHandlers "github.com/Confialink/wallet-accounts/internal/modules/common/http/handlers"
	countryHandler "github.com/Confialink/wallet-accounts/internal/modules/country/http/handler"
	feeHandler "github.com/Confialink/wallet-accounts/internal/modules/fee/http/handler"
	moneyRequestHdlr "github.com/Confialink/wallet-accounts/internal/modules/moneyrequest/http/handler"
	paymentMethodHandler "github.com/Confialink/wallet-accounts/internal/modules/payment-method/http/handler"
	paymentPeriodHandler "github.com/Confialink/wallet-accounts/internal/modules/payment-period/http/handler"
	"github.com/Confialink/wallet-accounts/internal/modules/permission"
	requestHandler "github.com/Confialink/wallet-accounts/internal/modules/request/http/handler"
	requestMw "github.com/Confialink/wallet-accounts/internal/modules/request/http/middleware"
	requestRepo "github.com/Confialink/wallet-accounts/internal/modules/request/repository"
	scheduledTransactionsHandler "github.com/Confialink/wallet-accounts/internal/modules/scheduled-transaction/http/handler"
	"github.com/Confialink/wallet-accounts/internal/modules/settings"
	settingsController "github.com/Confialink/wallet-accounts/internal/modules/settings/http/handler"
	"github.com/Confialink/wallet-accounts/internal/modules/tan"
	tanHandler "github.com/Confialink/wallet-accounts/internal/modules/tan/handler"
	transactionHandler "github.com/Confialink/wallet-accounts/internal/modules/transaction/http/handler"
	transactionMw "github.com/Confialink/wallet-accounts/internal/modules/transaction/http/middleware"
	transactionRepo "github.com/Confialink/wallet-accounts/internal/modules/transaction/repository"
	"github.com/Confialink/wallet-accounts/version"
)

func NewAPIRouter(
	logger log15.Logger,
	config *config.Config,
	tanService *tan.Service,
	contextService service.ContextInterface,
	settingsService *settings.Service,
	accountsHandler *accountHandler.AccountHandler,
	accountsCsvHandler *accountHandler.CsvHandler,
	accountsTypeHandler *accountTypeHandler.AccountTypeHandler,
	cardHandler *cardHandlers.CardHandler,
	cardListHandler *cardHandlers.CardListHandler,
	cardTypeHandler *cardTypeHandler.CardTypeHandler,
	cardTypeCategoryHandler *cardTypeCategoryHandler.CardTypeCategoryHandler,
	cardTypeFormatHandler *cardTypeFormatHandler.CardTypeFormatHandler,
	paymentMethodHandler *paymentMethodHandler.PaymentMethodHandler,
	paymentPeriodHandler *paymentPeriodHandler.PaymentPeriodHandler,
	settingsHandler *settingsController.SettingsController,
	tanHandler *tanHandler.Controller,
	bankDetailsHandler *bankDetailsHandler.IwtBankAccountHandler,
	revenueAccountHandler *accountHandler.RevenueAccountHandler,
	requestListHandler *requestHandler.ListHandler,
	requestHandler *requestHandler.RequestHandler,
	tbaHandler *requestHandler.TbaHandler,
	tbuHandler *requestHandler.TbuHandler,
	owtHandler *requestHandler.OwtHandler,
	cftHandler *requestHandler.CftHandler,
	caHandler *requestHandler.CaHandler,
	daHandler *requestHandler.DaHandler,
	draHandler *requestHandler.DraHandler,
	corsHandler *appHandler.CorsHandler,
	notFoundHandler *appHandler.NotFoundHandler,
	transactionsHandler *transactionHandler.TransactionHandler,
	transactionsHistoryHandler *transactionHandler.HistoryHandler,
	transactionsCsvHandler *transactionHandler.CsvHandler,
	countryHandler *countryHandler.CountryHandler,
	transferFeeHandler *feeHandler.TransferFee,
	templateHandler *requestHandler.TemplateHandler,
	requestCsvHandler *requestHandler.CsvHandler,
	cardsCsvHandler *cardHandlers.CsvHandler,
	scheduledTxHandler *scheduledTransactionsHandler.TransactionsHandler,
	authService authS.AuthServiceInterface,
	accountRepo *accountRepo.AccountRepository,
	cardRepo cardRepo.CardRepositoryInterface,
	requestRepo requestRepo.RequestRepositoryInterface,
	transactionRepo *transactionRepo.TransactionRepository,
	modelFormHndlr *commonHandlers.ModelFormHandler,
	moneyRequestHandler *moneyRequestHdlr.MoneyRequest,
	moneyRequestTBUHandler *requestHandler.MoneyRequestTbuHandler,
) *gin.Engine {
	r := gin.New()
	logger = logger.New("where", "routes.api")
	// Middleware

	mwAuth := authentication.Middleware(logger.New("Middleware", "Auth"))
	mwCors := appMiddleware.CorsMiddleware(config.Cors)
	mwAdminRoot := authMiddleware.AdminOrRoot()
	mwClient := authMiddleware.Client()
	mwPerm := authMiddleware.NewPermissionChecker(authService, contextService)

	// Routes

	r.GET("/accounts/health-check", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	r.GET("/accounts/build", func(c *gin.Context) {
		c.JSON(http.StatusOK, version.BuildInfo)
	})

	apiGroup := r.Group(service_names.Accounts.Internal, mwCors)
	apiGroup.Use(
		gin.Recovery(),
		gin.Logger(),
		errors.ErrorHandler(logger.New("Middleware", "Errors")),
	)

	privateGroup := apiGroup.Group("/private", mwAuth)
	{
		v1Group := privateGroup.Group("/v1")
		{
			adminGroup := v1Group.Group("/admin", mwAdminRoot)
			userGroup := v1Group.Group("/user")

			mwPermCreateAccount := mwPerm.CanDynamic(authS.ActionHas, authS.ResourcePermission, permission.CreateAccounts)

			mwPermViewSettings := mwPerm.CanDynamic(authS.ActionHas, authS.ResourcePermission, permission.ViewSettings)
			mwPermModifySettings := mwPerm.CanDynamic(authS.ActionHas, authS.ResourcePermission, permission.ModifySettings)
			mwPermCreateSettings := mwPerm.CanDynamic(authS.ActionHas, authS.ResourcePermission, permission.CreateSettings)
			mwPermRemoveSettings := mwPerm.CanDynamic(authS.ActionHas, authS.ResourcePermission, permission.RemoveSettings)

			mwPermManualDebitCredit := mwPerm.CanDynamic(authS.ActionHas, authS.ResourcePermission, permission.ManualDebitCreditAccounts)
			mwPermCreateCard := mwPerm.CanDynamic(authS.ActionHas, authS.ResourcePermission, permission.CreateCards)
			mwPermReadCardType := mwPerm.Can(authS.ActionRead, authS.CardTypeResource)

			usersGroup := v1Group.Group("/users")
			{
				usersGroup.GET("/:uid/accounts", accountsHandler.ListForUserHandler)
			}

			adminAccountsGroup := adminGroup.Group("/accounts")
			{
				mwPermViewAccounts := mwPerm.CanDynamic(authS.ActionReadList, authS.AccountsResource, nil)
				adminAccountsGroup.GET("", mwPermViewAccounts, accountsHandler.AdminListHandler)
			}

			accountsGroup := v1Group.Group("/accounts")
			{
				mwRequestedAccount := accountMw.RequestedAccount(contextService, accountRepo)
				accountsGroup.GET("/:id", mwRequestedAccount, mwPerm.CanDynamicWithAccount(authS.ActionRead, authS.AccountsResource), accountsHandler.GetHandler)
				accountsGroup.GET("", mwClient, accountsHandler.ListHandler)
				accountsGroup.POST("", mwAdminRoot, mwPermCreateAccount, accountsHandler.CreateHandler)
				update(accountsGroup, "/:id", mwAdminRoot, mwRequestedAccount, mwPerm.CanDynamic(authS.ActionHas, authS.ResourcePermission, permission.ModifyAccounts), accountsHandler.UpdateHandler)
				accountsGroup.DELETE("/:id", mwAdminRoot, accountsHandler.DeleteHandler)
			}

			accountTypesGroup := v1Group.Group("/account-types")
			{
				mwPermCreateModifyAccountTypes := mwPerm.CanDynamic(authS.ActionHas, authS.ResourcePermission, permission.CreateModifyAccountTypes)
				accountTypesGroup.GET("/:id", accountsTypeHandler.GetHandler)
				accountTypesGroup.GET("", accountsTypeHandler.ListHandler)
				accountTypesGroup.POST("", mwAdminRoot, mwPermCreateModifyAccountTypes, accountsTypeHandler.CreateHandler)
				update(accountTypesGroup, "/:id", mwAdminRoot, mwPermCreateModifyAccountTypes, accountsTypeHandler.UpdateHandler)
				accountTypesGroup.DELETE("/:id", mwAdminRoot, mwPermCreateModifyAccountTypes, accountsTypeHandler.DeleteHandler)
			}

			cardTypesGroup := v1Group.Group("/card-types")
			{
				cardTypesGroup.POST("", mwAdminRoot, mwPermCreateSettings, cardTypeHandler.CreateHandler)
				cardTypesGroup.GET("/:id", mwPermReadCardType, cardTypeHandler.ShowHandler)
				update(cardTypesGroup, "/:id", mwAdminRoot, mwPermModifySettings, cardTypeHandler.UpdateHandler)
				cardTypesGroup.GET("", mwPermReadCardType, cardTypeHandler.ListHandler)
				cardTypesGroup.DELETE("/:id", mwAdminRoot, mwPermRemoveSettings, cardTypeHandler.DeleteHandler)
			}

			adminCardTypesGroup := adminGroup.Group("/card-types", mwPermReadCardType)
			{
				adminCardTypesGroup.GET("/formats", cardTypeFormatHandler.ListHandler)
				adminCardTypesGroup.GET("/categories", cardTypeCategoryHandler.ListHandler)
			}

			cardsGroup := v1Group.Group("/cards")
			{
				mwRequestedCard := cardMw.RequestedCard(contextService, cardRepo)
				cardsGroup.POST("", mwAdminRoot, mwPermCreateCard, cardHandler.CreateCardHandler)
				cardsGroup.GET("/:id", mwRequestedCard, mwPerm.CanDynamicWithCard(authS.ActionRead, authS.CardResource), cardHandler.ShowCardHandler)
				cardsGroup.GET("", mwAdminRoot, mwPerm.CanDynamic(authS.ActionHas, authS.ResourcePermission, permission.ViewCards), cardListHandler.IndexCardsHandler)
				update(cardsGroup, "/:id", mwAdminRoot, mwRequestedCard, mwPerm.CanDynamic(authS.ActionHas, authS.ResourcePermission, permission.ModifyCards), cardHandler.UpdateCardHandler)
			}

			settingsGroup := v1Group.Group("/settings")
			{
				settingsGroup.POST("/:setting", mwAdminRoot, mwPermModifySettings, settingsHandler.UpdateSetting)
				settingsGroup.POST("", mwAdminRoot, mwPermModifySettings, settingsHandler.MassUpdateSetting)
				settingsGroup.GET("/:setting", mwPerm.CanDynamicWithParam(authS.ActionRead, authS.ResourceSetting, "setting"), settingsHandler.GetSetting)
				settingsGroup.GET("", mwAdminRoot, mwPermViewSettings, settingsHandler.ListSettings)
			}

			userTanGroup := userGroup.Group("/tan")
			{
				userTanGroup.GET("/count", tanHandler.GetOwnCount)
				userTanGroup.GET("/request/availability", tanHandler.UserCanRequestTan)
				userTanGroup.POST("", tanHandler.UserRequestOne)
			}

			adminTanGroup := adminGroup.Group("/tan")
			{
				adminTanGroup.GET("/count/:userId", mwPerm.CanDynamic(authS.ActionHas, authS.ResourcePermission, permission.ViewUserProfiles), tanHandler.GetCount)
				adminTanGroup.POST("/:userId", mwPerm.CanDynamic(authS.ActionHas, authS.ResourcePermission, permission.GenerateSendNewTans), tanHandler.Create)
			}

			v1Group.GET("/own-cards", mwClient, cardListHandler.IndexOwnCardsHandler)

			tbaRequestsGroup := v1Group.Group("/tba-requests", mwClient)
			{
				mwUseTan := tan.MiddlewareUseIfRequired(
					tanService,
					contextService,
					settingsService,
					"tba_tan_required",
				)
				tbaRequestsGroup.POST("/preview", tbaHandler.CreatePreviewUser)
				tbaRequestsGroup.POST("", mwUseTan, tbaHandler.CreateRequestUser)
			}

			mwInitiateExecuteUserTransfers := mwPerm.CanDynamic(authS.ActionHas, authS.ResourcePermission, permission.InitiateExecuteUserTransfers)
			mwExecuteCancelPendingTransferRequests := mwPerm.CanDynamic(authS.ActionHas, authS.ResourcePermission, permission.ExecuteCancelPendingTransferRequests)

			tbaRequestsAdminGroup := adminGroup.Group("/tba-requests", mwInitiateExecuteUserTransfers)
			{
				tbaRequestsAdminGroup.POST("/preview/user/:userId", tbaHandler.CreatePreviewAdmin)

				tbaRequestsAdminGroup.POST("/user/:userId", tbaHandler.CreateRequestAdmin)
			}

			tbuRequestsGroup := v1Group.Group("/tbu-requests", mwClient)
			{
				mwUseTan := tan.MiddlewareUseIfRequired(
					tanService,
					contextService,
					settingsService,
					"tbu_tan_required",
				)
				tbuRequestsGroup.POST("/preview", tbuHandler.CreatePreviewUser)
				tbuRequestsGroup.POST("", mwUseTan, tbuHandler.CreateRequestUser)
				tbuRequestsGroup.POST("/receive", tbuHandler.Receive)
			}

			tbuRequestsAdminGroup := adminGroup.Group("/tbu-requests", mwInitiateExecuteUserTransfers)
			{
				tbuRequestsAdminGroup.POST("/preview/user/:userId", tbuHandler.CreatePreviewAdmin)
				tbuRequestsAdminGroup.POST("/user/:userId", tbuHandler.CreateRequestAdmin)
			}

			tbuMoneyRequestsGroup := v1Group.Group("tbu-money-requests", mwClient)
			{
				mwUseTan := tan.MiddlewareUseIfRequired(
					tanService,
					contextService,
					settingsService,
					"tbu_tan_required",
				)
				tbuMoneyRequestsGroup.POST("/preview", moneyRequestTBUHandler.CreatePreviewUser)
				tbuMoneyRequestsGroup.POST("", mwUseTan, moneyRequestTBUHandler.CreateRequestUser)
			}

			owtRequestsAdminGroup := adminGroup.Group("/owt-requests", mwInitiateExecuteUserTransfers)
			{
				owtRequestsAdminGroup.POST("/preview/user/:userId", owtHandler.CreatePreviewAdmin)
				owtRequestsAdminGroup.POST("/user/:userId", owtHandler.CreateRequestAdmin)
			}

			owtRequestsGroup := v1Group.Group("/owt-requests", mwClient)
			{
				mwUseTan := tan.MiddlewareUseIfRequired(
					tanService,
					contextService,
					settingsService,
					"owt_tan_required",
				)
				owtRequestsGroup.POST("/preview", owtHandler.CreatePreviewUser)
				owtRequestsGroup.POST("", mwUseTan, owtHandler.CreateRequestUser)
			}

			cftRequestsAdminGroup := adminGroup.Group("/cft-requests", mwInitiateExecuteUserTransfers)
			{
				cftRequestsAdminGroup.POST("/preview/user/:userId", cftHandler.CreatePreviewAdmin)
				cftRequestsAdminGroup.POST("/user/:userId", cftHandler.CreateRequestAdmin)
			}

			cftRequestsGroup := v1Group.Group("/cft-requests", mwClient)
			{
				mwUseTan := tan.MiddlewareUseIfRequired(
					tanService,
					contextService,
					settingsService,
					"cft_tan_required",
				)
				cftRequestsGroup.POST("/preview", cftHandler.CreatePreviewUser)
				cftRequestsGroup.POST("", mwUseTan, cftHandler.CreateRequestUser)
			}

			caRequestsAdminGroup := adminGroup.Group("/ca-requests", mwPermManualDebitCredit)
			{
				caRequestsAdminGroup.POST("/preview", caHandler.CreatePreviewAdmin)
				caRequestsAdminGroup.POST("", caHandler.CreateRequestAdmin)
			}

			daRequestsAdminGroup := adminGroup.Group("/da-requests", mwPermManualDebitCredit)
			{
				daRequestsAdminGroup.POST("/preview", daHandler.CreatePreviewAdmin)
				daRequestsAdminGroup.POST("", daHandler.CreateRequestAdmin)
			}

			draRequestsAdminGroup := adminGroup.Group("/dra-requests", mwPerm.CanDynamic(authS.ActionUpdate, authS.ResourceRevenueAccount, nil))
			{
				draRequestsAdminGroup.POST("/preview", draHandler.CreatePreviewAdmin)
				draRequestsAdminGroup.POST("", draHandler.CreateRequestAdmin)
			}

			mwRequestedRequest := requestMw.RequestedRequest(requestRepo)
			requestsGroup := v1Group.Group("/requests")
			{
				requestsGroup.GET("", mwAdminRoot, mwPerm.CanDynamic(authS.ActionReadList, authS.ResourceRequest, nil), requestListHandler.ListAdmin)
				requestsGroup.GET("/:requestId", mwRequestedRequest, mwPerm.CanDynamicWithRequest(authS.ActionRead, authS.ResourceTransferRequest), requestHandler.ViewRequest)
			}

			requestsAdminGroup := adminGroup.Group("/requests")
			{
				requestsAdminGroup.POST("/csv/update", mwPerm.CanDynamic(authS.ActionHas, authS.ResourcePermission, permission.ImportTransferRequests), requestCsvHandler.UpdateFromCsv)
				requestsAdminGroup.POST("/csv/import", mwPermManualDebitCredit, requestCsvHandler.ImportFromCsv)
				requestsAdminGroup.POST("/cancel/:requestId", mwRequestedRequest, mwExecuteCancelPendingTransferRequests, requestHandler.CancelRequest)
				requestsAdminGroup.POST("/execute/:requestId", mwRequestedRequest, mwExecuteCancelPendingTransferRequests, requestHandler.ExecuteRequest)
				requestsAdminGroup.PATCH("/:requestId", mwRequestedRequest, requestHandler.ModifyRequest)
			}

			userTransactionsGroup := userGroup.Group("/transactions")
			{
				mwRequestedTransaction := transactionMw.RequestedTransaction(contextService, transactionRepo)
				userTransactionsGroup.GET("", mwPerm.CanDynamic(authS.ActionReadList, authS.ResourceTransaction, nil), transactionsHandler.ShowUserList)
				userTransactionsGroup.GET("/:id", mwRequestedTransaction, mwPerm.CanDynamicWithTransaction(authS.ActionRead, authS.ResourceTransaction), transactionsHandler.GetOne)
			}

			userTransactionsHistoryGroup := userGroup.Group("/transactions-history")
			{
				userTransactionsHistoryGroup.GET("", transactionsHistoryHandler.ShowUserHistory)
				userTransactionsHistoryGroup.GET("/export", transactionsCsvHandler.ExportHistoryToCsv)
			}

			templatesGroup := userGroup.Group("/templates")
			{
				templatesGroup.GET("/", templateHandler.ListAll)
				templatesGroup.GET("/:subject", templateHandler.List)
				templatesGroup.DELETE("/:id", templateHandler.Delete)
			}

			userRequestsGroup := userGroup.Group("/requests")
			{
				userRequestsGroup.GET("", requestListHandler.ListUser)
			}

			paymentMethods := v1Group.Group("/payment-methods")
			{
				paymentMethods.GET("", mwPerm.CanDynamic(authS.ActionReadList, authS.PaymentMethodsResource, nil), paymentMethodHandler.ListHandler)
			}

			paymentPeriods := v1Group.Group("/payment-periods")
			{
				paymentPeriods.GET("", mwPerm.CanDynamic(authS.ActionReadList, authS.PaymentPeriodsResource, nil), paymentPeriodHandler.ListHandler)
			}

			adminFeeGroup := adminGroup.Group("/fee")
			{
				adminFeeGroup.POST("/transfer", mwPermCreateSettings, transferFeeHandler.Create)
				adminFeeGroup.POST("/transfer/id/:id", mwPermModifySettings, transferFeeHandler.Update)
				adminFeeGroup.GET("/transfer/id/:id", mwPermViewSettings, transferFeeHandler.GetFee)
				adminFeeGroup.DELETE("/transfer/id/:id", mwPermRemoveSettings, transferFeeHandler.DeleteFee)
				adminFeeGroup.GET("/transfer/subject/:requestSubject", mwPermViewSettings, transferFeeHandler.ListFees)
				adminFeeGroup.GET("/transfer/parameters/:id", mwPermViewSettings, transferFeeHandler.ListFeeParameters)
			}

			userFeeGroup := userGroup.Group("/fee")
			{
				userFeeGroup.GET("/transfer/subject/:requestSubject", transferFeeHandler.ListUserFees)
			}

			iwtBankAccounts := v1Group.Group("/iwt-bank-accounts")
			{
				iwtBankAccounts.GET("/:id", mwPerm.CanDynamic(authS.ActionRead, authS.ResourceIwtBankAccount, nil), bankDetailsHandler.GetHandler)
				iwtBankAccounts.GET("", mwPerm.CanDynamic(authS.ActionReadList, authS.ResourceIwtBankAccount, nil), bankDetailsHandler.ListHandler)
				iwtBankAccounts.POST("", mwAdminRoot, mwPerm.CanDynamic(authS.ActionCreate, authS.ResourceIwtBankAccount, permission.CreateModifyIwtBankAccounts), bankDetailsHandler.CreateHandler)
				update(iwtBankAccounts, "/:id", mwAdminRoot, mwPerm.CanDynamic(authS.ActionUpdate, authS.ResourceIwtBankAccount, permission.CreateModifyIwtBankAccounts), bankDetailsHandler.UpdateHandler)
				iwtBankAccounts.DELETE("/:id", mwPerm.CanDynamic(authS.ActionDelete, authS.ResourceIwtBankAccount, permission.CreateModifyIwtBankAccounts), bankDetailsHandler.DeleteHandler)
				iwtBankAccounts.GET("/:id/by-account-id", mwPerm.CanDynamic(authS.ActionReadList, authS.ResourceIwtBankAccount, nil), bankDetailsHandler.GetByAccountIdHandler)
				iwtBankAccounts.GET("/:id/pdf-for-account/:accountId", mwPerm.CanDynamic(authS.ActionRead, authS.ResourceIwtBankAccount, nil), bankDetailsHandler.PdfForAccount)
			}

			revenueAccounts := v1Group.Group("/revenue-accounts", mwAdminRoot)
			{
				revenueAccounts.GET("", mwPerm.CanDynamic(authS.ActionReadList, authS.ResourceRevenueAccount, nil), revenueAccountHandler.ListHandler)
				revenueAccounts.GET("/:id", mwPerm.CanDynamic(authS.ActionRead, authS.ResourceRevenueAccount, nil), revenueAccountHandler.GetHandler)
			}

			adminExportGroup := adminGroup.Group("export")
			{
				mwPermViewAccounts := mwPerm.CanDynamic(authS.ActionHas, authS.ResourcePermission, permission.ViewAccounts)
				adminExportGroup.GET("/accounts", mwPermViewAccounts, accountsCsvHandler.AdminsExport)
				adminExportGroup.GET("/transfer-requests", mwPerm.CanDynamic(authS.ActionReadList, authS.ResourceRequest, nil), requestCsvHandler.DownloadAdminsReport)
				adminExportGroup.GET("/cards", mwPerm.Can(authS.ActionReadList, authS.CardResource), cardsCsvHandler.AdminsExport)
				adminExportGroup.GET("/scheduled-transactions", mwPerm.CanDynamic(authS.ActionReadList, authS.ResourceScheduledTransactions, nil), scheduledTxHandler.ExportToCsv)
			}

			userExportGroup := userGroup.Group("export")
			{
				userExportGroup.GET("/transactions-report", mwPerm.CanDynamic(authS.ActionReadList, authS.ResourceTransaction, nil), transactionsCsvHandler.ExportToCsv)
			}

			importGroup := v1Group.Group("import")
			{
				importGroup.POST("/accounts", mwAdminRoot, mwPermCreateAccount, accountsHandler.ImportCsvHandler)
				importGroup.POST("/cards", mwAdminRoot, mwPermCreateCard, cardHandler.ImportCsvHandler)
			}

			generateGroup := v1Group.Group("generate")
			{
				accountsGroup := generateGroup.Group("accounts")
				{
					accountsGroup.POST("/number", mwAdminRoot, mwPermCreateAccount, accountsHandler.GenerateNumberHandler)
				}
			}

			scheduledTransactionsGroup := v1Group.Group("scheduled-transactions")
			{
				scheduledTransactionsGroup.GET("", mwPerm.CanDynamic(authS.ActionReadList, authS.ResourceScheduledTransactions, nil), scheduledTxHandler.ListHandler)
				scheduledTransactionsGroup.GET("/:id", mwPerm.CanDynamic(authS.ActionReadList, authS.ResourceScheduledTransactions, nil), scheduledTxHandler.GetById)
			}

			formsGroup := v1Group.Group("/form")
			{
				formsGroup.GET("/:model/:type", modelFormHndlr.FieldsHandler)
			}

			moneyRequestsGroup := v1Group.Group("/money-requests")
			{
				moneyRequestsGroup.POST("", moneyRequestHandler.Create)
				moneyRequestsGroup.GET("/id/:id", moneyRequestHandler.Show)
				moneyRequestsGroup.PUT("/id/:id/mark-old", moneyRequestHandler.MarkOld)
				moneyRequestsGroup.GET("/incoming", moneyRequestHandler.Incoming)
				moneyRequestsGroup.GET("/outgoing", moneyRequestHandler.Outgoing)
			}
		}
	}

	public := apiGroup.Group("/public")
	{
		v1Group := public.Group("/v1")
		{
			countries := v1Group.Group("/countries")
			{
				countries.GET("/:id", countryHandler.GetHandler)
				countries.GET("", countryHandler.ListAllHandler)
			}
		}
	}

	// Handle OPTIONS request
	r.OPTIONS("/*cors", corsHandler.OptionsHandler, mwCors)

	r.NoRoute(notFoundHandler.NotFoundHandler)

	return r
}

func update(group *gin.RouterGroup, path string, handlers ...gin.HandlerFunc) {
	group.PATCH(path, handlers...)
	group.PUT(path, handlers...)
}
