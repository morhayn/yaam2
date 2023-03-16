package artifact

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/morhayn/yaam2/internal/project"
)

func TestRepoInConfigure(t *testing.T) {
	t.Run("simple RepoInConfigFile", func(t *testing.T) {
		project.Conf = project.ConfigFile{
			Port: "25213",
			Caches: project.Rep{
				Npm: map[string]project.Repos{
					"test": {
						Url:  "http://test.local",
						User: "user",
						Pass: "12345",
					},
				},
			},
		}
		pr, err := RepoInConfigFile("test/npm/package.tgz", "test", project.Conf.Caches.Npm)
		if err != nil {
			t.Fatal(err)
		}
		repo := PublicRepository{
			Name: "test",
			Url:  "http://test.local/",
			User: "user",
			Pass: "12345",
		}
		if pr != repo {
			t.Fatal("Response not wrong ", pr)
		}
	})
	t.Run("not pass RepoInConfigFile", func(t *testing.T) {
		project.Conf = project.ConfigFile{
			Port: "25213",
			Caches: project.Rep{
				Npm: map[string]project.Repos{
					"test": {
						Url:  "http://test.local",
						User: "user",
					},
				},
			},
		}
		pr, err := RepoInConfigFile("test/npm/package.tgz", "test", project.Conf.Caches.Npm)
		if err != nil {
			t.Fatal(err)
		}
		repo := PublicRepository{
			Name: "test",
			Url:  "http://test.local/",
			User: "",
			Pass: "",
		}
		if pr != repo {
			t.Fatal("Response not wrong ", pr)
		}
	})
	t.Run("not repositories RepoInConfigFile", func(t *testing.T) {
		project.Conf = project.ConfigFile{
			Port: "25213",
			Caches: project.Rep{
				Npm: map[string]project.Repos{},
			},
		}
		_, err := RepoInConfigFile("test/npm/package.tgz", "test", project.Conf.Caches.Npm)
		if err.Error() != "caches: 'test' not found in config file" {
			t.Fatal(err)
		}

	})
	t.Run("not test repositories RepoInConfigFile", func(t *testing.T) {
		project.Conf = project.ConfigFile{
			Port: "25213",
			Caches: project.Rep{
				Npm: map[string]project.Repos{
					"nottest": {
						Url: "http://test.local",
					},
				},
			},
		}
		_, err := RepoInConfigFile("test/npm/package.tgz", "test", project.Conf.Caches.Npm)
		if err.Error() != "Not repositori in config file test" {
			t.Fatal(err)
		}

	})
	t.Run("not url repository RepoInConfigFile", func(t *testing.T) {
		project.Conf = project.ConfigFile{
			Port: "25213",
			Caches: project.Rep{
				Npm: map[string]project.Repos{
					"test": {
						Url: "",
					},
				},
			},
		}
		_, err := RepoInConfigFile("test/npm/package.tgz", "test", project.Conf.Caches.Npm)
		if err.Error() != "Url empty in config file test" {
			t.Fatal(err)
		}

	})
}
func TestCloseUrlPath(t *testing.T) {
	t.Run("url close", func(t *testing.T) {
		url := "http://test.local/apt/debian/"
		resp := CloseUrlRepo(url)
		if url != resp {
			t.Fatalf("Not match %s %s \n", url, resp)
		}
	})
	t.Run("url not close", func(t *testing.T) {
		url := "http://test.local/apt/debian"
		resp := CloseUrlRepo(url)
		if url == resp {
			t.Fatalf("Error match %s %s \n", url, resp)
		}
		if resp != fmt.Sprintf("%s/", url) {
			t.Fatal("url not close")
		}
	})
}
func TestFilePathOnDisk(t *testing.T) {
	t.Run("simple test", func(t *testing.T) {
		dir := "/tmp/repos/"
		url := "npm/-/test"
		project.Conf.CacheDir = dir
		resp, err := filepathOnDisk(url)
		if err != nil {
			t.Fatal(err)
		}
		if resp != fmt.Sprintf("%srepositories/%s", dir, url) {
			t.Fatal("response not true ", resp)
		}
	})
	t.Run("error test", func(t *testing.T) {
		url := "/npm/-/test"
		project.Conf.CacheDir = ""
		_, err := filepathOnDisk(url)
		if err.Error() != "Cache Directory not in config file" {
			t.Fatal(err)
		}
	})
}
func TestDirCreate(t *testing.T) {
	t.Run("create /tmp/test/test1/ directory", func(t *testing.T) {
		project.Conf.CacheDir = "/tmp/"
		project.Conf.Port = "8080"
		path := "/tmp/repositories/test/test1/t.txt"
		err := DirCreate("test/test1/t.txt")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(path); err == nil {
			t.Fatal("file exist!!!")
		}
		if _, err := os.Stat(filepath.Dir(path)); err != nil {
			t.Fatal("Directory /tmp/repositories/test/test1 not create ", err)
		}
		err = os.RemoveAll("/tmp/repositories")
		if err != nil {
			t.Fatal("Error delete directory /tmp/repositories/ ", err)
		}
	})
}
func TestStoreOnDisk(t *testing.T) {
	t.Run("store file /tmp/repository/test.tmp", func(t *testing.T) {
		project.Conf.CacheDir = "/tmp/"
		path := "/tmp/repositories/test.tmp"
		data := io.NopCloser(strings.NewReader("Test"))
		err := StoreOnDisk("test.tmp", data)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(path); err != nil {
			t.Fatal("file not create")
		}
		if f, _ := os.ReadFile(path); string(f) != "Test" {
			t.Fatal("not Equal  `Test` != ", string(f))
		}
		if err = os.RemoveAll("/tmp/repositories"); err != nil {
			t.Fatal(err)
		}

	})
}
func TestNewArtifact(t *testing.T) {
	t.Run("simple test create struct", func(t *testing.T) {
		project.Conf.CacheDir = "/tmp"
		pr := PublicRepository{
			Name: "npm",
			Url:  "http://npmjs.org/",
			User: "test",
			Pass: "pass",
		}
		art, err := NewArtefact("tt/1.deb", "npm", pr)
		if err != nil {
			t.Fatal(err)
		}
		if art.Path != "/tmp/repositories/tt/1.deb" || art.Url != "http://npmjs.org/npm" {
			t.Fatal("Not Equal ", art)
		}
	})
}
