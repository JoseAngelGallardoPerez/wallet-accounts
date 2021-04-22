package handler

import (
	"github.com/Confialink/wallet-accounts/internal/modules/balance/service"
	"github.com/inconshreveable/log15"
)

var (
	snapshotService *service.Snapshot
	logger          log15.Logger
)

func LoadDependencies(snapshotServiceDep *service.Snapshot, loggerDep log15.Logger) {
	snapshotService = snapshotServiceDep
	logger = loggerDep.New("module", "balance")
}
