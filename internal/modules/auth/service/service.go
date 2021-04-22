package service

import (
	accountPoliciy "github.com/Confialink/wallet-accounts/internal/modules/account/policy"
	permissionPolicy "github.com/Confialink/wallet-accounts/internal/modules/permission/policy"
	"github.com/Confialink/wallet-accounts/internal/modules/policy"
	userpb "github.com/Confialink/wallet-users/rpc/proto/users"
	goAcl "github.com/kildevaeld/go-acl"
)

const (
	CardTypeResource = "card_type"
	CardResource     = "card"

	AccountsResource     = "private_accounts"
	UserAccountsResource = "private_user_accounts"
	AccountTypesResource = "private_account_types"
	UserAccountCsvReport = "user_account_csv_report"

	ResourceTransferRequest = "transfer_request"
	ResourceTransaction     = "transaction"
	ResourceTBARequest      = "tba_request"
	ResourceTBURequest      = "tbu_request"
	ResourceSetting         = "setting"

	PaymentMethodsResource = "private_payment_methods"
	PaymentPeriodsResource = "private_payment_periods"

	ResourcePermission            = "permission"
	ResourceIwtBankAccount        = "iwt_bank_account"
	ResourceRequest               = "request"
	ResourceRevenueAccount        = "revenue_account"
	ResourceCountry               = "country"
	ResourceScheduledTransactions = "scheduled_transactions"
)

const (
	ActionCreate      = "create"
	ActionUpdate      = "update"
	ActionRead        = "read"
	ActionReadList    = "read_list"
	ActionReadOwnList = "read_own_list"
	ActionDelete      = "delete"
	ActionHas         = "has"
)

const (
	RoleRoot   = "root"
	RoleAdmin  = "admin"
	RoleClient = "client"
)

type RequiredPolicies struct {
	ViewTransferRequest        policy.Policy
	ViewTransaction            policy.Policy
	CreateModifyIwtBankAccount policy.Policy
	RevenueManager             policy.Policy
	RevenueViewer              policy.Policy
	ViewAccount                policy.Policy
	ViewCard                   policy.Policy
	ViewSettings               policy.Policy
}

type PermissionMap map[string]map[string]map[string]policy.Policy

type AuthServiceInterface interface {
	Can(role string, action string, resource string) bool
	CanDynamic(user *userpb.User, action string, resourceName string, resource interface{}) bool
}

type AuthService struct {
	Acl                *goAcl.ACL
	DynamicPermissions PermissionMap
	Policies           *RequiredPolicies
}

func NewService(acl *goAcl.ACL, policies *RequiredPolicies) AuthServiceInterface {
	auth := AuthService{Acl: acl, Policies: policies}
	auth.registerPermissions()
	auth.DynamicPermissions = PermissionMap{
		RoleClient: {
			CardResource: {ActionRead: canClientReadCard, ActionReadOwnList: allowFunc},
			AccountsResource: {
				ActionRead:     accountPoliciy.CanClientRead,
				ActionReadList: accountPoliciy.CanClientReadList,
			},
			UserAccountsResource: {
				ActionReadList: allowFunc,
			},
			UserAccountCsvReport: {
				ActionRead: accountPoliciy.CanClientRead,
			},
			ResourceTransferRequest: {
				ActionRead: policies.ViewTransferRequest,
			},
			ResourceTransaction: {
				ActionRead:     policies.ViewTransaction,
				ActionReadList: allowFunc,
			},
			ResourceCountry: {
				ActionRead:     allowFunc,
				ActionReadList: allowFunc,
			},
			ResourceIwtBankAccount: {
				ActionReadList: allowFunc,
				ActionRead:     allowFunc,
			},
			ResourceSetting: {
				ActionRead: canClientReadSetting,
			},
		},
		RoleAdmin: {
			CardResource: {
				ActionRead:   policies.ViewCard,
				ActionUpdate: permissionPolicy.CheckPermission,
			},
			AccountsResource: {
				ActionRead:     policies.ViewAccount,
				ActionReadList: policies.ViewAccount,
				ActionCreate:   permissionPolicy.CheckPermission,
				ActionUpdate:   permissionPolicy.CheckPermission,
			},
			AccountTypesResource: {
				ActionRead:     allowFunc,
				ActionReadList: allowFunc,
				ActionCreate:   allowFunc,
				ActionUpdate:   allowFunc,
				ActionDelete:   allowFunc,
			},
			PaymentMethodsResource: {
				ActionReadList: allowFunc,
			},
			PaymentPeriodsResource: {
				ActionReadList: allowFunc,
			},
			ResourcePermission: {
				ActionHas: permissionPolicy.CheckPermission,
			},
			ResourceTransferRequest: {
				ActionRead: allowFunc,
			},
			ResourceTransaction: {
				ActionRead: allowFunc,
			},
			ResourceIwtBankAccount: {
				ActionRead:     allowFunc,
				ActionReadList: allowFunc,
				ActionCreate:   policies.CreateModifyIwtBankAccount,
				ActionUpdate:   policies.CreateModifyIwtBankAccount,
				ActionDelete:   policies.CreateModifyIwtBankAccount,
			},
			ResourceRequest: {
				ActionReadList: allowFunc,
			},
			ResourceRevenueAccount: {
				ActionRead:     policies.RevenueViewer,
				ActionReadList: policies.RevenueViewer,
				ActionUpdate:   policies.RevenueManager,
			},
			ResourceCountry: {
				ActionRead:     allowFunc,
				ActionReadList: allowFunc,
			},
			ResourceSetting: {
				ActionRead: policies.ViewSettings,
			},
		},
	}

	return &auth
}

