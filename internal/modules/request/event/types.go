package event

import (
	"github.com/Confialink/wallet-accounts/internal/modules/request/model"
	"github.com/Confialink/wallet-accounts/internal/modules/transaction/types"
	"github.com/jinzhu/gorm"
)

const (
	RequestPendingApproval  = "request:request-pending-approval"
	RequestExecuted         = "request:request-executed"
	PendingRequestCancelled = "request:pending-request-cancelled"
	RequestModified         = "request:request-modified"
)

type ContextRequestPending struct {
	Tx      *gorm.DB
	Request *model.Request
	Details types.Details
}

type ContextRequestExecuted struct {
	Tx      *gorm.DB
	Request *model.Request
	Details types.Details
}

type ContextRequestModified struct {
	Tx      *gorm.DB
	Request *model.Request
	Details types.Details
}

type ContextPendingRequestCancelled struct {
	Tx        *gorm.DB
	UserID    string
	RequestID uint64
	Reason    string
}
