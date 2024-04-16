package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
)

type Hasher struct {
	blockSize       int
	encodedToString bool
}

// NewHasher creates a new NewHasher with the given io.Reader and block size.
func NewHasher(blockSize int, encodedToString bool) *Hasher {
	return &Hasher{
		blockSize:       blockSize,
		encodedToString: encodedToString,
	}
}

// Hash calculates the hash of the next block of data in the file.
// Returns true if there is more data to be hashed, false otherwise.
func (fh *Hasher) Hash(r io.Reader) ([]byte, bool, error) {
	buf := make([]byte, fh.blockSize)
	bytesRead, err := r.Read(buf)
	if err != nil && err != io.EOF {
		return nil, false, err
	}

	if bytesRead == 0 {
		return nil, err != io.EOF, nil
	}

	hash := sha256.Sum256(buf[:bytesRead])
	result := hash[:]
	if fh.encodedToString {
		result = []byte(hex.EncodeToString(result))
	}

	return result, err != io.EOF, nil
}

// Sha256ToString converts the given SHA-256 hash to a hex string.
func Sha256ToString(hash []byte) string {
	return hex.EncodeToString(hash)
}
