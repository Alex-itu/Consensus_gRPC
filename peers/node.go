// how to call the peer to peer network
// have one terminal connect to
// go run peers/node.go :5000 localhost:5001
// have another terminal connect to
// go run peers/node.go :5001 localhost:5000
// the peers will then connect

//note, each node can act as both a client and a server.

package main

import (
	"context"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
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
	conn   *grpc.ClientConn
}

// SayHello is the RPC method that implements helloworld.GreeterServer
func (n *Node) SayHello(stream hs.HelloService_SayHelloServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		reply := &hs.HelloReply{Message: "Hello " + req.Name + " from Node " + strconv.Itoa(n.ID)}
		if err := stream.Send(reply); err != nil {
			return err
		}
	}
}

func main() {
	args := os.Args[1:]

	port := args[0]
	otherNodeAddress := args[1]

	serverNode, lis, err := createServerNode(port)
	if err != nil {
		log.Fatalf("failed to create server node: %v", err)
	}

	node := &Node{ID: 42, Client: nil}
	hs.RegisterHelloServiceServer(serverNode, node)
	reflection.Register(serverNode)

	go startServerNode(serverNode, lis)

	// wait for other nodes to be ready
	time.Sleep(30 * time.Second)

	// setup connection with other node
	err = connectToOtherNode(node, otherNodeAddress)
	if err != nil {
		log.Fatalf("could not connect or greet the other node: %v", err)
	}

	// wait for termination signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	// safely stop the node and close the connection
	serverNode.GracefulStop()
	node.conn.Close()
}

func createServerNode(port string) (*grpc.Server, net.Listener, error) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		return nil, nil, err
	}
	return grpc.NewServer(), lis, nil
}

func startServerNode(server *grpc.Server, lis net.Listener) {
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func connectToOtherNode(node *Node, address string) error {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return err
	}

	node.conn = conn
	node.Client = hs.NewHelloServiceClient(conn)

	stream, err := node.Client.SayHello(context.Background())
	if err != nil {
		return err
	}

	err = stream.Send(&hs.HelloRequest{Name: "John"})
	if err != nil {
		log.Printf("failed to send request: %v", err)
		return err
	}

	reply, err := stream.Recv()
	if err != nil {
		log.Printf("failed to receive reply: %v", err)
		return err
	}

	log.Printf("Greeting from the other node: %s", reply.Message)
	return nil
}