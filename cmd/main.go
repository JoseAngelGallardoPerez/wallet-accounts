package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Confialink/wallet-pkg-env_mods"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/inconshreveable/log15"
	"github.com/jinzhu/gorm"
	"github.com/kildevaeld/go-acl"
	"github.com/olebedev/emitter"
	"github.com/shopspring/decimal"
	"go.uber.org/dig"

	rpcAccounts "github.com/Confialink/wallet-accounts/rpc/accounts"
	rpcLimit "github.com/Confialink/wallet-accounts/rpc/limit"

	"github.com/Confialink/wallet-accounts/internal/accounts/protoserv"
	"github.com/Confialink/wallet-accounts/internal/commands"
	"github.com/Confialink/wallet-accounts/internal/config"
	"github.com/Confialink/wallet-accounts/internal/config/initializers"
	"github.com/Confialink/wallet-accounts/internal/database"
	"github.com/Confialink/wallet-accounts/internal/errcodes"
	"github.com/Confialink/wallet-accounts/internal/event"
	"github.com/Confialink/wallet-accounts/internal/limit"
	"github.com/Confialink/wallet-accounts/internal/limitserver"
	accountTypeProvider "github.com/Confialink/wallet-accounts/internal/modules/account-type/account-type-provider"
	accountProvider "github.com/Confialink/wallet-accounts/internal/modules/account/account-provider"
	accountSubscriber "github.com/Confialink/wallet-accounts/internal/modules/account/subscriber"
	"github.com/Confialink/wallet-accounts/internal/modules/app"
	"github.com/Confialink/wallet-accounts/internal/modules/app/validator"
	authModule "github.com/Confialink/wallet-accounts/internal/modules/auth"
	balanceProvider "github.com/Confialink/wallet-accounts/internal/modules/balance/balance-provider"
	balanceSubscriber "github.com/Confialink/wallet-accounts/internal/modules/balance/subscriber"
	balanceSubscriptionHandler "github.com/Confialink/wallet-accounts/internal/modules/balance/subscriber/handler"
	bankDetailsProvider "github.com/Confialink/wallet-accounts/internal/modules/bank-details/bank-details-provider"
	"github.com/Confialink/wallet-accounts/internal/modules/calculation"
	cardType "github.com/Confialink/wallet-accounts/internal/modules/card-type"
	cardTypeCategory "github.com/Confialink/wallet-accounts/internal/modules/card-type-category"
	cardTypeFormat "github.com/Confialink/wallet-accounts/internal/modules/card-type-format"
	cardProvider "github.com/Confialink/wallet-accounts/internal/modules/card/card-provider"
	commonProvider "github.com/Confialink/wallet-accounts/internal/modules/common/common-provider"
	"github.com/Confialink/wallet-accounts/internal/modules/country"
	currencyProvider "github.com/Confialink/wallet-accounts/internal/modules/currency/currency-provider"
	feeProvider "github.com/Confialink/wallet-accounts/internal/modules/fee/fee-provider"
	moneyRequest "github.com/Confialink/wallet-accounts/internal/modules/moneyrequest/provider"
	notificationsProvider "github.com/Confialink/wallet-accounts/internal/modules/notifications/notifications-provider"
	notificationsSubscriber "github.com/Confialink/wallet-accounts/internal/modules/notifications/subscriber"
	notificationsSubscriberHandler "github.com/Confialink/wallet-accounts/internal/modules/notifications/subscriber/handler"
	paymentMethodProvider "github.com/Confialink/wallet-accounts/internal/modules/payment-method/payment-method-provider"
	paymentPeriodProvider "github.com/Confialink/wallet-accounts/internal/modules/payment-period/payment-period-provider"
	permissionProvider "github.com/Confialink/wallet-accounts/internal/modules/permission/permission-provider"
	"github.com/Confialink/wallet-accounts/internal/modules/request"
	requestProvider "github.com/Confialink/wallet-accounts/internal/modules/request/request-provider"
	scheduledTransaction "github.com/Confialink/wallet-accounts/internal/modules/scheduled-transaction"
	stp "github.com/Confialink/wallet-accounts/internal/modules/scheduled-transaction/scheduled-transaction-provider"
	scheduledTransactionSubscriber "github.com/Confialink/wallet-accounts/internal/modules/scheduled-transaction/subscriber"
	settingsProvider "github.com/Confialink/wallet-accounts/internal/modules/settings/settings-provider"
	systemLogsProvider "github.com/Confialink/wallet-accounts/internal/modules/system-logs/system-logs-provider"
	tanProvider "github.com/Confialink/wallet-accounts/internal/modules/tan/tan-provider"
	transactionProvider "github.com/Confialink/wallet-accounts/internal/modules/transaction/transaction-provider"
	transactionView "github.com/Confialink/wallet-accounts/internal/modules/transaction/transaction-view"
	userProvider "github.com/Confialink/wallet-accounts/internal/modules/user/user-provider"
	"github.com/Confialink/wallet-accounts/internal/routes"
)

