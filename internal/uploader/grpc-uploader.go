package uploader

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	arch "grpc-html-to-pdf/internal/archiver"
	conv "grpc-html-to-pdf/internal/converter"
	"grpc-html-to-pdf/internal/event"
	pb "grpc-html-to-pdf/internal/uploader/proto"
	"io"
	"log"
	"os"

	"github.com/google/uuid"
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
	pool  *event.Pool
}

func NewUploaderService() *Server {
	var tasks []*event.Task
	return &Server{
		UnimplementedUploaderServer: pb.UnimplementedUploaderServer{},
		tasks:                       tasks,
		pool:                        event.NewPool(tasks, 4),
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
	return tempDir, nil
}

// Upload ...
func (s *Server) Upload(stream pb.Uploader_UploadServer) error {
	e := &event.Event{
		UUID: uuid.New(),
	}

	tempDir, err := tempDirectoriesPreparation()
	if err != nil {
		logError(fmt.Errorf("cannot create temp directory: %w", err))
	}

	fileData := bytes.Buffer{}
	var fileSize int64
	fileSize = 0
	// Reading the stream
	for {

		//log.Print("waiting to receive more data")

		req, err := stream.Recv()
		if err == io.EOF {
			log.Print("no more data")
			break
		}
		if err != nil {
			return logError(status.Errorf(codes.Unknown, "cannot receive chunk data: %v", err))
		}

		chunk := req.GetChunk()
		size := int64(len(chunk))
		fileSize += size

		//log.Printf("new chunk: %d \t filesize: %d", size, fileSize)

		if fileSize > maxFileSize {
			return logError(status.Errorf(codes.InvalidArgument, "file is too large: %d > %d", fileSize, maxFileSize))
		}

		_, err = fileData.Write(chunk)
		if err != nil {
			return logError(status.Errorf(codes.Internal, "cannot write chunk data: %v", err))
		}
	}
	stream.SendAndClose(&pb.UploadResponse{
		Answer: e.UUID.String(),
	})

	filePath := tempDir + "/tmp.zip"

	file, err := os.Create(filePath)
	if err != nil {
		return logError(fmt.Errorf("cannot create file: %w", err))
	}

	_, err = fileData.WriteTo(file)
	if err != nil {
		return logError(fmt.Errorf("cannot write to file: %w", err))
	}

	e.PostUpload(filePath, tempDir, fileSize)
	log.Print("work done: " + e.String())

	/*t := event.NewTask(func(i interface{}) error {
		err := processConvertion(i.(*event.Event))
		if err != nil {
			logError(err)
		}
		return nil
	}, e)

	s.pool.AddTask(t)*/

	t := event.NewTask(func(i interface{}) error {
		conv.CountTillFifty(i.(*event.Event))
		return nil
	}, e)

	s.pool.AddTask(t)

	return nil
}

// processConvertion
/* describes the process of unzip and convert zip archive*/
func processConvertion(e *event.Event) error {
	err := arch.UnzipSource(e.FilePath, e.TempFolder)
	if err != nil {
		log.Fatal(err)
	}

	//if err = conv.ConvertADRG(e); err != nil {
	if err = conv.PDFg(e); err != nil {
		logError(err)
	}

	os.RemoveAll(e.TempFolder)

	return nil
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
