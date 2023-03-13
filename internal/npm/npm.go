package npm

import (
	"crypto/sha1" // #nosec
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/morhayn/yaam2/internal/artifact"
	"github.com/morhayn/yaam2/internal/file"
	"github.com/morhayn/yaam2/internal/project"

	log "github.com/sirupsen/logrus"
)

// Structure for npm manifest
type NpmPackage struct {
	Id             string             `json:"_id,omitempty"`
	Rev            string             `json:"_rev,omitempty"`
	Name           string             `json:"name,omitempty"`
	Description    interface{}        `json:"description,omitempty"`
	ReadMe         interface{}        `json:"readme,omitempty"`
	ReadmeFilename interface{}        `json:"readmeFilename,omitempty"`
	HomePage       interface{}        `json:"homepage,omitempty"`
	Keywords       interface{}        `json:"keywords,omitempty"`
	Author         interface{}        `json:"author,omitempty"`
	Bugs           interface{}        `json:"bugs,omitempty"`
	License        interface{}        `json:"license,omitempty"`
	DistTags       interface{}        `json:"dist-tags,omitempty"`
	Time           interface{}        `json:"time,omitempty"`
	Repository     interface{}        `json:"repository,omitempty"`
	Users          interface{}        `json:"users,omitempty"`
	Versions       map[string]Package `json:"versions"`
	Maintainers    interface{}        `json:"maintainers,omitempty"`
}
type Package struct {
	Id                     string            `json:"_id,omitempty"`
	NpmVersion             string            `json:"_npmVersion,omitempty"`
	NodeVersion            string            `json:"_nodeVersion,omitempty"`
	NpmUser                map[string]string `json:"_npmUser,omitempty"`
	From                   string            `json:"_from,omitempty"`
	EngineSupported        bool              `json:"_engineSupported,omitempty"`
	DefaultsLoaded         bool              `json:"_defaultsLoaded,omitempty"`
	HasShrinkwrap          bool              `json:"_hasShrinkwrap,omitempty"`
	Name                   string            `json:"name,omitempty"`
	GitHead                interface{}       `json:"gitHead,omitempty"`
	Description            interface{}       `json:"description,omitempty"`
	Version                interface{}       `json:"version,omitempty"`
	Main                   interface{}       `json:"main,omitempty"`
	PreferGlobal           interface{}       `json:"preferGlobal,omitempty"`
	HomePage               interface{}       `json:"homepage,omitempty"`
	Deprecared             interface{}       `json:"deprecated,omitempty"`
	License                interface{}       `json:"license,omitempty"`
	Keywords               interface{}       `json:"keywords,omitempty"`
	Author                 interface{}       `json:"author,omitempty"`
	Funding                interface{}       `json:"funding,omitempty"`
	Repository             interface{}       `json:"repository,omitempty"`
	Scripts                interface{}       `json:"scripts,omitempty"`
	Bin                    interface{}       `json:"bin,omitempty"`
	Bugs                   interface{}       `json:"bugs,omitempty"`
	Dependencies           interface{}       `json:"dependencies,omitempty"`
	DevDependencies        interface{}       `json:"devDependencies,omitempty"`
	OptionalDependencies   interface{}       `json:"optionalDependencies,omitempty"`
	Directories            interface{}       `json:"directories,omitempty"`
	Engines                interface{}       `json:"engines,omitempty"`
	Dist                   Dist              `json:"dist"`
	NpmOperationalInternal interface{}       `json:"_npmOperationalInternal,omitempty"`
	Maintainers            interface{}       `json:"maintainers,omitempty"`
}

// type Maintainer struct {
// Name  string `json:"name,omitempty"`
// Email string `json:"email,omitempty"`
// }
type Dist struct {
	ShaSum       string `json:"shasum,omitempty"`
	Tarball      string `json:"tarball"`
	Integrity    string `json:"integrity,omitempty"`
	Signatures   []Sign `json:"signatures,omitempty"`
	NpmSignature string `json:"npm-signature,omitempty"`
}
type Sign struct {
	KeyId string `json:"keyid,omitempty"`
	Sig   string `json:"sig,omitempty"`
}

type Npm struct {
	ResponseWriter http.ResponseWriter
	RequestBody    io.ReadCloser
	RequestURI     string
	Repo           string
	Artifact       string
}

var (
	NpmManifestEmpty    = errors.New("Npm manifest ZERO")
	NpmManifestNotValid = errors.New("npm manifest is invalid")
	ShortUrlString      = errors.New("short url string")
	CheckSumNotValid    = errors.New("checksum not match")
)

