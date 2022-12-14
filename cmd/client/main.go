package main

import (
	"bufio"
	"context"
	"flag"
	"io"
	"log"
	"os"

	pb "grpc-html-to-pdf/internal/uploader/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
)

func main() {
	var filePath string
	flag.StringVar(&filePath, "file", "", "pointing out the path where the original file stores")
	flag.Parse()

	//fmt.Println(os.Getwd())
	log.Printf("dial server %s", *addr)

	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("cannot dial server: ", err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal("cannot open a file: ", err)
	}
	defer file.Close()

	client := pb.NewUploaderClient(conn)
	stream, err := client.Upload(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("cannot read chunk to buffer: ", err)
		}

		req := &pb.UploadRequest{
			Chunk: buffer[:n],
		}

		err = stream.Send(req)
		if err != nil {
			log.Fatal("cannot send chunk to server: ", err, stream.RecvMsg(nil))
		}
	}

	eventUUID, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal("cannot receive response: ", err)
	}
	log.Println(eventUUID)
}
