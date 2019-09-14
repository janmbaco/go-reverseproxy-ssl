package disk

import (
	"github.com/fsnotify/fsnotify"
	"github.com/janmbaco/go-reverseproxy-ssl/cross"
	"github.com/janmbaco/go-reverseproxy-ssl/events"
)

var Watcher *fsnotify.Watcher

func init() {
	var err error
	Watcher, err = fsnotify.NewWatcher()
	cross.TryPanic(err)
	go watchFiles()
}

func watchFiles(){
		for {
			select {
			case event, ok := <-Watcher.Events:
				if ok {
					publish := func(ev string){
						events.Publish(ev, events.NewEventArgs(nil, event.Name))
					}
					if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Rename == fsnotify.Rename{
						publish("ModifiedFile")
					}
					if event.Op&fsnotify.Chmod == fsnotify.Chmod {
						publish("ChangedModFile")
					}
				}
			}
		}
	}

