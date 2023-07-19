package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/KevinWang15/k/pkg/consts"
	"github.com/KevinWang15/k/pkg/model"
)

func GetConfig() model.Config {
	if os.Getenv(consts.K_CONFIG_FILE) == "" {
		panic("K_CONFIG_FILE is not set")
	}

	bytes, err := ioutil.ReadFile(os.Getenv(consts.K_CONFIG_FILE))
	if err != nil {
		panic(fmt.Errorf("read file %s error: %s", os.Getenv(consts.K_CONFIG_FILE), err.Error()))
	}

	var config model.Config

	err = json.Unmarshal(bytes, &config)
	if err != nil {
		panic(fmt.Errorf("unmarshal clusters error: %s", err.Error()))
	}

	return config
}

func EnsureKHomeDir() {
	err := os.MkdirAll(consts.K_HOME_DIR, os.ModePerm)
	if err != nil {
		panic(fmt.Errorf("create dir %q error: %s", consts.K_HOME_DIR, err.Error()))
	}
}
