package main

import (
	"log"
	"net"

	"grpc-html-to-pdf/internal/uploader"

	pb "grpc-html-to-pdf/internal/uploader/proto"

	"google.golang.org/grpc"
)

func main() {
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
