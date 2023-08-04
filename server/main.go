package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"reflect"

	"google.golang.org/grpc"

	pb "github.com/palage4a/tnt-go-grpc/proto"
	tnt "github.com/tarantool/go-tarantool"
	crud "github.com/tarantool/go-tarantool/crud"
)

var (
	port = flag.Int("port", 50051, "The server port")
	tnt_space = flag.String("tntspace", "keyvalue", "The space used to operate")
	tnt_host = flag.String("tnthost", "127.0.0.1", "The tnt host")
	tnt_port = flag.Int("tntport", 3300, "The tnt port")
	tnt_user = flag.String("tntuser", "admin", "User for connect to tnt")
	tnt_passwd = flag.String("tntpasswd", "secret-cluster-cookie", "Password for connect to tnt")
)

type server struct {
	pb.UnimplementedTntServer
}

func getConnection() (*tnt.Connection, error) {
	log.Println("Connecting to tarantool...")
	// NOTE: admin credentials of "cartridge create" app
	opts := tnt.Opts{User: *tnt_user, Pass: *tnt_passwd}
	uri := fmt.Sprintf("%s:%d", *tnt_host, *tnt_port)
	return tnt.Connect(uri, opts)
}

func (s *server) Replace(c context.Context, req *pb.ReplaceRequest) (resp *pb.ReplaceResponse, err error) {
	conn, err := getConnection()
	defer conn.Close()
	if err != nil {
		return nil, fmt.Errorf("Error: %s", err)
	}
	log.Println("Calling 'crud.replace(<space_name>, ...)'...")
	res := crud.MakeResult(reflect.TypeOf(&pb.ReplaceResponse{}))
	err = conn.Do(crud.MakeReplaceRequest(*tnt_space).
		Opts(crud.SimpleOperationOpts{
			Fields: crud.MakeOptTuple(
				[]interface{}{"key", "value", "timestamp", "meta"},
			),
		}).
		Tuple([]crud.Tuple{
			req.GetKey(),
			nil,
			req.GetValue(),
			req.GetTimestamp(),
			req.GetMeta(),
		}),
	).GetTyped(&res)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	rows := res.Rows.([]*pb.ReplaceResponse)
	if len(rows) != 0 {
		return rows[0], nil
	}
	return nil, fmt.Errorf("unknown error: replace call return nothing")
}

func (s *server) Get(c context.Context, req *pb.GetRequest) (resp *pb.GetResponse, err error) {
	conn, err := getConnection()
	defer conn.Close()
	if err != nil {
		return nil, fmt.Errorf("Error: %s", err)
	}
	log.Println("Calling 'crud.select(<space_name>,..)'...")
	res := crud.MakeResult(reflect.TypeOf(&pb.GetResponse{}))
	conditions := []crud.Condition{
		{Operator: crud.Eq, Field: "key", Value: req.GetKey()},
	}
	err = conn.Do(crud.MakeSelectRequest(*tnt_space).
		Conditions(conditions).
		Opts(crud.SelectOpts{
			First: crud.MakeOptInt(1),
			Fields: crud.MakeOptTuple(
				[]interface{}{"key", "value", "timestamp", "meta"},
			),
		}),
	).GetTyped(&res)

	if err != nil {
		log.Println(err)
		return nil, err
	}
	rows := res.Rows.([]*pb.GetResponse)
	if len(rows) > 0 {
		return rows[0], nil
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
