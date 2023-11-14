// how to call the peer to peer network

// go run .\node.go -name alice -port 5000 -portfor localhost:5010 -portback localhost:5020
// go run .\node.go -name bob -port 5010 -portfor localhost:5020 -portback localhost:5000
// go run .\node.go -name charlie -port 5020 -portfor localhost:5000 -portback localhost:5010

// Inspiration taken from Github: https://github.com/JonasSkjodt/chitty-chat

package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"time"

	hs "github.com/Alex-itu/Consensus_gRPC/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Node structure
type Node struct {
	hs.UnimplementedTokenServiceServer
	name           string
	Port           string
	Client         hs.TokenServiceClient
	Clientforward  hs.TokenServiceClient
	Clientbackward hs.TokenServiceClient
}

var clientservertokenstream hs.TokenServiceClient
var forwardclienttokenstream hs.TokenServiceClient
var backwardclienttokenstream hs.TokenServiceClient

var clientServerTokenStream hs.TokenService_TokenChatClient
var forwardClientTokenStream hs.TokenService_TokenChatClient
var backwardClientTokenStream hs.TokenService_TokenChatClient

var nodeServerconn *grpc.ClientConn

var token bool

var name = flag.String("name", "John", "The name for the node")
var connPort = flag.String("port", "8080", "port to another node")
var connPortforward = flag.String("portfor", "8080", "port to another node")
var conPortbackward = flag.String("portback", "8080", "port to another node")
var tokenflag = flag.Bool("hasToken", false, "Is the token")

func main() {
	flag.Parse()
	token = *tokenflag

	f := setLog() //uncomment this line to log to a log.txt file instead of the console
	defer f.Close()

	node := &Node{
		name:           *name,
		Port:           *connPort,
		Client:         nil,
		Clientforward:  nil,
		Clientbackward: nil,
	}

	fmt.Printf("nodeID: " + node.name + " and port to connect to: " + node.Port + " \n \n")

	// start server
	list, err := net.Listen("tcp", fmt.Sprintf(":%s", node.Port))
	if err != nil {
		fmt.Printf("Server : Failed to listen on port : %v \n", err)
		return
	}

	go func() {
		var opt []grpc.ServerOption
		clientServer := grpc.NewServer(opt...)

		hs.RegisterTokenServiceServer(clientServer, node)

		if err := clientServer.Serve(list); err != nil {
			fmt.Printf("failed to serve %v", err)
		}
	}()

	time.Sleep(1 * time.Second)

	// create conn
	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	conn, err := grpc.Dial(fmt.Sprintf(":%s", node.Port), opts...)
	if err != nil {
		fmt.Printf("failed on Dial: %v", err)
	}

	fmt.Printf("Dialing the server from client. \n \n")

	node.Client = hs.NewTokenServiceClient(conn)
	nodeServerconn = conn

	defer nodeServerconn.Close()

	connforward, err := grpc.Dial(*connPortforward, opts...)
	if err != nil {
		fmt.Printf("whats the error: %v\n", err)
	}

	fmt.Println("Successfully connected to forward node. \n \n")

	// delete if not used
	node.Clientforward = hs.NewTokenServiceClient(connforward)

	//need to check for if nodeID is 10, since the backward node is 30 or the highst id
	connBackward, err := grpc.Dial(*conPortbackward, opts...)
	if err != nil {
		fmt.Printf("Failed to connect to backward node: %v\n", err)
		connforward.Close()
	}
	fmt.Println("Successfully connected to backward node. \n \n")

	node.Clientbackward = hs.NewTokenServiceClient(connBackward)

	//connectoOtherNode end

	if err != nil {
		fmt.Printf("Error on connecting to other nodes: %v \n", err)
		log.Fatalf("failed to connect to other nodes: %v", err)
	}

	clientServerTokenStream, err := node.Client.TokenChat(context.Background())
	if err != nil {
		fmt.Printf("Error on receive: %v \n", err)
	}

	if node.Clientforward != nil {
		forwardClientTokenStream, err = node.Clientforward.TokenChat(context.Background())
		if err != nil {
			fmt.Printf("Error while establishing forward token chat: %v \n", err)
			return
		}
	} else {
		fmt.Printf("Error: node.Clientforward is nil\n")
		return
	}

	backwardClientTokenStream, err := node.Clientbackward.TokenChat(context.Background())
	if err != nil {
		fmt.Printf("Error while establishing backward token chat: %v \n", err)
		return
	}

	defer connforward.Close()
	defer connBackward.Close()

	ListenInternal(clientServerTokenStream, forwardClientTokenStream, backwardClientTokenStream)
}

func createClientServerConn(node Node) {

	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.Dial(node.Port, opts...)
	if err != nil {
		fmt.Printf("failed on Dial: %v", err)
	}

	fmt.Printf("Dialing the server from client")

	node.Client = hs.NewTokenServiceClient(conn)
	nodeServerconn = conn
}

// connectToOtherNode establishes a connection with the other node and performs a greeting
func connectToOtherNode(node Node) (*grpc.ClientConn, *grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	connforward, err := grpc.Dial(*connPortforward, opts...)
	if err != nil {
		fmt.Printf("whats the error: %v\n", err)
		return nil, nil, err

	}

	fmt.Println("Successfully connected to forward node. \n \n")

	node.Clientforward = hs.NewTokenServiceClient(connforward)

	//need to check for if nodeID is 10, since the backward node is 30 or the highst id
	connBackward, err := grpc.Dial(*conPortbackward, opts...)
	if err != nil {
		fmt.Printf("Failed to connect to backward node: %v\n", err)
		connforward.Close()
		return nil, nil, err
	}
	fmt.Println("Successfully connected to backward node. \n \n")

	node.Clientbackward = hs.NewTokenServiceClient(connBackward)

	return connforward, connBackward, nil
}

// this is for server listening
func (s *Node) TokenChat(msgStream hs.TokenService_TokenChatServer) error {
	// get the next message from the stream
	for {
		msg, err := msgStream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		if msg.Token == "token" {
			token = true
			//clientServerTokenStream.Send(&hs.TokenRequest{Token: "token"})
		}
	}

	return nil
}

// this is for client listening
func ListenInternal(stream hs.TokenService_TokenChatClient, forwardStream hs.TokenService_TokenChatClient, backwardStream hs.TokenService_TokenChatClient) {
	for {
		time.Sleep(1 * time.Second)
		if token {
			fmt.Printf("%s recieved token\n", *name)
			log.Printf("%s recieved token", *name)

			fmt.Printf("%s is writing to the critical section\n", *name)
			log.Printf("%s is writing to the critical section", *name)

			fmt.Printf("%s is sending the token to the next node\n", *name)
			log.Printf("%s is sending the token to the next node", *name)

			token = false
			forwardStream.Send(&hs.TokenRequest{Token: "token"})
		}
	}
}

// sets the logger to use a log.txt file instead of the console
func setLog() *os.File {
	// Clears the log.txt file when a new server is started
	if err := os.Truncate("log.txt", 0); err != nil {
		fmt.Printf("Failed to truncate: %v \n", err)
		log.Printf("Failed to truncate: %v", err)
	}

	// This connects to the log file/changes the output of the log informaiton to the log.txt file.
	f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("error opening file: %v", err)
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
	return f

}
