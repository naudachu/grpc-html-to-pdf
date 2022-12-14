package main

import (
	"log"
	"net"

	"grpc-html-to-pdf/internal/uploader"

	pb "grpc-html-to-pdf/internal/uploader/proto"

	pdf "github.com/adrg/go-wkhtmltopdf"
	"google.golang.org/grpc"
)

func main() {
	if err := pdf.Init(); err != nil {
		log.Fatal(err)
	}
	defer pdf.Destroy()

	s := grpc.NewServer()
	srv := uploader.NewUploaderService()
	pb.RegisterUploaderServer(s, srv)

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		srv.RunBackground()
	}()

	if err := s.Serve(listener); err != nil {
		log.Fatal(err)
	}
}
