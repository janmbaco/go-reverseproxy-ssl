package disk

import (
	"encoding/json"
	"github.com/janmbaco/go-reverseproxy-ssl/cross"
	"github.com/janmbaco/go-reverseproxy-ssl/events"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

type configFile struct{
	filePath string
	OnModifiedConfigFile events.SubscribeFunc
	config interface{}
}

var configFileInstance *configFile

func (configFile *configFile) CreateEvent(){
	configFile.OnModifiedConfigFile = func(args *events.EventArgs) {
		modifiedFile := args.Args.(string)
		if modifiedFile == configFile.filePath {
			time.Sleep(100)
			LoadConfig(configFile.config)
		}
	}
}

func LoadConfig(config interface {}) {

	if configFileInstance == nil{
		configFilePath, _ := filepath.Abs(os.Args[0])
		configFilePath+= ".config"
		configFileInstance = &configFile{
			filePath:  configFilePath,
			config: config,
			}
		configFileInstance.CreateEvent()
	}

	Watcher.Remove(configFileInstance.filePath)
	successOnReading := readConfigFile()
	if !ExistsPath(configFileInstance.filePath) || !successOnReading {
		cross.Log.Warning("The config file is empty")
		createConfigFile()
	}
	if successOnReading  {
		events.Publish("ConfigFileChanged", events.NewEventArgs(config, nil))
	}
	Watcher.Add(configFileInstance.filePath)
}

func  createConfigFile() {
	json, err := json.Marshal(configFileInstance.config)
	cross.TryPanic(err)
	os.MkdirAll(filepath.Dir(configFileInstance.filePath), 0666)
	cross.TryPanic(CreateFile(configFileInstance.filePath, json))
}

func  readConfigFile() bool {
	content, err := ioutil.ReadFile(configFileInstance.filePath)
	if err == nil {
		err = json.Unmarshal(content, configFileInstance.config)
	}
	return err != nil
}
