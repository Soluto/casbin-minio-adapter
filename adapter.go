package minioadapter

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/casbin/casbin/model"
	"github.com/casbin/casbin/persist"
	"github.com/casbin/casbin/util"
	minio "github.com/minio/minio-go"
)

// MinioAdapter the struct that implements
type MinioAdapter struct {
	client     *minio.Client
	bucket     string
	objectName string
}

// LoadPolicy loads all policy rules from the storage.
func (a *MinioAdapter) LoadPolicy(model model.Model) error {
	obj, err := a.client.GetObject(a.bucket, a.objectName, minio.GetObjectOptions{})
	if err != nil {
		return err
	}
	buf := bufio.NewReader(obj)
	for {
		line, err := buf.ReadString('\n')
		line = strings.TrimSpace(line)
		persist.LoadPolicyLine(line, model)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}

// SavePolicy saves all policy rules to the storage.
func (a *MinioAdapter) SavePolicy(model model.Model) error {
	var tmp bytes.Buffer
	var streamLength int64

	for ptype, ast := range model["p"] {
		for _, rule := range ast.Policy {
			l, err := tmp.WriteString(fmt.Sprintf("%s, %s\n", ptype, util.ArrayToString(rule)))
			if err != nil {
				return err
			}
			streamLength += int64(l)
		}
	}

	for ptype, ast := range model["g"] {
		for _, rule := range ast.Policy {
			l, err := tmp.WriteString(fmt.Sprintf("%s, %s\n", ptype, util.ArrayToString(rule)))
			if err != nil {
				return err
			}
			streamLength += int64(l)
		}
	}

	a.client.PutObject(a.bucket, a.objectName, &tmp, streamLength, minio.PutObjectOptions{})

	return nil
}

// AddPolicy adds a policy rule to the storage.
// This is part of the Auto-Save feature.
func (a *MinioAdapter) AddPolicy(sec string, ptype string, rule []string) error {
	return errors.New("Not implemented")
}

// RemovePolicy removes a policy rule from the storage.
// This is part of the Auto-Save feature.
func (a *MinioAdapter) RemovePolicy(sec string, ptype string, rule []string) error {
	return errors.New("Not implemented")
}

// RemoveFilteredPolicy removes policy rules that match the filter from the storage.
// This is part of the Auto-Save feature.
func (a *MinioAdapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	return errors.New("Not implemented")
}

// NewAdapter create new MinioAdapter
// Parameters:
// 	- endpoint
//		URL to object storage service.
// 	- accessKey
//		Access key is the user ID that uniquely identifies your account.
//	- secretKey
//		Secret key is the password to your account.
//  - secure
//		Set this value to 'true' to enable secure (HTTPS) access.
//  - bucket
//		Name of the bucket where the policy is stored
//  - objectName
//		Name of the object that contains policy
func NewAdapter(endpoint string, accessKey string, secretKey string, secure bool, bucket string, objectName string) (persist.Adapter, error) {
	client, err := minio.New(endpoint, accessKey, secretKey, secure)
	if err != nil {
		return nil, err
	}

	ma := &MinioAdapter{
		client:     client,
		bucket:     bucket,
		objectName: objectName,
	}

	ok, err := ma.client.BucketExists(ma.bucket)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("MinioAdapterError: bucket %s doesn't exist", ma.bucket)
	}

	return ma, nil
}
