package response

import (
	"time"
)

type FileType int

const (
	NOT_FOUND = iota
	FILE
	FINGERPRINTED_FILE
	INDEX
	INDEX_PROXY
	VERSION
)

type ResponseEntity struct {
	Path          string
	fileType      FileType
	Size          int64
	ModTime       time.Time
	ContentType   string
	Compressed    bool
	Content       []byte
	ContentBrotli []byte
	ContentGzip   []byte
}

func (entity ResponseEntity) IsNotFound() bool {
	return entity.fileType == NOT_FOUND
}

func (entity ResponseEntity) IsIndex() bool {
	return entity.fileType == INDEX
}

func (entity ResponseEntity) IsIndexProxy() bool {
	return entity.fileType == INDEX_PROXY
}

func (entity ResponseEntity) IsFingerprinted() bool {
	return entity.fileType == FINGERPRINTED_FILE
}
