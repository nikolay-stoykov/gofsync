package watcher

import (
	"fmt"
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
)

type PathFileProcessor interface {
	Process(path string) error
}

type Router struct {
	fileProcessor PathFileProcessor
}

func NewRouter(fileProcessor PathFileProcessor) *Router {
	return &Router{
		fileProcessor: fileProcessor,
	}
}

func (r *Router) HFunc(ev fsnotify.Event) error {
	if ev.Op == fsnotify.Chmod {
		return nil
	}

	log.Println("event for", ev.Name, ev.Op.String())

	finfo, err := os.Stat(ev.Name)
	if err != nil {
		return fmt.Errorf("could not get file info: %w", err)
	}

	if !finfo.Mode().IsRegular() {
		return nil
	}

	return r.fileProcessor.Process(ev.Name)
}

// func (r *Router) FileChanged(ev fsnotify.Event) error {
// 	if ev.Op == fsnotify.Chmod {
// 		return nil
// 	}

// 	log.Println("event for", ev.Name, ev.Op.String())

// }
