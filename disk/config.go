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

type configFile struct{
	filePath             string
	onModifiedConfigFile events.SubscribeFunc
	config               interface{}
	watcherActive        bool
	ConstructorContent   func() interface{}
	CopyContent          func(interface{}, interface{})
}

var ConfigFile *configFile

func init(){
	ConfigFile = &configFile{}
}

func (this *configFile) SetDir(dir string){
	abs, _ :=  filepath.Abs(dir)
	abs = strings.ReplaceAll(abs, "\\", "/")
	this.filePath = abs + "/" + filepath.Base(os.Args[0]) + ".config"
}


func (this *configFile) Load(config interface {}) {
	this.config = config
	if this.onModifiedConfigFile != nil {
		events.UnSubscribe("ModifiedFile", &this.onModifiedConfigFile)
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
		events.Publish("ConfigFileChanged", events.NewEventArgs(config, nil))
	}

	events.Subscribe("ModifiedFile", &this.onModifiedConfigFile)
}

func  (this *configFile) create() {
	if this.watcherActive {
		Watcher.Remove(this.filePath)
		this.watcherActive = false
	}
	json, err := json.Marshal(this.config)
	cross.TryPanic(err)
	os.MkdirAll(filepath.Dir(this.filePath), 0666)
	cross.TryPanic(CreateFile(this.filePath, json))
	cross.TryPanic(Watcher.Add(this.filePath))
	this.watcherActive = true
}

func (this *configFile) read() bool {
	//is posible that it still modifying
	time.Sleep(100)
	content, err := ioutil.ReadFile(this.filePath)
	if err == nil {
		conf := this.ConstructorContent()
		err = json.Unmarshal(content, conf)
		if err == nil{
			this.CopyContent(conf, this.config)
		}
	}
	if err != nil {
		cross.Log.Warning(err.Error())
	}
	return err == nil
}
