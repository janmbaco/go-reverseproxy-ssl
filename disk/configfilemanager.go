package disk

import (
	"encoding/json"
	"github.com/janmbaco/go-reverseproxy-ssl/cross"
	"github.com/janmbaco/go-reverseproxy-ssl/events"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	ConfigFileChangedEvent = "ConfigFileChangedEvent"
)

type (
	ConstructorContentFunc func() interface {}
	CopyContentFunc func(interface{}, interface{})
)


type configFileManager struct{
	filePath             string
	onModifiedConfigFile events.SubscribeFunc
	config               interface{}
	watcherActive        bool
	constructorContent   ConstructorContentFunc
	copyContent          CopyContentFunc
}

func NewConfigFileManager(filepath string, constructor ConstructorContentFunc, copy CopyContentFunc) *configFileManager{
	return &configFileManager{
		filePath:             filepath,
		constructorContent:   constructor,
		copyContent:          copy,
	}
}

func (this *configFileManager) Load(config interface {}) {
	this.config = config
	if this.onModifiedConfigFile != nil {
		events.UnSubscribe(ModifiedFileEvent, &this.onModifiedConfigFile)
	}

	this.onModifiedConfigFile= func(args *events.EventArgs) {
		modifiedFile := strings.ReplaceAll(args.Args.(string), "\\", "/")
		if modifiedFile == this.filePath {
			time.Sleep(100)
			this.Load(config)
		}
	}
	successOnReading := this.read()
	if !ExistsPath(this.filePath) || !successOnReading {
		cross.Log.Warning("Creating config file from memory")
		this.create()
	}
	if !this.watcherActive {
		cross.TryPanic(Watcher.Add(this.filePath))
		this.watcherActive = true
	}
	if successOnReading  {
		events.Publish(ConfigFileChangedEvent, events.NewEventArgs(config, nil))
	}

	events.Subscribe(ModifiedFileEvent, &this.onModifiedConfigFile)
}

func  (this *configFileManager) create() {
	if this.watcherActive {
		_ = Watcher.Remove(this.filePath)
		this.watcherActive = false
	}
	configFile, err := json.Marshal(this.config)
	cross.TryPanic(err)
	_ = os.MkdirAll(filepath.Dir(this.filePath), 0666)
	cross.TryPanic(CreateFile(this.filePath, configFile))
	cross.TryPanic(Watcher.Add(this.filePath))
	this.watcherActive = true
}

func (this *configFileManager) read() bool {
	//it's possible that it still modifying
	time.Sleep(100)
	content, err := ioutil.ReadFile(this.filePath)
	if err == nil {
		conf := this.constructorContent()
		err = json.Unmarshal(content, conf)
		if err == nil{
			this.copyContent(conf, this.config)
		}
	}
	if err != nil {
		cross.Log.Warning(err.Error())
	}
	return err == nil
}
