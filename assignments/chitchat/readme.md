# Chit Chat

From `assignments/chitchat`, regenerate the Go files after editing the proto:

```sh
cd grpc
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative chitchat.proto
cd ..
```

Start the server:

```sh
go run ./server
```

In another terminal, start a client. Run more clients in separate terminals to see join and leave notices:

```sh
go run ./client
```

Enter a name to join, then press Enter to leave.
