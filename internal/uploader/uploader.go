package uploader

import (
	"bytes"
	"context"
	"fmt"
	arch "grpc-html-to-pdf/internal/archiver"
	conv "grpc-html-to-pdf/internal/converter"
	pb "grpc-html-to-pdf/internal/uploader/proto"
	"io"
	"log"
	"os"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

/*

const (
	[ ]  maxFileSize = 1 << 20
)*/

// GRPCServer ...
type Server struct {
	pb.UnimplementedUploaderServer
}

// PingPong ...
func (s *Server) PingPong(ctx context.Context, req *pb.Ping) (*pb.Pong, error) {
	fmt.Println(req.Text)
	return &pb.Pong{Text: "pong!"}, nil
}

// Upload ...
func (s *Server) Upload(stream pb.Uploader_UploadServer) error {
	/*req, err := stream.Recv()
	if err != nil {
		return logError(status.Errorf(codes.Unknown, "cannot receive an info"))
	}*/

	fileData := bytes.Buffer{}
	//fileSize := 0
	for {
		err := contextError(stream.Context())
		if err != nil {
			return err
		}

		log.Print("waiting to receive more data")

		req, err := stream.Recv()
		if err == io.EOF {
			log.Print("no more data")
			break
		}
		if err != nil {
			return logError(status.Errorf(codes.Unknown, "cannot receive chunk data: %v", err))
		}

		chunk := req.GetChunk()
		size := len(chunk)

		log.Printf("received a chunk with size: %d", size)

		//[ ] Implement filesize comparison
		/*fileSize += size
		if fileSize > maxFileSize {
			return logError(status.Errorf(codes.InvalidArgument, "file is too large: %d > %d", fileSize, maxFileSize))
		}*/

		_, err = fileData.Write(chunk)
		if err != nil {
			return logError(status.Errorf(codes.Internal, "cannot write chunk data: %v", err))
		}
	}

	filePath := "tmp.zip"

	file, err := os.Create(filePath)
	if err != nil {
		return logError(fmt.Errorf("cannot create file: %w", err))
	}

	_, err = fileData.WriteTo(file)
	if err != nil {
		return logError(fmt.Errorf("cannot write to file: %w", err))
	}

	tempDir := "./tmp"
	if err := os.Mkdir(tempDir, 0750); err != nil {
		return logError(fmt.Errorf("cannot create temp directory: %w", err))
	}

	err = arch.UnzipSource(filePath, tempDir)
	if err != nil {
		log.Fatal(err)
	}

	fileName := "CURVES.html"

	conv.Convert(tempDir + "/" + fileName)

	os.RemoveAll(tempDir)
	os.Remove(filePath)
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
