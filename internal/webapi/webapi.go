package webapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/morhayn/yaam2/internal/api"
	"github.com/morhayn/yaam2/internal/apt"
	"github.com/morhayn/yaam2/internal/artifact"
	"github.com/morhayn/yaam2/internal/maven"
	"github.com/morhayn/yaam2/internal/npm"
	"github.com/morhayn/yaam2/internal/project"

	"github.com/030/logging/pkg/logging"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

const serverLogMsg = "check the server logs"

var Version string

func httpNotFoundReadTheLogs(w http.ResponseWriter, err error, req string) {
	log.Error(err)
	fmt.Println(req)
	http.Error(w, serverLogMsg, http.StatusNotFound)
}

func httpInternalServerErrorReadTheLogs(w http.ResponseWriter, err error, req string) {
	log.Error(err)
	fmt.Println(req)
	http.Error(w, serverLogMsg, http.StatusInternalServerError)
}

func npmBulk(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := r.Body.Close(); err != nil {
			panic(err)
		}
	}()
	if r.Method == "POST" {
		var data struct{}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(data)
		// curl -XPOST -H "Content-type: application/json" -d '{ "ping": [ "0.4.2" ], "q": [ "1.5.1" ], "underscore": [ "1.13.6" ] }' 'https://registry.npmjs.org/-/npm/v1/security/advisories/bulk' -v
	}
}

func repoInterface(w http.ResponseWriter, r *http.Request, ar artifact.Artifacter, method string) {
	defer func() {
		if err := r.Body.Close(); err != nil {
			panic(err)
		}
	}()
	if err := api.Validation(r.Method, r, w); err != nil {
		httpInternalServerErrorReadTheLogs(w, err, r.RequestURI)
		return
	}
	if r.Method == method {
		if err := ar.Publish(); err != nil {
			httpInternalServerErrorReadTheLogs(w, err, r.RequestURI)
			return
		}
		return
	}

	if err := ar.Preserve(); err != nil {
		httpNotFoundReadTheLogs(w, fmt.Errorf("maven artifact caching failed. Error: '%v'", err), r.RequestURI)
		return
	}

	if err := ar.Read(); err != nil {
		httpNotFoundReadTheLogs(w, fmt.Errorf("cannot read artifact from disk. Error: '%v'. Perhaps it resides in another repository?", err), r.RequestURI)
		return
	}
}
func repository(w http.ResponseWriter, r *http.Request) {
	var ar artifact.Artifacter
	method := "PUT"
	vars := mux.Vars(r)
	switch vars["pack"] {
	case "npm":
		ar = npm.Npm{RequestBody: r.Body, RequestURI: r.RequestURI, ResponseWriter: w, Repo: vars["repo"], Artifact: vars["artifact"]}
		method = "POST"
	case "apt":
		ar = apt.Apt{RequestBody: r.Body, RequestURI: r.RequestURI, ResponseWriter: w, Repo: vars["repo"], Artifact: vars["artifact"]}
	case "maven":
		ar = maven.Maven{RequestBody: r.Body, RequestURI: r.RequestURI, ResponseWriter: w, Repo: vars["repo"], Artifact: vars["artifact"]}
	default:
		httpNotFoundReadTheLogs(w, errors.New("not found repository"), r.RequestURI)
	}
	repoInterface(w, r, ar, method)
}
func status(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := r.Body.Close(); err != nil {
			panic(err)
		}
	}()

	if _, err := io.WriteString(w, "ok"); err != nil {
		httpNotFoundReadTheLogs(w, err, r.RequestURI)
		return
	}
}

func Webapi(conf string) {
	project.Conf.ReadConfig(conf)
	logLevel := "info"
	logLevelEnv := os.Getenv("YAAM_LOG_LEVEL")
	if logLevelEnv != "" {
		logLevel = logLevelEnv
	}
	h := project.Conf.CacheDir

	dir := filepath.Join(h, "logs")
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		log.Fatal(err)
	}
	l := logging.Logging{File: filepath.Join(dir, "yaam.log"), Level: logLevel, Syslog: true}
	if _, err := l.Setup(); err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/npm/{repo}/-/npm/v1/security/advisories/bulk", npmBulk)
	r.HandleFunc("/npm/{repo}/-/npm/v1/security/audits/quick", npmBulk)
	r.HandleFunc("/{pack}/{repo}/{artifact:.*}", repository)
	// r.HandleFunc("/{pack}/{repo}/{artifact:.*}", Artifact)
	// r.HandleFunc("/{pack}/{repo}/{artifact:.*}", Artifact)
	// r.HandleFunc("/generic/{repo}/{artifact:.*}", genericArtifact)
	// r.HandleFunc("/maven/groups/{name}/{artifact:.*}", mavenGroup)
	r.HandleFunc("/status", status)

	srv := &http.Server{
		Addr: "0.0.0.0:" + project.Conf.Port, // project.PortString,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 120,
		ReadTimeout:  time.Second * 180,
		IdleTimeout:  time.Second * 240,
		Handler:      r, // Pass our instance of gorilla/mux in.
	}

	log.Infof("Starting YAAM version: '%s' on localhost on port: '%d'...", Version, project.Conf.Port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
