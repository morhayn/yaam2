package project

import (
	"os"
	"testing"
)

func TestRepositoriesHome(t *testing.T) {
	Conf.CacheDir = "/tmp/yaam2/"
	dir, err := RepositoriesHome()
	if err != nil {
		t.Fatal(err)
	}
	if dir != "/tmp/yaam2/repositories" {
		t.Fatal("Error Repo dir ", dir)
	}
}
func TestReadConfig(t *testing.T) {
	config := "port: 1000\nuser: " + `"test"` + "\npass: " + `"pass"` + "\ncachedir: " + `"/tmp/"` + "\ncaches:\n  apt:\n    debian:\n      url:" + ` "http://test.local"` + "\n"
	f, err := os.Create("/tmp/yaam2.yml")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	_, err = f.WriteString(config)
	if err != nil {
		t.Fatal(err)
	}
	Conf.ReadConfig("/tmp/yaam2.yml")
	if Conf.Caches.Apt["debian"].Url != "http://test.local" {
		t.Fatal(" Wrong Read Config File Cache Url")
	}
	if Conf.CacheDir != "/tmp/" {
		t.Fatal("Wrong Read Config File CacheDir")
	}
	err = os.Remove("/tmp/yaam2.yml")
	if err != nil {
		t.Fatalf("Fatal Remove file %s", err)
	}
}
