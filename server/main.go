package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	tnt "github.com/tarantool/go-tarantool"
	pb "github.com/palage4a/tnt-go-grpc/proto"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

type server struct {
	pb.UnimplementedTntServer
}

func getConnection() (*tnt.Connection, error) {
	log.Println("Connecting to tarantool...")
	opts := tnt.Opts{User: "operator", Pass: "operator_pass"}
	return tnt.Connect("127.0.0.1:3301", opts)
}

func (s *server) Replace(c context.Context, req *pb.ReplaceRequest) (resp *pb.ReplaceResponse, err error) {
	conn, err := getConnection()
	defer conn.Close()
	if err != nil {
		return nil, fmt.Errorf("Error: %s", err)
	}

	log.Println("Calling 'box.space.<space_name>:replace'...")
	var res []*pb.ReplaceResponse
	err = conn.Do(tnt.NewReplaceRequest("keyvalue").
		Tuple([]interface{}{
			req.GetKey(),
			req.GetValue(),
			req.GetTimestamp(),
			req.GetMeta(),
		},
		),
	).GetTyped(&res)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	if len(res) > 0 {
		return res[0], nil
	}

	return nil, fmt.Errorf("unknown error: replace call return nothing")
}

func (s *server) Get(c context.Context, req *pb.GetRequest) (resp *pb.GetResponse, err error) {
	conn, err := getConnection()
	defer conn.Close()
	if err != nil {
		return nil, fmt.Errorf("Error: %s", err)
	}

	log.Println("Calling 'box.space.<space_name>:select'...")
	var res []*pb.GetResponse
	err = conn.Do(tnt.NewSelectRequest("keyvalue").
		Iterator(tnt.IterEq).
		Limit(1).
		Index("primary").
		Key(tnt.StringKey{S: req.GetKey()}),
	).GetTyped(&res)

	if err != nil {
		log.Println(err)
		return nil, err
	}
	if len(res) > 0 {
		return res[0], nil
	}

	return nil, fmt.Errorf("tuple with key %s is not found", req.GetKey())
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