// main: main function
func main() {
	c := initContainer()

	decimal.DivisionPrecision = 32
	loadDependencies(c)
	err := c.Invoke(func(
		appConfig *config.Config,
		v validator.Interface,
		eventEmitter *emitter.Emitter,
	) {
		ginMode := env_mods.GetMode(appConfig.Env)
		gin.SetMode(ginMode)
		initializers.Initialize(v)
		initializers.Initialize(binding.Validator.Engine().(validator.Interface))

		cmd := flag.String("cmd", "", "Execute command. Run -cmd help to see a list of available commands")
		flag.Parse()
		if *cmd != "" {
			commands.Process(*cmd, c)
			os.Exit(0)
		}
	})

	if err != nil {
		log.Fatal(err)
	}
	subscribeModules(c)

	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	readTimeout, _ := time.ParseDuration("30s")
	writTimeout, _ := time.ParseDuration("30s")
	idleTimeout, _ := time.ParseDuration("60s")

	err = c.Invoke(func(
		appConfig *config.Config,
		accountsMixedServer *protoserv.ProtobufServer,
		limitServer *limitserver.Server,
		apiRouter *gin.Engine,
	) {
		log.Printf("Starting RPC server on port %s\n", appConfig.ProtobufPort)
		mixedAccountHandler := rpcAccounts.NewAccountsProcessorServer(accountsMixedServer, nil)
		limitHandler := rpcLimit.NewLimitsServer(limitServer, nil)

		rpcMux := http.NewServeMux()
		rpcMux.Handle(rpcAccounts.AccountsProcessorPathPrefix, mixedAccountHandler)
		rpcMux.Handle(rpcLimit.LimitsPathPrefix, limitHandler)
		rpcServer := &http.Server{
			Addr:         fmt.Sprintf(":%s", appConfig.ProtobufPort),
			Handler:      rpcMux,
			ReadTimeout:  readTimeout,
			WriteTimeout: writTimeout,
			IdleTimeout:  idleTimeout,
		}

		wg.Add(1)
		go serve("accounts RPC", wg, ctx, rpcServer)

		log.Printf("Starting API server on port %s\n", appConfig.Port)
		apiServer := &http.Server{
			Addr:         fmt.Sprintf(":%s", appConfig.Port),
			Handler:      apiRouter,
			ReadTimeout:  readTimeout,
			WriteTimeout: writTimeout,
			IdleTimeout:  idleTimeout,
		}

		wg.Add(1)
		go serve("accounts API", wg, ctx, apiServer)
	})

	if err != nil {
		log.Fatal(err)
	}

	err = c.Invoke(runScheduledJobs)

	if err != nil {
		log.Fatal(err)
	}

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down accounts service...")
	// Notify servers about termination process
	cancel()
	// Wait for all servers to gracefully shutdown
	wg.Wait()
}

func serve(name string, wg *sync.WaitGroup, ctx context.Context, srv *http.Server) {
	var err error
	defer wg.Done()

	go func() {
		if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("%s listen:%s\n", name, err.Error())
		}
	}()

	log.Printf("%s server started\n", name)

	<-ctx.Done()

	log.Printf("%s server stopped\n", name)

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	if err = srv.Shutdown(ctxShutDown); err != nil {
		log.Fatalf("%s server Shutdown Failed:%s", name, err)
	}

	log.Printf("%s server exited properly\n", name)
}

