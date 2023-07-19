package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"google.golang.org/grpc"

	tnt "github.com/tarantool/go-tarantool"
	pb "github.com/palage4a/tnt-go-grpc/tnt"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

type server struct {
	pb.UnimplementedTntServer
}

func (s *server) SayHello(c context.Context,in *pb.Person) (*pb.Greeting, error) {
	opts := tnt.Opts{User: "guest"}
	conn, err := tnt.Connect("127.0.0.1:3301", opts)
	if err != nil {
		fmt.Println("Connection refused:", err)
	}

	resp, err := conn.Eval("name = ... return 'Hello from tnt to ' .. name", []interface{}{in.GetName()})
	if err != nil {
		return nil, fmt.Errorf("Error: ", err)
	}

	res := fmt.Sprintf("%s", resp.Data)
	greeting := &pb.Greeting{Message: res}
	return greeting, nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterTntServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
