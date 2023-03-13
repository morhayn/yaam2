package artifact

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/morhayn/yaam2/internal/file"
	"github.com/morhayn/yaam2/internal/project"

	log "github.com/sirupsen/logrus"
)

type Artefact struct {
	Path, Url string
}
type PublicRepository struct {
	Name, Url, User, Pass string
}

// Create new structure
func NewArtefact(url, artifact string, repos PublicRepository) (Artefact, error) {
	CloseUrlRepo(repos.Url)
	path, err := filepathOnDisk(url)
	if err != nil {
		return Artefact{}, err
	}
	if err := DirCreate(url); err != nil {
		return Artefact{}, err
	}
	du := fmt.Sprintf("%s%s", repos.Url, artifact)
	log.Debugf("completeFile: '%s', downloadUrl: '%s'", path, du)
	return Artefact{Path: path, Url: du}, nil
}

// Create file if not exists
func createIfDoesNotExist(path string, requestBody io.ReadCloser) error {
	if _, fileExists := file.Exists(path); !fileExists {
		dst, err := os.Create(filepath.Clean(path))
		if err != nil {
			return err
		}
		defer func() {
			if err := dst.Close(); err != nil {
				panic(err)
			}
		}()
		w, err := io.Copy(dst, requestBody)
		if err != nil {
			log.Error(err)
		}
		log.Debugf("file: '%s' created and it contains: '%d' bytes", path, w)
		if err := dst.Sync(); err != nil {
			return err
		}
	} else {
		log.Tracef("file: '%s' exists already", path)
	}
	return nil
}

// Push package to disk
func StoreOnDisk(requestURI string, requestBody io.ReadCloser) error {
	// path, err := createHomeAndReturnPath(requestURI)
	path, err := filepathOnDisk(requestURI)
	if err != nil {
		return err
	}
	err = DirCreate(requestURI)
	if err != nil {
		return err
	}
	if err := createIfDoesNotExist(path, requestBody); err != nil {
		return err
	}
	return nil
}

// Path from url request
func filepathOnDisk(url string) (string, error) {
	h, err := project.RepositoriesHome()
	if err != nil {
		return "", err
	}
	f := filepath.Join(h, url)
	log.Debugf("constructed filepath: '%s' after concatenating home: '%s' to url: '%s'", f, h, url)
	return f, nil
}

// Read and send to Response file
func ReadFromDisk(w http.ResponseWriter, reqURL string) error {
	f, err := filepathOnDisk(reqURL)
	if err != nil {
		return err
	}
	log.Tracef("reading file: '%s' from disk...", f)

	if filepath.Ext(f) == ".tmp" {
		w.Header().Set("Content-Type", "application/json")
	}
	if filepath.Ext(f) == ".tgz" {
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	size, ok := file.Exists(f)
	if !ok {
		return errors.New(fmt.Sprintf("File not exists %s", f))
	}
	w.Header().Set("Content-Length", fmt.Sprint(size))
	b, err := os.ReadFile(filepath.Clean(f))
	if err != nil {
		return err
	}
	if _, err := io.WriteString(w, string(b)); err != nil {
		return err
	}

	return nil
}

// ReadRepositoriesAndUrlsFromConfigFileAndCacheArtifact reads a repositories
// yaml file that contains repositories and their URLs. If a request is
// attempted to download a file, it will look up the name in the config file
// and find the public URLs so it can download the file from the public maven
// repository and cache it on disk.
func RepoInConfigFile(urlString, repo string, configRepos map[string]project.Repos) (PublicRepository, error) {
	if len(configRepos) == 0 {
		return PublicRepository{}, fmt.Errorf("caches: '%s' not found in config file", repo)
	}
	r, ok := configRepos[repo]
	if !ok {
		return PublicRepository{}, errors.New(fmt.Sprintf("Not repositori in config file %s", repo))
	}
	url := r.Url
	if url == "" {
		return PublicRepository{}, errors.New(fmt.Sprintf("Url empty in config file %s", repo))
	}
	user := r.User
	if user == "" {
		log.Tracef("user: '%s'", user)
	}
	pass := r.Pass
	if pass == "" {
		log.Tracef("pass: **********")
	}

	log.Debugf("trying to cache artifact from: '%s'...", urlString)

	pr := PublicRepository{Name: repo, Url: url}
	if user != "" && pass != "" {
		pr.User = user
		pr.Pass = pass
	}

	return pr, nil
}

// Create directoryes
func DirCreate(url string) error {
	path, err := filepathOnDisk(url)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	return nil
}

// Add close '/' to end url
func CloseUrlRepo(url string) string {
	if string(url[len(url)-1:]) != "/" {
		url = fmt.Sprintf("%s/", url)
	}
	return url
}
