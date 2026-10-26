# Chit Chat

## Prerequisites

Make sure Go and the Protocol Buffers compiler (`protoc`) are installed. The project uses gRPC and Protocol Buffers, with generated Go files stored in `grpc/`.

## Build the project

Run these commands from the project root (the directory containing `go.mod`):

```sh
go build ./...
```

If the build completes without errors, the project compiles successfully.

## Regenerate the Go files after editing the proto

Only do this if `grpc/chitchat.proto` has changed. Make sure the Go Protocol Buffers and gRPC code-generation plugins are installed and available on your `PATH`.

From the project root, run:

```sh
cd grpc
protoc --proto_path=. \
  --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  chitchat.proto
cd ..
```

## Run the server

From the project root, open a terminal and run:

```sh
go run ./server
```

The server listens on `localhost:5050`. Keep this terminal open while using the chat.

## Run the clients

Open a separate terminal for each participant. In each terminal, run this command from the project root:

```sh
go run ./client
```

Enter a display name when prompted. To satisfy the assignment's concurrent-client scenario, run at least three clients in separate terminals while the server is running.

## Use the chat

- Enter your name when prompted to join the chat.
- Type a message and press Enter to send it.
- Messages are broadcast to the active participants and displayed with their logical timestamps.
- Messages containing only whitespace are ignored by the client.
- Messages longer than 128 characters are rejected by the client with an explanatory message.
- Type `/quit` and press Enter to leave the chat. The remaining participants receive a leave notice.
- If a client disconnects unexpectedly, the server detects the failed stream, removes the participant, and broadcasts a leave notice to the remaining participants.

## RPC overview

The `ChitChatService` exposes three RPC methods:

- `Join`: server-side streaming RPC. Registers a participant and keeps a stream open to deliver chat events.
- `Publish`: unary RPC. Sends a message to the server, which assigns a logical timestamp and broadcasts it to active participants.
- `Leave`: unary RPC. Removes the participant and broadcasts a leave notice to the remaining participants.

## Testing checklist

For a basic manual test:

1. Start the server.
2. Start three clients in separate terminals and give each a different name.
3. Send a message from one client and confirm the active clients receive it with the same logical timestamp.
4. Type `/quit` in one client and confirm the others receive the leave notice.
5. Send another message from a remaining client to confirm the chat continues to work.
6. Try entering a message longer than 128 characters and confirm the client rejects it.
7. Optionally, close a client terminal without typing `/quit` and confirm the server logs the disconnection and the remaining clients receive a leave notice.

The server and client use Go's standard `log` package to record events with timestamps and fields such as component, event, client ID, and logical time.

## Notes

- Run commands from the project root unless a step explicitly says to change directory.
- Keep the server running in one terminal and each client in its own terminal.
- If you edit the `.proto` file, regenerate the Go files and then run `go build ./...` again.
