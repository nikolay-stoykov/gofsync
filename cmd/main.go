package main

import (
	"path/filepath"

	"github.com/pankrator/gofsync/internal/hash"
	"github.com/pankrator/gofsync/internal/watcher"
)

const (
	controlBlockSize = 1024 * 8
	fsDirectory      = "/home/pankrator/go"
	hexMetadata      = false
)

func main() {
	metadataFileDir := filepath.Join(fsDirectory, ".metadata")
	hasher := hash.NewHasher(controlBlockSize, hexMetadata)

	fileProcessor := hash.NewFileProcessor(metadataFileDir, hasher)

	dirHasher := hash.NewDirectoryHasher(
		fsDirectory,
		filepath.Join(fsDirectory, ".metadata"),
		fileProcessor,
		100,
	)

	if err := dirHasher.ReadFiles(); err != nil {
		panic(err)
	}

	router := watcher.NewRouter(fileProcessor)

	// TODO: Run the watcher before the rehasher but buffer events until the rehasher has finished

	started := make(chan struct{})
	fsWatcher := watcher.NewWatcher(fsDirectory, metadataFileDir)
	fsWatcher.AddHandler(router.HFunc)

	go fsWatcher.Start(started)
	<-started

	select {}
}
