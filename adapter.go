package minioadapter

import (
	"bufio"
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
	client   *minio.Client
	bucket   string
	filename string
}

// LoadPolicy loads all policy rules from the storage.
func (a *MinioAdapter) LoadPolicy(model model.Model) error {
	obj, err := a.client.GetObject(a.bucket, a.filename, minio.GetObjectOptions{})
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
	pr, pw := io.Pipe()
	var streamLength int64

	for ptype, ast := range model["p"] {
		for _, rule := range ast.Policy {
			l, err := pw.Write([]byte(fmt.Sprintf("%s,%s/n", ptype, util.ArrayToString(rule))))
			if err != nil {
				return err
			}
			streamLength += int64(l)
		}
	}

	for ptype, ast := range model["g"] {
		for _, rule := range ast.Policy {
			l, err := pw.Write([]byte(fmt.Sprintf("%s,%s/n", ptype, util.ArrayToString(rule))))
			if err != nil {
				return err
			}
			streamLength += int64(l)
		}
	}

	a.client.PutObject(a.bucket, a.filename, pr, streamLength, minio.PutObjectOptions{})

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
func NewAdapter(endpoint string, accessKeyID string, secretAccessKey string, secure bool, bucket string, filename string) (persist.Adapter, error) {
	client, err := minio.New(endpoint, accessKeyID, secretAccessKey, secure)
	if err != nil {
		return nil, err
	}

	ma := &MinioAdapter{
		client:   client,
		bucket:   bucket,
		filename: filename,
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
