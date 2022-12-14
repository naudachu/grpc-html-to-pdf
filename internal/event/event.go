package event

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Event struct {
	UUID       uuid.UUID
	FilePath   string
	TempFolder string

	Dur         time.Duration
	Start       time.Time
	ArchiveSize int64

	Counter int
}

func NewEvent(path, dir string, size int64) *Event {
	return &Event{
		UUID:        uuid.New(),
		FilePath:    "",
		TempFolder:  "",
		Start:       time.Time{},
		ArchiveSize: 0,

		Dur: 0,
	}
}

func (e *Event) PostUpload(path, dir string, size int64) *Event {
	e.FilePath = path
	e.TempFolder = dir
	e.ArchiveSize = size
	e.Start = time.Now().UTC()
	return e
}

func (e *Event) String() string {
	return fmt.Sprintf("Task: %s; Duration: %s, Size: %s, Timestamp: %s", e.UUID.String(), e.Dur.String(), fmt.Sprint(e.ArchiveSize), e.Start)
}
