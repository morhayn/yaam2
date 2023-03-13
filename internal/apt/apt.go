package apt

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

type Apt struct {
	ResponseWriter http.ResponseWriter
	RequestBody    io.ReadCloser
	RequestURI     string
	Repo           string
	Artifact       string
}

func (a Apt) downloadAgainIfInvalid(atf artifact.Artefact, resp *http.Response) error {
	log.Trace(resp.StatusCode)
	if resp.StatusCode == http.StatusOK {
		if err := file.CreateIfDoesNotExistInvalidOrEmpty(atf.Url, atf.Path, resp.Body, false); err != nil {
			return err
		}
	}

	if file.EmptyFile(atf.Path) {
		if err := a.Preserve(); err != nil {
			return err
		}
	}

	return nil
}

func (a Apt) Preserve(urlStrings ...string) error {
	urlString := a.RequestURI
	if len(urlStrings) > 0 {
		urlString = urlStrings[0]
	}
	log.Tracef("urlString: '%s'", urlString)

	repoInConfigFile, err := artifact.RepoInConfigFile(urlString, a.Repo, project.Conf.Caches.Apt)
	if err != nil {
		return err
	}

	if !reflect.ValueOf(repoInConfigFile).IsZero() {
		atf, err := artifact.NewArtefact(urlString, a.Artifact, repoInConfigFile)
		if err != nil {
			return err
		}

		resp, err := file.DownloadWithRetries(atf.Url, repoInConfigFile.User, repoInConfigFile.Pass)
		if err != nil {
			return err
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				panic(err)
			}
		}()

		if err := a.downloadAgainIfInvalid(atf, resp); err != nil {
			return err
		}
	}

	return nil
}
func (a Apt) Publish() error {
	if err := artifact.StoreOnDisk(a.RequestURI, a.RequestBody); err != nil {
		return err
	}

	return nil
}

func (a Apt) Read() error {
	if err := artifact.ReadFromDisk(a.ResponseWriter, a.RequestURI); err != nil {
		return fmt.Errorf(file.CannotReadErrMsg, err)
	}

	return nil
}
