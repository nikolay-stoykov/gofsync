package hash

import "io"

type MetadataFile struct {
	blockSize int64
	writer    io.WriteCloser
}

func NewMetadataFile(blockSize int64) *MetadataFile {
	return &MetadataFile{
		blockSize: blockSize,
	}
}

// func (mf *MetadataFile) WriteNext([]byte) error {

// }