// Unmarshal npm manifest and replace url path to yaam repository
func replaceUrlPublicNpmWithYaamHost(f, repo, artifact string) error {
	input, err := os.ReadFile(filepath.Clean(f))
	if err != nil {
		return err
	}
	npm := NpmPackage{}
	err = json.Unmarshal([]byte(input), &npm)
	if err != nil {
		//Remove brocken manifest - no Unmarshell - no data for package
		err = os.Remove(filepath.Clean(f))
		if err != nil {
			panic(fmt.Sprintf("!!! broken npm not delete !!! %s", f))
		}
		return NpmManifestNotValid
	}
	// If Version in npm manifeste not exists manifest not valide
	if len(npm.Versions) == 0 {
		//Remove brocken npm manifest - no Versions - no Packages
		err = os.Remove(filepath.Clean(f))
		if err != nil {
			panic("!!! broken npm not delete !!!")
		}
		return NpmManifestEmpty
	}
	//For all version packages replace url path
	for key, vers := range npm.Versions {
		pack := path.Base(vers.Dist.Tarball)
		v := npm.Versions[key]
		v.Dist.Tarball = fmt.Sprintf("http://%s/npm/%s/%s/-/%s", project.Conf.HostAndPort(), repo, npm.Name, pack)
		npm.Versions[key] = v
	}
	//Write to disk new manifest for use and send clients
	file, _ := json.Marshal(npm)
	// output := strings.Replace(string(input), "https://registry.npmjs.org", "http://"+host+"/npm/3rdparty-npm", -1)
	err = os.WriteFile(f, []byte(file), 0o600)
	if err != nil {
		return err
	}
	//Check json format in file
	b, err := os.ReadFile(filepath.Clean(f))
	if err != nil {
		return err
	}
	if !json.Valid(b) {
		return fmt.Errorf("json for file: '%s' is invalid", f)
	}
	return nil
}

func firstMatch(f, regex string) (string, error) {
	re := regexp.MustCompile(regex)
	match := re.FindStringSubmatch(f)
	matchLength := len(match)
	log.Tracef("regex: '%s', match: '%v' matchLength: '%d' for file: '%s'", regex, match, matchLength, f)
	if matchLength <= 1 {
		return "", fmt.Errorf("no match was found for: '%s' with regex: '%s'", f, regex)
	}
	m := match[1]
	log.Tracef("firstMatch: '%s'", m)

	return m, nil
}

// Extract patch to matifest in- npm/react/-/react-3.3.0.tgz out- npm/react.tmp
func pathTmp(f string) (string, error) {
	arpath := strings.Split(f, "/")
	if len(arpath) < 4 {
		return "", ShortUrlString
	}
	path := filepath.Join(arpath[:len(arpath)-2]...)
	return filepath.Join("/" + path + ".tmp"), nil
}

// Get checkSum field from manifest for version package
func versionShasum(f string) (string, error) {
	arpath := strings.Split(f, "/")
	if len(arpath) < 1 {
		return "", ShortUrlString
	}
	//version package in request transform-29.5.0.tgz -- version="29.05.0"
	version, err := firstMatch(arpath[len(arpath)-1], `-([0-9]+\.[0-9]+\.[0-9]+(-.+)?)\.tgz$`)
	if err != nil {
		return "", err
	}
	pt, err := pathTmp(f)
	if err != nil {
		return "", err
	}
	//Open manifest file for packages
	b, err := os.ReadFile(filepath.Clean(pt))
	if err != nil {
		return "", err
	}
	npm := NpmPackage{}
	json.Unmarshal([]byte(b), &npm)
	//Get checkSum for "version" package
	value := npm.Versions[version].Dist.ShaSum
	return value, nil
}

// Calculate sha1 from file package on disk and check with manifest checksum
func compareChecksumOnDiskWithExpectedSha(expChecksum, pathTmp string) (bool, error) {
	checksumValid := true
	f, err := os.Open(filepath.Clean(pathTmp))
	if err != nil {
		return checksumValid, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()
	/* #nosec */
	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return checksumValid, err
	}
	// fmt.Printf("%x", h.Sum(nil))
	checksum := fmt.Sprintf("%x", h.Sum(nil))
	if checksum != expChecksum {
		log.Errorf("file: '%s' checksum on disk: '%s' does not match expected checksum: '%s'", pathTmp, checksum, expChecksum)
		checksumValid = false
		// time.Sleep(file.RetryDuration)
	}

	return checksumValid, nil
}

