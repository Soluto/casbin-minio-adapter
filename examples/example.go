package example

import (
	minioadapter "github.com/Soluto/casbin-minio-adapter"
	"github.com/casbin/casbin"
)

func main() {
	adapter, _ := minioadapter.NewAdapter("http://minio-endpoint", "accessKey", "secretKey", false, "casbin-bucker", "policy.csv")

	enforcerer := casbin.NewSyncedEnforcer("rbac_model.conf", adapter)

	enforcerer.EnableEnforce(true)
}
