# CASBIN-MINIO-ADAPTER Example

```go
import (
    minioadapter "github.com/Soluto/casbin-minio-adapter"
    "github.com/casbin/casbin"
)

func main() {
    adapter, _ := minioadapter.NewAdapter("http://minio-endpoint", "accessKey", "secretKey", false, "casbin-bucker", "policy.csv")

    enforcer := casbin.NewSyncedEnforcer("rbac_model.conf", adapter)

    enforcer.EnableEnforce(true)
}
```