// getRolesResources returns permissions data
func (auth AuthService) getRolesResources() map[string]map[string][]string {
	return map[string]map[string][]string{
		RoleClient: {
			AccountsResource:     {ActionRead},
			AccountTypesResource: {ActionRead},
		},
		RoleAdmin: {
			AccountsResource:     {ActionCreate, ActionUpdate, ActionRead, ActionDelete},
			AccountTypesResource: {ActionCreate, ActionUpdate, ActionRead, ActionDelete},
			CardTypeResource:     {ActionRead},
			CardResource:         {ActionCreate, ActionReadList},
		},
		RoleRoot: {
			AccountsResource:     {ActionCreate, ActionUpdate, ActionRead, ActionDelete},
			AccountTypesResource: {ActionCreate, ActionUpdate, ActionRead, ActionDelete},
			CardTypeResource:     {ActionCreate, ActionUpdate, ActionRead, ActionDelete},
			CardResource:         {ActionCreate, ActionReadList},
		},
	}
}

// Can checks action is allowed
func (auth AuthService) Can(role string, action string, resource string) bool {
	if role == RoleRoot {
		return true
	}
	return auth.Acl.Can(role, action, resource)
}

// CanDynamic checks action is allowed by calling associated function
func (auth *AuthService) CanDynamic(user *userpb.User, action string, resourceName string, resource interface{}) bool {
	if user.RoleName == RoleRoot {
		return true
	}

	function := auth.getPermissionFunc(user.RoleName, action, resourceName)
	return function(resource, user)
}

// allowFunc always allows access
func allowFunc(_ interface{}, _ *userpb.User) bool {
	return true
}

// blockFunc always block access
func blockFunc(_ interface{}, _ *userpb.User) bool {
	return false
}

// getPermissionFunc returns function by role, action and resourceName.
// Returns blockFunc if proposed func not found
func (auth *AuthService) getPermissionFunc(role string, action string, resourceName string) policy.Policy {
	if rolePermission, ok := auth.DynamicPermissions[role]; ok {
		if resourcePermission, ok := rolePermission[resourceName]; ok {
			if actionPermission, ok := resourcePermission[action]; ok {
				return actionPermission
			}
		}
	}
	return blockFunc
}

// registerPermissions registers allowed actions for roles
func (auth AuthService) registerPermissions() {
	for role, resources := range auth.getRolesResources() {
		auth.Acl.Role(role, "")
		for resource, actions := range resources {
			for _, action := range actions {
				auth.Acl.Allow(role, action, resource)
			}
		}
	}
}
