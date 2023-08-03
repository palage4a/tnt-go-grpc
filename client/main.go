package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	pb "github.com/palage4a/tnt-go-grpc/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
)

func main() {
	flag.Parse()
	// Set up a connection to the server.
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewTntClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Work representation

	// Prepare data
	key := "SESSION_ID_1234"
	value := "USER_ID_1234"

	// Replace (or insert if it is not exists) a tuple
	log.Println("Calling 'Replace' procedure...")
	repl_resp, repl_err := c.Replace(ctx, &pb.ReplaceRequest{
		Key: key,
		Value: value,
		Timestamp: time.Now().Unix(),
	})
	if repl_err != nil {
		fmt.Print(repl_err)
	}
	log.Printf("Replace response: { %s, %s, %d}", repl_resp.GetKey(), repl_resp.GetValue(), repl_resp.GetTimestamp())

	// Get the inserted tuple
	log.Println("Calling 'Get' procedure...")
	get_resp, get_err := c.Get(ctx, &pb.GetRequest{
		Key: key,
	})
	if get_err != nil {
		fmt.Print(repl_err)
	}
	log.Printf("Get response: {%s, %s, %d }", get_resp.GetKey(), get_resp.GetValue(), get_resp.GetTimestamp())
}
