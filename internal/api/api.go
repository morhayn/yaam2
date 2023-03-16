package api

import (
	"fmt"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
)

func basicAuth(r *http.Request) error {
	u, p, ok := r.BasicAuth()
	log.Debugf("user: '%s', pass: '********', basicAuthUsed?: '%t'", u, ok)
	if ok {
		if !(u == os.Getenv("YAAM_USER") && p == os.Getenv("YAAM_PASS")) {
			// return fmt.Errorf("auth failed")
			return nil //off Auth
		}
	} else {
		// return fmt.Errorf("request is NOT using basic authentication")
		return nil
	}

	return nil
}

func Validation(method string, r *http.Request, w http.ResponseWriter) error {
	if !(method == "PUT" || method == "POST" || method == "GET" || method == "HEAD") {
		return fmt.Errorf("only PUTs, POSTs, GETs and HEADs are supported. Used method: '%s'", method)
	}

	if err := basicAuth(r); err != nil {
		return fmt.Errorf("basic auth failed. Error: '%v'", err)
	}

	return nil
}
