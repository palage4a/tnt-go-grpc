package main

import (
	"fmt"
	tnt "github.com/tarantool/go-tarantool"
)

func main() {
	fmt.Println("Hello, world")
	opts := tnt.Opts{User: "guest"}
	conn, err := tnt.Connect("127.0.0.1:3301", opts)
	if err != nil {
		fmt.Println("Connection refused:", err)
	}
	resp, err := conn.Eval("return 'Hello world from tnt'", []interface{}{})
	if err != nil {
		fmt.Println("Error", err)
		fmt.Println("Code", resp.Code)
	}
	fmt.Println(resp.Data)
}
