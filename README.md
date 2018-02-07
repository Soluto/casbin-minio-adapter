# casbin-minio-adapter

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/Soluto/casbin-minio-adapter)](https://goreportcard.com/report/github.com/Soluto/casbin-minio-adapter)
[![Build Status](https://travis-ci.org/Soluto/casbin-minio-adapter.svg?branch=master)](https://travis-ci.org/Soluto/casbin-minio-adapter)
[![Build Status](https://ci.appveyor.com/api/projects/status/github/Soluto/casbin-minio-adapter?branch=master&svg=true)](https://ci.appveyor.com/project/Soluto/casbin-minio-adapter)
[![Coverage Status](https://coveralls.io/repos/github/Soluto/casbin-minio-adapter/badge.svg?branch=master)](https://coveralls.io/github/Soluto/casbin-minio-adapter?branch=master)
[![Godoc](https://godoc.org/github.com/Soluto/casbin-minio-adapter?status.svg)](https://godoc.org/github.com/Soluto/casbin-minio-adapter)

[Casbin](https://github.com/casbin/casbin) adapter implementation using [Minio](https://github.com/minio/minio)/AWS S3 policy storage

## Installation

    go get github.com/Soluto/casbin-minio-adapter

## Usage

```go
import (
    minioadapter "github.com/Soluto/casbin-minio-adapter"
    "github.com/casbin/casbin"
)

func main() {

    adapter, _ := minioadapter.NewAdapter("http://minio-endpoint", "accessKey", "secretKey", false, "casbin-bucker", "policy.csv")

    enforcer := casbin.NewSyncedEnforcer("rbac_model.conf", adapter)

}
```

## Related pojects

- [Casbin](https://github.com/casbin/casbin)
- [Minio](https://github.com/minio/minio)

## Additional Usage Examples

For real-world example visit [Tweek](https://github.com/Soluto/tweek).

## License

This project is under MIT License. See the [LICENSE](LICENSE) file for the full license text.