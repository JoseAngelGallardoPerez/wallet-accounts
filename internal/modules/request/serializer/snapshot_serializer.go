package serializer

import (
	"github.com/Confialink/wallet-pkg-acl"
	"github.com/Confialink/wallet-pkg-model_serializer"
	"github.com/inconshreveable/log15"

	requestModel "github.com/Confialink/wallet-accounts/internal/modules/request/model"
	"github.com/Confialink/wallet-accounts/internal/modules/request/policy"
	"github.com/Confialink/wallet-users/rpc/proto/users"
)

const snapshotsFieldName = "snapshots"

var clientViewSnapshotPolicy = policy.ProvideClientViewBalanceSnapshot()

func ProvideSnapshotsSerializer(user *users.User, logger log15.Logger) model_serializer.FieldSerializer {
	return func(model interface{}) (fieldName string, value interface{}) {
		request := model.(*requestModel.Request)
		if request.BalanceSnapshots == nil {
			return snapshotsFieldName, nil
		}
		logger = logger.New("where", "SnapshotsSerializer")

		snapshots := request.BalanceSnapshots
		result := make([]map[string]interface{}, 0, len(snapshots))
		role := acl.RolesHelper.FromName(user.RoleName)

		for _, snapshot := range snapshots {
			if role < acl.Admin && !clientViewSnapshotPolicy(snapshot, user) {
				continue
			}
			value, err := snapshot.GetValue()
			if err != nil {
				logger.Warn("unable to get snapshot value", "error", err)
				continue
			}

			item := map[string]interface{}{
				"value":     value,
				"balanceId": snapshot.BalanceId,
			}
			if snapshot.BalanceType != nil {
				item["balanceType"] = snapshot.BalanceType
			}
			result = append(result, item)
		}

		return snapshotsFieldName, result
	}
}
