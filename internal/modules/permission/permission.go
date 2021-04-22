package permission

type Permission string

const (
	ExecuteCancelPendingTransferRequests = Permission("execute_cancel_pending_transfer_requests")
	ImportTransferRequests               = Permission("import_transfer_request_updates")
	ManualDebitCreditAccounts            = Permission("manual_debit_credit_accounts")
	InitiateExecuteUserTransfers         = Permission("initiate_execute_user_transfers")
	GenerateSendNewTans                  = Permission("generate_send_new_tans")
	ViewSettings                         = Permission("view_settings")
	ModifySettings                       = Permission("modify_settings")
	CreateSettings                       = Permission("create_settings")
	RemoveSettings                       = Permission("remove_settings")
	CreateModifyIwtBankAccounts          = Permission("create_modify_iwt_bank_accounts")
	ManageRevenue                        = Permission("manage_revenue")
	ViewRevenue                          = Permission("view_revenue")
	CreateModifyAccountTypes             = Permission("modify_account_types")
	CreateAccounts                       = Permission("create_accounts")
	ModifyAccounts                       = Permission("modify_accounts")
	CreateAccountsWithInitialBalance     = Permission("create_accounts_with_initial_balance")
	ViewAccounts                         = Permission("view_accounts")
	ViewCards                            = Permission("view_cards")
	CreateCards                          = Permission("create_cards")
	ModifyCards                          = Permission("modify_cards")
	ViewUserProfiles                     = Permission("view_user_profiles")
	ViewUserReports                      = Permission("view_user_reports")
)
