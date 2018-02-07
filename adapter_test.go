package minioadapter

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/casbin/casbin/file-adapter"

	"github.com/casbin/casbin"
	"github.com/casbin/casbin/model"
	dc "github.com/fsouza/go-dockerclient"
	minio "github.com/minio/minio-go"
	dockertest "gopkg.in/ory-am/dockertest.v3"
)

var resource *dockertest.Resource
var client *minio.Client

func TestMain(m *testing.M) {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	runOptions := &dockertest.RunOptions{
		Repository: "minio/minio",
		Tag:        "latest",
		Cmd:        []string{"server", "/data"},
		PortBindings: map[dc.Port][]dc.PortBinding{
			"9000": []dc.PortBinding{{HostPort: "9000"}},
		},
		Env: []string{"MINIO_ACCESS_KEY=ACCESSKEY", "MINIO_SECRET_KEY=SECRETKEY"},
	}
	// pulls an image, creates a container based on it and runs it
	resource, err = pool.RunWithOptions(runOptions)
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}
	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		var err error
		endpoint := fmt.Sprintf("localhost:%v", resource.GetPort("9000/tcp"))
		client, err = minio.New(endpoint, "ACCESSKEY", "SECRETKEY", false)
		if err != nil {
			log.Println("Failed to create minio client:", err)
			return err
		}

		exists, err := client.BucketExists("casbin-bucket")
		if err != nil {
			log.Println("Failed check bucket existence:", err)
			return err
		}

		if !exists {
			err = client.MakeBucket("casbin-bucket", "location-1")
			if err != nil {
				log.Println("Failed to create bucket:", err)
				return err
			}
		}

		return nil
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func TestLoadPolicy(t *testing.T) {
	if _, err := client.FPutObject("casbin-bucket", "policy.csv", "./examples/rbac_policy.csv", minio.PutObjectOptions{ContentType: "application/csv"}); err != nil {
		log.Fatal("Test setup: Failed to upload policy:", err)
	}

	endpoint := fmt.Sprintf("localhost:%v", resource.GetPort("9000/tcp"))
	adapter, err := NewAdapter(endpoint, "ACCESSKEY", "SECRETKEY", false, "casbin-bucket", "policy.csv")
	if err != nil {
		t.Fatal("Failed to create adapter:", err)
	}

	m := model.Model{}
	m.LoadModel("examples/rbac_model.conf")

	err = adapter.LoadPolicy(m)
	if err != nil {
		t.Fatal("Failed to load policy:", err)
	}

	if m.GetValuesForFieldInPolicy("p", "p", 0)[0] != "alice" {
		t.Fatal("Policy wasn't loaded properly")
	}
}

func TestSavePolicy(t *testing.T) {
	m := model.Model{}
	m.LoadModel("examples/rbac_model.conf")

	fi, err := os.Stat("examples/rbac_policy.csv")
	if err != nil {
		t.Fatal("Testing policy file error:", err)
	}

	fAdapter := fileadapter.NewAdapter("examples/rbac_policy.csv")
	fAdapter.LoadPolicy(m)

	endpoint := fmt.Sprintf("localhost:%v", resource.GetPort("9000/tcp"))
	adapter, err := NewAdapter(endpoint, "ACCESSKEY", "SECRETKEY", false, "casbin-bucket", "policy2.csv")
	if err != nil {
		t.Fatal("Failed to create adapter:", err)
	}

	err = adapter.SavePolicy(m)
	if err != nil {
		t.Fatal("Failed to save policy:", err)
	}

	objectInfo, err := client.StatObject("casbin-bucket", "policy2.csv", minio.StatObjectOptions{})
	if err != nil {
		t.Fatal("Failed to get storage policy info:", err)
	}
	if objectInfo.Size != fi.Size() {
		t.Fatalf("Policy file size %v not equal to stored policy size %v", fi.Size(), objectInfo.Size)
	}
}

func TestWithEnforcer(t *testing.T) {

	endpoint := fmt.Sprintf("localhost:%v", resource.GetPort("9000/tcp"))
	adapter, err := NewAdapter(endpoint, "ACCESSKEY", "SECRETKEY", false, "casbin-bucket", "policy.csv")
	if err != nil {
		t.Fatal("Failed to create adapter:", err)
	}

	m := model.Model{}
	m.LoadModel("examples/rbac_model.conf")

	enforcer := casbin.NewSyncedEnforcer(m, adapter)

	enforcer.EnableEnforce(true)
}
