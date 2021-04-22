package system_logs

import (
	"context"

	"github.com/Confialink/wallet-accounts/internal/modules/system-logs/connection"
	pb "github.com/Confialink/wallet-logs/rpc/logs"
	"github.com/inconshreveable/log15"
)

type logsServiceWrap struct {
	logsService pb.LogsService
	logger      log15.Logger
}

func newLogsServiceWrap(logger log15.Logger) *logsServiceWrap {
	return &logsServiceWrap{logger: logger}
}

func (self *logsServiceWrap) createLog(
	subject string,
	userId string,
	logTime string,
	dataTitle string,
	data []byte,
) {
	resp, err := self.systemLogger().CreateLog(context.Background(), &pb.CreateLogReq{
		Subject:    subject,
		UserId:     userId,
		LogTime:    logTime,
		DataTitle:  dataTitle,
		DataFields: data,
	})
	if err != nil {
		self.logger.Error("Failed to create system log", "error", err)
		return
	}
	if resp.Error != nil {
		self.logger.Error("Failed to create system log on logs service", "Error title", resp.Error.Title, "Error details", resp.Error.Details)
	}
}

func (self *logsServiceWrap) systemLogger() pb.LogsService {
	if self.logsService == nil {
		var err error
		self.logsService, err = connection.GetSystemLogsClient()
		if err != nil {
			self.logger.Error("Failed to connect to logger service", "error", err)
			return nil
		}
	}
	return self.logsService
}
