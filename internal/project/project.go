package project

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var Conf ConfigFile

type ConfigFile struct {
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Pass     string `yaml:"pass"`
	CacheDir string `yaml:"cachedir"`
	Caches   Rep    `yaml:"caches"`
}
type Rep struct {
	Apt   map[string]Repos `yaml:"apt"`
	Npm   map[string]Repos `yaml:"npm"`
	Maven map[string]Repos `yaml:"maven"`
}
type Repos struct {
	Url  string `yaml:"url"`
	User string `yaml:"user"`
	Pass string `yaml:"pass"`
}

func (c *ConfigFile) GetRepos(t string) {

}

const (
	// hiddenFolderName = ".yaam"
	// Port             = 25213
	host = "localhost"
	// Scheme = "http"
)

// var (
// PortString  = strconv.Itoa(Port)
// HostAndPort = Host + ":" + PortString
// Url         = Scheme + "://" + HostAndPort
// )

func (c *ConfigFile) ReadConfig(file string) error {
	f, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(f, &c)
	if err != nil {
		return err
	}
	return nil
}
func (c *ConfigFile) HostAndPort() string {
	return fmt.Sprintf("%s:%s", host, c.Port)
}

func RepositoriesHome() (string, error) {
	h := Conf.CacheDir
	if h == "" {
		return h, errors.New("Cache Directory not in config file")
	}
	h = filepath.Join(h, "repositories")
	return h, nil
}
