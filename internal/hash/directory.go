package hash

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type PathFileProcessor interface {
	Process(path string) error
}

type DirectoryHasher struct {
	metadataDirPath string
	path            string
	fileProcessor   PathFileProcessor

	tokens      chan struct{}
	files       chan *fileInfo
	concurrency int
}

type fileInfo struct {
	path  string
	entry fs.DirEntry
}

// NewDirectoryHasher creates a new FileReader with the given directory path and FileHasher.
func NewDirectoryHasher(path, metadataFilesDir string, fileProcessor PathFileProcessor, concurrency int) *DirectoryHasher {
	return &DirectoryHasher{
		path:            path,
		fileProcessor:   fileProcessor,
		metadataDirPath: metadataFilesDir,

		tokens:      make(chan struct{}, concurrency),
		files:       make(chan *fileInfo, 1000),
		concurrency: concurrency,
	}
}

// ReadFiles reads all files in the directory and returns their contents as a map,
// with the SHA-256 hash of each file as the value.
func (fr *DirectoryHasher) ReadFiles() error {
	if err := fr.ensureMetadataDir(); err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	start := time.Now()

	for i := 0; i < fr.concurrency; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for file := range fr.files {
				if strings.Contains(filepath.Dir(file.path), fr.metadataDirPath) {
					continue
				}

				if !file.entry.Type().IsRegular() {
					continue
				}

				if err := fr.fileProcessor.Process(file.path); err != nil {
					log.Println(err)
				}
			}
		}()
	}

	startWalk := time.Now()

	if err := filepath.WalkDir(fr.path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		fr.files <- &fileInfo{
			path:  path,
			entry: d,
		}

		return nil
	}); err != nil {
		return err
	}

	close(fr.files)

	elapsedWalk := time.Since(startWalk)

	wg.Wait()

	elapsed := time.Since(start)
	log.Println("Rehasher finished in", elapsed.Milliseconds())
	log.Println("File walk finished in", elapsedWalk.Milliseconds())

	return nil
}

func (fr *DirectoryHasher) ensureMetadataDir() error {
	_, err := os.Stat(fr.metadataDirPath)
	if os.IsNotExist(err) {
		err = os.Mkdir(fr.metadataDirPath, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}
