package consts

import (
	"fmt"
	"os"
	"path"
)

var K_HOME_DIR = func() string {
	dir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("get user home dir error: %s", err.Error()))
	}

	return path.Join(dir, ".k")
}()
