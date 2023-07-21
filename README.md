# tnt-go-grpc

## Run

### setup tarantool
You should start a tarantool instance and run this code on it:
```sh
cd tnt;
tarantool init.lua;
```

### run server
```sh
go run server/main.ru
```

### run client
```sh
go run client/main.ru
```

## generation
```sh
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/tnt.proto
```
