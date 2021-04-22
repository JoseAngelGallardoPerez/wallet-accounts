package permission

import (
	"github.com/Confialink/wallet-accounts/internal/srvdiscovery"
	"context"
	"net/http"

	"github.com/Confialink/wallet-permissions/rpc/permissions"
)

type Service struct {
}

func NewPermissionService() *Service {
	return &Service{}
}

//Check checks if specified user is granted permission to perform some action
func (p *Service) Check(userId, actionKey string) (bool, error) {
	request := &permissions.PermissionReq{UserId: userId, ActionKey: actionKey}

	checker, err := p.checker()
	if nil != err {
		return false, err
	}

	response, err := checker.Check(context.Background(), request)
	if nil != err {
		return false, err
	}
	return response.IsAllowed, nil
}

func (p *Service) checker() (permissions.PermissionChecker, error) {
	permissionsUrl, err := srvdiscovery.ResolveRPC(srvdiscovery.ServiceNamePermissions)
	if nil != err {
		return nil, err
	}
	checker := permissions.NewPermissionCheckerProtobufClient(permissionsUrl.String(), http.DefaultClient)
	return checker, nil
}
