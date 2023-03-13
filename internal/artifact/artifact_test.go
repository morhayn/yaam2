package artifact

import (
	"fmt"
	"github.com/morhayn/yaam2/internal/project"
	"testing"
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
			t.Error(err)
		}
		if pr.Name != "test" && pr.Url != "http://test.local" && pr.Name != "user" && pr.Pass != "12345" {
			t.Error("Response not wrong ", pr)
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
			t.Error(err)
		}
		if pr.Name != "test" && pr.Url != "http://test.local" && pr.Name != "" && pr.Pass != "" {
			t.Error("Response not wrong ", pr)
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
			t.Error(err)
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
			t.Error(err)
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
			t.Error(err)
		}

	})
}
func TestCloseUrlPath(t *testing.T) {
	t.Run("url close", func(t *testing.T) {
		url := "http://test.local/apt/debian/"
		resp := CloseUrlRepo(url)
		if url != resp {
			t.Errorf("Not match %s %s \n", url, resp)
		}
	})
	t.Run("url not close", func(t *testing.T) {
		url := "http://test.local/apt/debian"
		resp := CloseUrlRepo(url)
		if url == resp {
			t.Errorf("Error match %s %s \n", url, resp)
		}
		if resp != fmt.Sprintf("%s/", url) {
			t.Error("url not close")
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
			t.Error(err)
		}
		if resp != fmt.Sprintf("%srepositories/%s", dir, url) {
			t.Error("response not true ", resp)
		}
	})
	t.Run("error test", func(t *testing.T) {
		url := "/npm/-/test"
		project.Conf.CacheDir = ""
		_, err := filepathOnDisk(url)
		if err.Error() != "Cache Directory not in config file" {
			t.Error(err)
		}
	})
}
