package event

import (
	pb "grpc-html-to-pdf/internal/uploader/proto"
	"time"
)

type Status int

const (
	pending Status = iota
	upload
	unzip
	convert
	finished
	failed
)

type Event struct {
	Stream      pb.Uploader_UploadServer
	FilePath    string
	Mem         int
	Dur         time.Duration
	Start       time.Time
	ArchiveSize int

	TempFolder string
	Status     Status
	Err        error
}
