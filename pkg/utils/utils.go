package utils

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"

	"github.com/KevinWang15/k/pkg/consts"
	"github.com/KevinWang15/k/pkg/model"
)

func GetConfigPath() string {

	dir := consts.K_HOME_DIR
	err := os.MkdirAll(consts.K_HOME_DIR, os.ModePerm)
	if err != nil {
		panic(fmt.Errorf("trying to initialize k config: create dir %q error: %s", consts.K_HOME_DIR, err.Error()))
	}

	configJson := fmt.Sprintf("%s/%s", dir, "config.json")
	if _, err := os.Stat(configJson); os.IsNotExist(err) {
		err = ioutil.WriteFile(configJson, []byte("{}"), fs.ModePerm)
		if err != nil {
			panic(fmt.Errorf("trying to initialize k config: write file %q error: %s", configJson, err.Error()))
		}
	}

	return configJson
}

func GetConfig() model.Config {

	configPath := GetConfigPath()
	bytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		panic(fmt.Errorf("read file %s error: %s", configPath, err.Error()))
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
