package constants

import (
	"strings"

	"github.com/Confialink/wallet-pkg-errors"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
)

type Subject string

func (s Subject) EqualsTo(subject Subject) bool {
	return strings.ToLower(string(s)) == strings.ToLower(string(subject))
}

const (
	StatusNew       = "new"
	StatusPending   = "pending"
	StatusExecuted  = "executed"
	StatusCancelled = "cancelled"
)

const (
	SubjectTransferBetweenAccounts      = Subject("TBA")
	SubjectTransferBetweenUsers         = Subject("TBU")
	SubjectTransferOutgoingWireTransfer = Subject("OWT")
	SubjectCardFundingTransfer          = Subject("CFT")
	SubjectCreditAccount                = Subject("CA")
	SubjectDebitAccount                 = Subject("DA")
	SubjectTransferIncomingWireTransfer = Subject("IWT")
	SubjectDebitRevenueAccount          = Subject("DRA")
)

var knownSubjects = map[string]Subject{
	string(SubjectTransferBetweenAccounts):      SubjectTransferBetweenAccounts,
	string(SubjectTransferBetweenUsers):         SubjectTransferBetweenUsers,
	string(SubjectTransferOutgoingWireTransfer): SubjectTransferOutgoingWireTransfer,
	string(SubjectCardFundingTransfer):          SubjectCardFundingTransfer,
	string(SubjectCreditAccount):                SubjectCreditAccount,
	string(SubjectDebitAccount):                 SubjectDebitAccount,
	string(SubjectTransferIncomingWireTransfer): SubjectTransferIncomingWireTransfer,
	string(SubjectDebitRevenueAccount):          SubjectDebitRevenueAccount,
}

func (s Subject) String() string {
	return string(s)
}

func SubjectFromString(subject string) (Subject, *errors.PublicError) {
	if result, ok := knownSubjects[subject]; ok {
		return result, nil
	}
	return "", errcodes.CreatePublicError(errcodes.CodeUnknownRequestSubject)
}
