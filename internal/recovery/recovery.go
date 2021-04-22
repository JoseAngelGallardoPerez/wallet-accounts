package recovery

import (
	"os"

	"github.com/Confialink/wallet-pkg-utils/recovery"
)

func DefaultRecoverer() func() {
	return recovery.RecoveryWithWriter(os.Stdout)
}
