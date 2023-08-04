package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	pb "github.com/palage4a/tnt-go-grpc/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
	port = flag.Int("port", 8080, "http port")
	key = flag.String("key", "key", "key")
	value = flag.String("value", "value", "value")
	httpMode = flag.Bool("http", false, "start http server")
	c pb.TntClient
)

func demo() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 10)
	defer cancel()

	log.Println("Calling 'Replace' procedure...")
	repl_resp, repl_err := c.Replace(ctx, &pb.ReplaceRequest{
		Key: *key,
		Value: *value,
		Timestamp: time.Now().Unix(),
	})
	if repl_err != nil {
		fmt.Print(repl_err)
	}
	log.Printf("Replace response: { %s, %s, %d}", repl_resp.GetKey(), repl_resp.GetValue(), repl_resp.GetTimestamp())

	log.Println("Calling 'Get' procedure...")
	get_resp, get_err := c.Get(ctx, &pb.GetRequest{
		Key: *key,
	})
	if get_err != nil {
		fmt.Print(repl_err)
	}
	log.Printf("Get response: {%s, %s, %d }", get_resp.GetKey(), get_resp.GetValue(), get_resp.GetTimestamp())

	log.Println("Setup bidirectional RPC call")
	waitAll := make(chan struct{}, 2)
	go func() {
		stream, err := c.GetStream(ctx)
		if err != nil {
			log.Fatal(err)
		}
		waitc := make(chan struct{})
		go func() {
			for {
				in, err := stream.Recv()
				if err == io.EOF {
					close(waitc)
					return
				}
				if err != nil {
					log.Println("get recv")
					log.Fatal(err)
				}
				log.Printf("Get response: {%s, %s, %d }", in.GetKey(), in.GetValue(), in.GetTimestamp())
			}
		}()
		tuples := []*pb.GetRequest{
			{Key: "1"},
			{Key: "2"},
			{Key: "3"},
			{Key: "4"},
		}
		for _, t := range tuples {
			if err := stream.Send(t); err != nil {
				log.Println("get send")
				log.Fatal(err)
			}
		}
		stream.CloseSend()
		<-waitc
		waitAll <- struct{}{}
	}()

	go func() {
		stream, err := c.ReplaceStream(ctx)
		if err != nil {
			log.Fatal(err)
		}
		waitc := make(chan struct{})
		go func() {
			for {
				in, err := stream.Recv()
				if err == io.EOF {
					close(waitc)
					return
				}
				if err != nil {
					log.Println("replace recv")
					log.Fatal(err)
				}
				log.Printf("Replace response: {%s, %s, %d }", in.GetKey(), in.GetValue(), in.GetTimestamp())
			}
		}()
		tuples := []*pb.ReplaceRequest{
			{Key: "1", Value: "value1"},
			{Key: "2", Value: "value2"},
			{Key: "3", Value: "value3"},
			{Key: "4", Value: "value4"},
		}
		for _, t := range tuples {
			if err := stream.Send(t); err != nil {
				log.Println("replace send")
				log.Fatal(err)
			}
			time.Sleep(200 * time.Millisecond)
		}
		stream.CloseSend()
		<-waitc
		waitAll <- struct{}{}
	}()

	<-waitAll
	<-waitAll
}

func replace(rw http.ResponseWriter, rq *http.Request) {
	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Replace (or insert if it is not exists) a tuple
	log.Println("Calling 'Replace' procedure...")
	key_form := rq.FormValue("key")
	if key_form == "" {
		key_form = *key
	}
	value_form := rq.FormValue("value")
	if value_form == "" {
		value_form = *value
	}
	meta_form := rq.FormValue("meta")
	resp, err := c.Replace(ctx, &pb.ReplaceRequest{
		Key: key_form,
		Value: value_form,
		Timestamp: time.Now().Unix(),
		Meta: &meta_form,
	})
	if err != nil {
		http.Error(rw, err.Error(), 500)
	}
	log.Printf("Replace response: { %s, %s, %d, %s}", resp.GetKey(), resp.GetValue(), resp.GetTimestamp(), resp.GetMeta())

	bts, err := json.Marshal(resp)
	if err != nil {
		http.Error(rw, err.Error(), 500)
	}

	fmt.Fprint(rw, string(bts))
}
func get(rw http.ResponseWriter, rq *http.Request) {
	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Get a tuple
	log.Println("Calling 'Get' procedure...")
	key_query := rq.FormValue("key")
	if key_query == "" {
		key_query = *key
	}
	resp, err := c.Get(ctx, &pb.GetRequest{
		Key: key_query,
	})
	if err != nil {
		http.Error(rw, err.Error(), 500)
	}
	log.Printf("Get response: {%s, %s, %d, %s}", resp.GetKey(), resp.GetValue(), resp.GetTimestamp(), resp.GetMeta())
	bts, err := json.Marshal(resp)
	if err != nil {
		http.Error(rw, err.Error(), 500)
	}

	fmt.Fprint(rw, string(bts))
}

func serveHttp(c pb.TntClient) {
	http.HandleFunc("/replace", replace)
	http.HandleFunc("/get", get)
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("starting http server on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func main() {
	flag.Parse()
	// Set up a connection to the server.
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c = pb.NewTntClient(conn)

	if *httpMode != false {
		serveHttp(c)
	}
	log.Println("Staring demo")
	demo()
	return
}