func checksum(f string) (bool, error) {
	checksumValid := true
	_, fileExists := file.Exists(f)
	if fileExists && filepath.Ext(f) == ".tgz" {
		fmt.Println(f)
		vs, err := versionShasum(f)
		if err != nil {
			return checksumValid, err
		}
		checksumValid, err := compareChecksumOnDiskWithExpectedSha(vs, f)
		if err != nil {
			fmt.Println("Checksum error")
			return checksumValid, err
		}
	}
	return checksumValid, nil
}

// Save file manifest(.tmp) and package(.tgz) to disk  validate json
func (n Npm) SaveToDisk(a artifact.Artefact, resp *http.Response) error {
	if resp.StatusCode == http.StatusOK {
		if err := file.CreateIfDoesNotExistInvalidOrEmpty(a.Url, a.Path, resp.Body, false); err != nil {
			fmt.Println(err)
			return err
		}
		if filepath.Ext(a.Path) == ".tgz" {
			checksumValid, err := checksum(a.Path)
			if err != nil {
				return err
			}
			// Remove brocken file with checksum not valid
			if !checksumValid {
				os.Remove(a.Path)
				return CheckSumNotValid
			}
		}
		if filepath.Ext(a.Path) == ".tmp" {
			b, err := os.ReadFile(filepath.Clean(a.Path))
			if err != nil {
				return err
			}
			//Remove brocken file if json not valid
			if !json.Valid(b) {
				log.Errorf("json file: '%s' is invalid", a.Path)
				os.Remove(a.Path)
				return NpmManifestNotValid
			}
		}
	} else {
		return errors.New(fmt.Sprintf("Download not comleated statusCode %d", resp.StatusCode))
	}
	return nil
}

// Load from external repository package and manifest
func (n Npm) Preserve(urlStrings ...string) error {
	fmt.Println(n.RequestURI)
	urlString := n.RequestURI
	if len(urlStrings) > 0 {
		urlString = urlStrings[0]
	}
	repoInConfigFile, err := artifact.RepoInConfigFile(urlString, n.Repo, project.Conf.Caches.Npm)
	if err != nil {
		return err
	}
	if !reflect.ValueOf(repoInConfigFile).IsZero() {
		h, err := project.RepositoriesHome()
		if err != nil {
			return err
		}
		dir := strings.Replace(urlString, "%2f", "/", -1)
		log.Debugf("extension found: '%s', file: '%s'", filepath.Ext(dir), dir)
		if filepath.Ext(dir) != ".tgz" {
			log.Debugf("file: '%s' does not have an extension", dir)
			dir = dir + ".tmp"
		}
		if err := artifact.DirCreate(dir); err != nil {
			return err
		}

		// log.Tracef("downloadUrl before entering downloadUrl method: '%s', regex: '%s'", urlString, repoInConfigFile.Regex)
		rep, ok := project.Conf.Caches.Npm[n.Repo]
		if !ok {
			return errors.New(fmt.Sprintf("Not repository in config file %s", n.Repo))
		}
		artifact.CloseUrlRepo(rep.Url)
		du := fmt.Sprintf("%s%s", rep.Url, n.Artifact)

		completeFile := filepath.Join(h, dir)
		// Fail .tgz exists and not download again
		if filepath.Ext(dir) == ".tgz" {
			if _, err = os.Stat(completeFile); err == nil {
				return nil
			}
		}

		a := artifact.Artefact{Path: completeFile, Url: du}
		resp, err := file.DownloadWithRetries(a.Url)
		if err != nil {
			return err
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				panic(err)
			}
		}()
		if err := n.SaveToDisk(a, resp); err != nil {
			return err
		}
		if filepath.Ext(dir) != ".tgz" {
			if err := replaceUrlPublicNpmWithYaamHost(completeFile, n.Repo, n.Artifact); err != nil {
				return err
			}
		}
	}
	return nil
}

// push npm package
func (n Npm) Publish() error {
	if err := artifact.StoreOnDisk(n.RequestURI, n.RequestBody); err != nil {
		return err
	}
	return nil
}

// Send to clent npm package or manifest
func (n Npm) Read() error {
	reqUrlString := strings.Replace(n.RequestURI, "%2f", "/", -1)
	if filepath.Ext(reqUrlString) != ".tgz" {
		log.Tracef("file: '%s' does not have an extension", reqUrlString)
		reqUrlString = reqUrlString + ".tmp"
	}
	if err := artifact.ReadFromDisk(n.ResponseWriter, reqUrlString); err != nil {
		return fmt.Errorf(file.CannotReadErrMsg, err)
	}
	return nil
}
