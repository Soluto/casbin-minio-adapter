package minioadapter

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/casbin/casbin"
	"github.com/casbin/casbin/model"
	dc "github.com/fsouza/go-dockerclient"
	minio "github.com/minio/minio-go"
	dockertest "gopkg.in/ory-am/dockertest.v3"
)

var resource *dockertest.Resource

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
		client, err := minio.New(endpoint, "ACCESSKEY", "SECRETKEY", false)
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

		if _, err = client.FPutObject("casbin-bucket", "policy.csv", "./examples/rbac_policy.csv", minio.PutObjectOptions{ContentType: "application/csv"}); err != nil {
			log.Println("Failed to upload policy:", err)
			return err
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
	endpoint := fmt.Sprintf("localhost:%v", resource.GetPort("9000/tcp"))
	adapter, err := NewAdapter(endpoint, "ACCESSKEY", "SECRETKEY", false, "casbin-bucket", "policy.csv")
	if err != nil {
		t.Fatal("Failed to create adapter:", err)
	}

	model := model.Model{}
	model.LoadModel("examples/rbac_model.conf")

	err = adapter.LoadPolicy(model)
	if err != nil {
		t.Fatal("Failed to load policy:", err)
	}
}

func TestSavePolicy(t *testing.T) {
	endpoint := fmt.Sprintf("localhost:%v", resource.GetPort("9000/tcp"))
	adapter, err := NewAdapter(endpoint, "ACCESSKEY", "SECRETKEY", false, "casbin-bucket", "empty_policy.csv")
	if err != nil {
		t.Fatal("Failed to create adapter:", err)
	}

	model := model.Model{}
	model.LoadModel("examples/rbac_model.conf")

	err = adapter.SavePolicy(model)
	if err != nil {
		t.Fatal("Failed to save policy:", err)
	}
}

func TestWithEnforcer(t *testing.T) {

	endpoint := fmt.Sprintf("localhost:%v", resource.GetPort("9000/tcp"))
	adapter, err := NewAdapter(endpoint, "ACCESSKEY", "SECRETKEY", false, "casbin-bucket", "policy.csv")
	if err != nil {
		t.Fatal("Failed to create adapter:", err)
	}

	enforcerer := casbin.NewSyncedEnforcer("examples/rbac_model.conf", adapter)

	enforcerer.EnableEnforce(true)
}
