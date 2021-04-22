package constants

// transaction purpose must uniquely identify transaction
type Purpose string

const (
	PurposeTBAOutgoing   = Purpose("tba_outgoing")
	PurposeTBAIncoming   = Purpose("tba_incoming")
	PurposeTBUOutgoing   = Purpose("tbu_outgoing")
	PurposeTBUIncoming   = Purpose("tbu_incoming")
	PurposeOWTOutgoing   = Purpose("owt_outgoing")
	PurposeCFTOutgoing   = Purpose("cft_outgoing")
	PurposeCFTIncoming   = Purpose("cft_incoming")
	PurposeCreditAccount = Purpose("credit_account")
	PurposeDebitRevenue  = Purpose("debit_revenue")
	PurposeDebitAccount  = Purpose("debit_account")
	PurposeCreditRevenue = Purpose("credit_revenue")

	PurposeFeeExchangeMargin = Purpose("fee_exchange_margin")
	PurposeFeeTransfer       = Purpose("fee_default_transfer")
	PurposeFeeIWT            = Purpose("fee_iwt")

	PurposeRevenueExchangeMargin = Purpose("revenue_exchange_margin")
	PurposeRevenueIwt            = Purpose("revenue_iwt_transfer_fee")
)

// MainTransactions is a slice of Purposes that are main in context of request (All transactions excepts fee, revenue, etc.)
var MainTransactions = []Purpose{PurposeTBAOutgoing, PurposeTBAIncoming,
	PurposeTBUOutgoing, PurposeTBUIncoming, PurposeOWTOutgoing,
	PurposeCFTOutgoing, PurposeCFTIncoming, PurposeCreditAccount,
	PurposeDebitRevenue, PurposeDebitAccount, PurposeCreditRevenue}

func (p Purpose) String() string {
	return string(p)
}
