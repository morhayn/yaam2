package maven

import (
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/morhayn/yaam2/internal/artifact"
	"github.com/morhayn/yaam2/internal/file"
	"github.com/morhayn/yaam2/internal/project"

	log "github.com/sirupsen/logrus"
)

type Maven struct {
	ResponseWriter http.ResponseWriter
	RequestBody    io.ReadCloser
	RequestURI     string
	Repo           string
	Artifact       string
}

func (m Maven) downloadAgainIfInvalid(a artifact.Artefact, resp *http.Response) error {
	log.Trace(resp.StatusCode)
	fmt.Println(resp.StatusCode)
	if resp.StatusCode == http.StatusOK {
		if err := file.CreateIfDoesNotExistInvalidOrEmpty(a.Url, a.Path, resp.Body, false); err != nil {
			fmt.Println("Save file", err)
			return err
		}
	}

	if file.EmptyFile(a.Path) {
		if err := m.Preserve(); err != nil {
			return err
		}
	}

	return nil
}

func (m Maven) Preserve(urlStrings ...string) error {
	fmt.Println("MAVEN", m)
	urlString := m.RequestURI
	if len(urlStrings) > 0 {
		urlString = urlStrings[0]
	}
	log.Tracef("urlString: '%s'", urlString)

	repoInConfigFile, err := artifact.RepoInConfigFile(urlString, m.Repo, project.Conf.Caches.Maven)
	if err != nil {
		return err
	}

	if !reflect.ValueOf(repoInConfigFile).IsZero() {
		a, err := artifact.NewArtefact(urlString, m.Artifact, repoInConfigFile) // m.Repo repoInConfigFile)
		if err != nil {
			return err
		}
		fmt.Println(a.Url, repoInConfigFile)
		resp, err := file.DownloadWithRetries(a.Url, repoInConfigFile.User, repoInConfigFile.Pass)
		if err != nil {
			return err
		}
		fmt.Println("+++++", a.Url, resp.Header)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				panic(err)
			}
		}()

		if err := m.downloadAgainIfInvalid(a, resp); err != nil {
			return err
		}
	}

	return nil
}

func (m Maven) Publish() error {
	if err := artifact.StoreOnDisk(m.RequestURI, m.RequestBody); err != nil {
		return err
	}

	return nil
}

func (m Maven) Read() error {
	if err := artifact.ReadFromDisk(m.ResponseWriter, m.RequestURI); err != nil {
		return fmt.Errorf(file.CannotReadErrMsg, err)
	}

	return nil
}

// func (m Maven) Unify(name string) error {
// repos, err := artifact.AllowedRepos(name)
// if err != nil {
// return err
// }

// log.Debugf("repos: '%v'", repos)
// for _, repo := range repos {
// log.Tracef("repo: '%s'", repo)
// urlString := "/" + repo + "/" + m.RequestURI
// log.Debugf("urlString: '%s'", urlString)

// h, err := project.RepositoriesHome()
// if err != nil {
// return err
// }

// if err := m.Preserve(urlString); err != nil {
// log.Errorf("maven artifact caching failed. Error: '%v'", err)
// }

// if _, fileExists := file.Exists(filepath.Join(h, urlString)); fileExists {
// if err := artifact.ReadFromDisk(m.ResponseWriter, urlString); err != nil {
// log.Warnf(file.CannotReadErrMsg, err)
// }
// return nil
// }
// }

// return nil
// }
