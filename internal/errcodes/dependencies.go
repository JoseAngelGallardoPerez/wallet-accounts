package errcodes

import "github.com/inconshreveable/log15"

var logger log15.Logger

func LoadDependencies(loggerDep log15.Logger) {
	logger = loggerDep
}
