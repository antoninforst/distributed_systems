# Ricart--Agrawala Distributed Mutual Exclusion
Implementation of the Ricart--Agrawala distributed mutual exclusion
algorithm in Go, using gRPC for communication between three
independent nodes.

Each node acts both as a gRPC server and as a client of the other two
nodes. A node can enter the simulated critical section only after
receiving permission from both peers.

## 1. Requirements
- Go 1.27.0 or later.
- The three nodes must be able to communicate over TCP.
- The project dependencies listed in go.mod.
- The generated Protocol Buffers and gRPC Go files are included in the
repository, so protoc is not required to run the project.

## 2. Project structure
The project should have the following structure:
```text
ricart-agrawala/
├── go.mod
├── go.sum
├── main.go
├── node.go
└── proto/
    ├── mutual_exclusion.proto
    ├── mutual_exclusion.pb.go
    └── mutual_exclusion_grpc.pb.go
```
Keep the generated files inside the proto/ directory. The Go source
imports them using ricart-agrawala/proto.

## 3. Node configuration
The three node addresses are configured in main.go:

Node Address Port

A localhost:50051 50051
B localhost:50052 50052
C localhost:50053 50053

No external service-discovery system is required: each process uses this
fixed peer configuration.

## 4. Build and run
Open three separate terminals in the project root directory.

First, start Node A:

go run -race . -id=A
Then, in the second terminal, start Node B:

go run -race . -id=B
Finally, in the third terminal, start Node C:

go run -race . -id=C
Start the three processes close together in time. Each process waits
five seconds after it becomes ready and then automatically requests
access to the critical section. Starting them close together makes their
requests more likely to overlap and allows the deferred-reply behaviour
to be observed.

The -race flag enables Go's race detector. It is useful during
testing; it is not required for a normal run. To run without it, use
go run . -id=A (and likewise for B and C).

Keep all three terminals and processes running. Each
node needs the other two nodes to answer its permission request. The
program keeps its gRPC server alive after its own critical section
execution so it can continue handling incoming requests.

## 5. How the algorithm works
Each node maintains a Lamport logical clock and its local request state.

When a node requests the critical section, it increments its Lamport
clock, records the request timestamp and sends a REQUEST to both
peers.
Each receiving node updates its Lamport clock and decides whether to
reply immediately or defer its reply.
A node defers a reply if it is currently inside the critical
section. If it is also requesting access, it defers the incoming
request when its own request has higher priority.
Request priority is determined lexicographically by
(Lamport timestamp, node ID). The lower timestamp has priority; if
timestamps are equal, the alphabetically lower node ID has priority.
The requester enters the critical section only after receiving
replies from both other nodes.
When it exits, it releases all replies that it had deferred,
allowing waiting nodes to proceed.
The critical section is simulated by a two-second delay. The program
logs requests, granted or deferred replies, entry and exit events.

## 6. Expected output
The exact order depends on process startup timing and message delivery.
Log messages should include events similar to:

Node A is ready
Node A requests the critical section (timestamp: 1)
Node B defers REPLY to Node A
Node C grants permission to Node A
Node A received REPLY from Node C
Node B sends deferred REPLY to Node A
Node A ENTERS the critical section
Node A is executing the critical section
Node A EXITS the critical section
This is an example, not a transcript. The order and timestamps may differ.

## 7. Testing and evidence
For the demonstration:

Run all three nodes in separate terminals using the commands above.
Keep the complete output from all three terminals.
Check that each requesting node receives permissions from both peers
before entering.
Check that a node inside the critical section defers incoming
requests.
Check that deferred replies are sent after the node exits.
Check that the ENTER and EXIT intervals do not overlap.
When using -race, check whether the output contains
WARNING: DATA RACE.
Include genuine terminal logs in the project report. A written summary
of the events is useful for explaining the test, but it is not a
substitute for the original logs. A run without a race-detector warning
is evidence for that run only; it is not a proof that no data race can
occur under every possible schedule.

## 8. Assumptions and limitations
The configuration contains exactly three nodes: A, B and C.
The addresses are hardcoded in main.go; changing the deployment
addresses requires updating that configuration.
The implementation assumes that all nodes remain running and that
messages are eventually delivered.
It does not implement crash recovery or fault tolerance. If a peer
stops responding, a permission request may fail or time out.
The current main.go performs one automatic critical-section
request per process. It does not provide an interactive loop for
repeated requests.

## 9. Protocol definition
The gRPC service is defined in proto/mutual_exclusion.proto. It
exposes one RPC:
rpc RequestPermission (Request) returns (Reply);
The Request and Reply messages contain the node ID and Lamport timestamp.

rpc RequestPermission (Request) returns (Reply);
The Request and Reply messages contain the node ID and Lamport
timestamp.