func runScheduledJobs(
	db *gorm.DB,
	scheduledTxRepo *scheduledTransaction.Repository,
	scheduledTxService *scheduledTransaction.Service,
	requestCreator *request.Creator,
	logger log15.Logger,
) {
	scheduleTransactionsCron, err := scheduledTransaction.Schedule(
		scheduledTxRepo,
		requestCreator,
		db,
		scheduledTxService,
		logger.New("module", "scheduled_transaction"),
	)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Starting scheduled transactions jobs")
	scheduleTransactionsCron.Start()
}

func subscribeModules(c *dig.Container) {
	consumers := []interface{}{
		accountSubscriber.Subscribe,
		scheduledTransactionSubscriber.Subscribe,
		notificationsSubscriber.Subscribe,
		balanceSubscriber.Subscribe,
	}
	for _, consumer := range consumers {
		err := c.Invoke(consumer)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func loadDependencies(c *dig.Container) {
	loaders := []interface{}{
		initializers.LoadDependencies,
		notificationsSubscriberHandler.LoadDependencies,
		transactionView.LoadDependencies,
		balanceSubscriptionHandler.LoadDependencies,
		errcodes.LoadDependencies,
	}

	for _, loader := range loaders {
		err := c.Invoke(loader)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// initContainer creates and initialize di container.
func initContainer() *dig.Container {
	container := dig.New()

	providers := make([]interface{}, 0)

	providers = append(providers, mainProviders()...)
	providers = append(providers, accountProvider.Providers()...)
	providers = append(providers, accountTypeProvider.Providers()...)
	providers = append(providers, app.Providers()...)
	providers = append(providers, authModule.Providers()...)
	providers = append(providers, balanceProvider.Providers()...)
	providers = append(providers, bankDetailsProvider.Providers()...)
	providers = append(providers, config.Providers()...)
	providers = append(providers, country.Providers()...)
	providers = append(providers, cardProvider.Providers()...)
	providers = append(providers, cardType.Providers()...)
	providers = append(providers, cardTypeCategory.Providers()...)
	providers = append(providers, cardTypeFormat.Providers()...)
	providers = append(providers, currencyProvider.Providers()...)
	providers = append(providers, database.Providers()...)
	providers = append(providers, feeProvider.Providers()...)
	providers = append(providers, notificationsProvider.Providers()...)
	providers = append(providers, paymentMethodProvider.Providers()...)
	providers = append(providers, paymentPeriodProvider.Providers()...)
	providers = append(providers, permissionProvider.Providers()...)
	providers = append(providers, requestProvider.Providers()...)
	providers = append(providers, stp.Providers()...)
	providers = append(providers, settingsProvider.Providers()...)
	providers = append(providers, systemLogsProvider.Providers()...)
	providers = append(providers, tanProvider.Providers()...)
	providers = append(providers, transactionProvider.Providers()...)
	providers = append(providers, userProvider.Providers()...)
	providers = append(providers, commonProvider.Providers()...)
	providers = append(providers, calculation.Providers()...)
	providers = append(providers, limit.Providers()...)
	providers = append(providers, moneyRequest.Providers()...)
	// API router
	providers = append(providers, routes.NewAPIRouter)
	// rpc servers
	providers = append(
		providers,
		protoserv.NewProtobufServer,
		limitserver.NewServer,
	)

	for _, provider := range providers {
		err := container.Provide(provider)
		if err != nil {
			panic("unable to init container " + err.Error())
		}
	}

	return container
}

func mainProviders() []interface{} {
	return []interface{}{
		/*
		 |--------------------------------------------------------------------------
		 | Service logger
		 |--------------------------------------------------------------------------
		 | Default service logger
		*/
		func() log15.Logger {
			return log15.New("service", "accounts")
		},
		/*
		 |--------------------------------------------------------------------------
		 | ACL
		 |--------------------------------------------------------------------------
		 | This package is used in order to simplify work with internal permissions
		*/
		func() *acl.ACL {
			return acl.New(acl.NewMemoryStore())
		},
		/*
			|--------------------------------------------------------------------------
			| Global event emitter
			|--------------------------------------------------------------------------
			| Used in order to provide a way to different modules
			| to communicate via events
		*/
		event.Emitter,
	}
}
