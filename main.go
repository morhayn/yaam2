package main

import (
	"os"
	"path/filepath"

	"github.com/morhayn/yaam2/internal/webapi"
)

const (
	Conf = "yaam2.conf"
)

func main() {
	path, err := os.Executable()
	if err != nil {
		panic("Error get executable program file")
	}
	prDir := filepath.Dir(path)
	os.Chdir(prDir)
	webapi.Webapi(Conf)
}
