package watcher

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
)

type eventFunc func(fsnotify.Event) error

type Watcher struct {
	dirPath         string
	metadataFileDir string
	watcher         *fsnotify.Watcher
	started         bool

	mutex sync.Mutex

	eventHandlers []eventFunc
}

func NewWatcher(dirPath, metadataFileDir string) *Watcher {
	return &Watcher{
		dirPath:         dirPath,
		metadataFileDir: metadataFileDir,
		started:         false,
		eventHandlers:   make([]eventFunc, 0),
		mutex:           sync.Mutex{},
	}
}

func recursiveWatchHandler(w *Watcher) eventFunc {
	return func(e fsnotify.Event) error {
		if e.Op != fsnotify.Create {
			return nil
		}

		finfo, err := os.Stat(e.Name)
		if err != nil {
			return err
		}

		return w.watchDir(e.Name, finfo, nil)
	}
}

func (w *Watcher) Start(started chan<- struct{}) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	w.watcher = watcher

	w.eventHandlers = append(w.eventHandlers, recursiveWatchHandler(w))

	if err := filepath.Walk(w.dirPath, w.watchDir); err != nil {
		return err
	}

	go w.process()

	close(started)

	return nil
}

func (w *Watcher) AddHandler(h eventFunc) {
	w.eventHandlers = append(w.eventHandlers, h)
}

func (w *Watcher) process() {
	for ev := range w.watcher.Events {
		for _, eh := range w.eventHandlers {
			if err := eh(ev); err != nil {
				log.Println("could not process event", err)
				// TODO: Add event for re-process
				// TODO: Handle errors on watcher
			}
		}
	}
}

// watchDir gets run as a walk func, searching for directories to add watchers to
func (w *Watcher) watchDir(path string, fi os.FileInfo, err error) error {
	// since fsnotify can watch all the files in a directory, watchers only need
	// to be added to each nested directory
	if strings.HasSuffix(path, w.metadataFileDir) {
		return nil
	}

	if fi.Mode().IsDir() {
		return w.watcher.Add(path)
	}

	return nil
}
