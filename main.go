package main

import (
	"path/filepath"

	"github.com/pankrator/gofsync/internal/hash"
	"github.com/pankrator/gofsync/internal/watcher"
)

const (
	controlBlockSize = 1024 * 8
	fsDirectory      = "/Users/nikolay.stoykov/work/deploy-infra"
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
	)

	if err := dirHasher.ReadFiles(); err != nil {
		panic(err)
	}

	router := watcher.NewRouter(fileProcessor)

	started := make(chan struct{})
	fsWatcher := watcher.NewWatcher(fsDirectory, metadataFileDir)
	fsWatcher.AddHandler(router.HFunc)

	go fsWatcher.Start(started)
	<-started

	select {}
}
