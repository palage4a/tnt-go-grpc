# tnt-go-grpc

## Run

### setup tarantool
You should start a tarantool instance and run this code on it:
```lua
box.cfg{listen=3301}
box.schema.user.grant("guest", "execute", "universe")
```

### run applications
```sh
go run server/main.ru
```

```sh
go run client/main.ru
```

## generation
```sh
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    simple.proto
```
