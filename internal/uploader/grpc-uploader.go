package uploader

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	conv "grpc-html-to-pdf/internal/converter"
	"grpc-html-to-pdf/internal/event"
	pb "grpc-html-to-pdf/internal/uploader/proto"
	"io"
	"log"
	"os"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	maxFileSize = 2 << 30
)

// GRPCServer ...
type Server struct {
	pb.UnimplementedUploaderServer
	tasks []*event.Task
	pool  event.Pool
}

func NewUploaderService() *Server {
	var tasks []*event.Task
	return &Server{
		UnimplementedUploaderServer: pb.UnimplementedUploaderServer{},
		tasks:                       tasks,
		pool:                        *event.NewPool(tasks, 10),
	}
}

func (s *Server) RunBackground() {
	s.pool.RunBackground()
}

// PingPong ...
func (s *Server) PingPong(ctx context.Context, req *pb.Ping) (*pb.Pong, error) {
	fmt.Println(req.Text)
	return &pb.Pong{Text: "pong!"}, nil
}

func tempDirectoriesPreparation() (string, error) {
	//Create a folder for all Uploaded files
	path := "uploads"
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			logError(err)
		}
	}

	//Prepare temporary folder for uploaded archive
	tempDir, err := os.MkdirTemp("./uploads/", "tmp-*")
	if err != nil {
		return "", logError(err)
	}
	log.Print("temp dirs created")
	return tempDir, nil
}

// Upload ...
func (s *Server) Upload(stream pb.Uploader_UploadServer) error {

	e := &event.Event{
		Stream:      stream,
		FilePath:    "",
		Mem:         0,
		Dur:         0,
		Start:       time.Now().UTC(),
		ArchiveSize: 0,
		TempFolder:  "",
		Status:      0,
	}

	t := event.NewTask(func(i interface{}) error {
		err := processConvertion(i.(event.Event))
		if err != nil {
			logError(err)
		}
		return nil
	}, *e)

	s.pool.AddTask(t)

	return nil
}

func processConvertion(e event.Event) error {

	tempDir, err := tempDirectoriesPreparation()
	if err != nil {
		logError(fmt.Errorf("cannot create temp directory: %w", err))
	}

	fileData := bytes.Buffer{}
	fileSize := 0
	// Reading the stream
	for {
		err := contextError(e.Stream.Context())
		if err != nil {
			return err
		}

		log.Print("waiting to receive more data")

		req, err := e.Stream.Recv()
		if err == io.EOF {
			log.Print("no more data")
			break
		}
		if err != nil {
			return logError(status.Errorf(codes.Unknown, "cannot receive chunk data: %v", err))
		}

		chunk := req.GetChunk()
		size := len(chunk)
		fileSize += size

		log.Printf("new chunk: %d \t filesize: %d", size, fileSize)

		if fileSize > maxFileSize {
			return logError(status.Errorf(codes.InvalidArgument, "file is too large: %d > %d", fileSize, maxFileSize))
		}

		_, err = fileData.Write(chunk)
		if err != nil {
			return logError(status.Errorf(codes.Internal, "cannot write chunk data: %v", err))
		}
	}
	log.Printf("stream is finished")

	filePath := tempDir + "/tmp.zip"

	file, err := os.Create(filePath)
	if err != nil {
		return logError(fmt.Errorf("cannot create file: %w", err))
	}

	_, err = fileData.WriteTo(file)
	if err != nil {
		return logError(fmt.Errorf("cannot write to file: %w", err))
	}

	/*err = arch.UnzipSource(filePath, tempDir)
	if err != nil {
		log.Fatal(err)
	}

	c := make(chan event.Event)
	go convertThread(c)

	c <- event.Event{
		FilePath:    filePath,
		Mem:         0,
		Dur:         0,
		Start:       time.Now(),
		TempFolder:  tempDir,
		ArchiveSize: fileSize,
	}*/
	return nil
}

func convertThread(c chan event.Event) {
	ev := <-c
	conv.Convert(ev.TempFolder + "/index.html")

	os.RemoveAll(ev.TempFolder)
	ev.Dur = time.Since(ev.Start)
}

//contextError
/* retrievs an error from context, log and return*/
func contextError(ctx context.Context) error {
	switch ctx.Err() {
	case context.Canceled:
		return logError(status.Error(codes.Canceled, "request is canceled"))
	case context.DeadlineExceeded:
		return logError(status.Error(codes.DeadlineExceeded, "deadline is exceeded"))
	default:
		return nil
	}
}

// logError
/* logging an error that returns it*/
func logError(err error) error {
	if err != nil {
		log.Print(err)
	}
	return err
}
