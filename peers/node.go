// how to call the peer to peer network
// have one terminal connect to
// go run peers/node.go :5000 localhost:5001
// have another terminal connect to
// go run peers/node.go :5001 localhost:5000
// the peers will then connect

package main

import (
    "context"
    "log"
    "net"
    "os"
    "strconv"
    "time"

    hs "github.com/Alex-itu/Consensus_gRPC/proto"
    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"
)

// Node structure
type Node struct {
    hs.UnimplementedHelloServiceServer
    ID     int
    Client hs.HelloServiceClient
}

// SayHello is the RPC method that implements helloworld.GreeterServer
func (n *Node) SayHello(ctx context.Context, in *hs.HelloRequest) (*hs.HelloReply, error) {
    return &hs.HelloReply{Message: "Hello " + strconv.Itoa(n.ID)}, nil
}

func main() {
    args := os.Args[1:]

    // example arg[0] -> :5000
    port := args[0]
    otherNodeAddress := args[1]

    server, lis, err := createServer(port)
    if err != nil {
        log.Fatalf("failed to create server: %v", err)
    }

    node := &Node{ID: 42, Client: nil}
    hs.RegisterHelloServiceServer(server, node)
    reflection.Register(server)

    startServer(server, lis)

    // wait for other nodes to be ready
    time.Sleep(30 * time.Second)

    // setup connection with other node
    err = connectToOtherNode(node, otherNodeAddress)
    if err != nil {
        log.Fatalf("could not connect or greet the other node: %v", err)
    }

    for {
        time.Sleep(10 * time.Second)
    }
}

// createServer creates a gRPC server and returns it along with its listener
func createServer(port string) (*grpc.Server, net.Listener, error) {
    lis, err := net.Listen("tcp", port)
    if err != nil {
        return nil, nil, err
    }
    return grpc.NewServer(), lis, nil
}

// startServer starts the passed in gRPC server
func startServer(server *grpc.Server, lis net.Listener) {
    go func() {
        if err := server.Serve(lis); err != nil {
            log.Fatalf("failed to serve: %v", err)
        }
    }()
}

// connectToOtherNode establishes a connection with the other node and performs a greeting
func connectToOtherNode(node *Node, address string) error {
    conn, err := grpc.Dial(address, grpc.WithInsecure())
    if err != nil {
        return err
    }
    defer conn.Close()

    node.Client = hs.NewHelloServiceClient(conn)

    r, err := node.Client.SayHello(context.Background(), &hs.HelloRequest{Name: "John"})
    if err != nil {
        return err
    }

    log.Printf("Greeting from the other node: %s", r.Message)
    return nil
}
