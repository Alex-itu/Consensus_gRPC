// how to call the peer to peer network
// total mess
// go run .\node.go -name alice -port 5000 -portfor 5010 -portback 5010
// go run .\node.go -name bob -port 5010 -portfor 5000 -portback 5000
// go run .\node.go -name charlie -port 5020 -portfor 5000 -portback 5010

// go run .\node.go -name alice -port 5000 -portfor localhost:5010 -portback localhost:5020
// go run .\node.go -name bob -port 5010 -portfor localhost:5020 -portback localhost:5000
// go run .\node.go -name charlie -port 5020 -portfor localhost:5000 -portback localhost:5010

package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"strings"
	"time"

	hs "github.com/Alex-itu/Consensus_gRPC/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	//"google.golang.org/grpc/reflection"
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

	fmt.Printf("nodeID: " + node.name + " and port to connect to: " + node.Port + " \n")

	// start server
	list, err := net.Listen("tcp", fmt.Sprintf(":%s", node.Port))
	if err != nil {
		fmt.Printf("Server : Failed to listen on port : %v \n", err)
		return
	}

	go func() {
		var opts []grpc.ServerOption
		clientServer := grpc.NewServer(opts...)

		hs.RegisterTokenServiceServer(clientServer, node)

		if err := clientServer.Serve(list); err != nil {
			fmt.Printf("failed to serve %v", err)
		}
	}()
	// started server
	//createServer(*node)

	time.Sleep(10 * time.Second)

	// create conn
	optst := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	conn, err := grpc.Dial(fmt.Sprintf(":%s", node.Port), optst...)
	if err != nil {
		fmt.Printf("failed on Dial: %v", err)
	}

	fmt.Printf("Dialing the server from client. \n")

	node.Client = hs.NewTokenServiceClient(conn)
	nodeServerconn = conn
	//createClientServerConn(*node)
	// created conn

	defer nodeServerconn.Close()

	connforward, connBackward, err := connectToOtherNode(*node)
	if err != nil {
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

	node.Clientforward = hs.NewTokenServiceClient(connforward)
	node.Clientbackward = hs.NewTokenServiceClient(connBackward)

	go ListenInternal(clientServerTokenStream, forwardClientTokenStream, backwardClientTokenStream)

	parseInput(clientServerTokenStream)

}

/*func createServer(node Node) {
	list, err := net.Listen("tcp", fmt.Sprintf(":%s", node.Port))
	if err != nil {
		fmt.Printf("Server : Failed to listen on port : %v \n", err)
		log.Printf("Server  Failed to listen on port : %v", err) //If it fails to listen on the port, run launchServer method again with the next value/port in ports array
		return
	}

	var opts []grpc.ServerOption
	clientServer := grpc.NewServer(opts...)

	hs.RegisterTokenServiceServer(clientServer, node)

	if err := clientServer.Serve(list); err != nil {
		fmt.Printf("failed to serve %v", err)
	}
}*/

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
		return nil, nil, err
	}
	fmt.Println("Successfully connected to forward node. \n")

	// delete if not used
	node.Clientforward = hs.NewTokenServiceClient(connforward)

	//need to check for if nodeID is 10, since the backward node is 30 or the highst id
	connBackward, err := grpc.Dial(*conPortbackward, opts...)
	if err != nil {
		connforward.Close()
		return nil, nil, err
	}
	fmt.Println("Successfully connected to backward node.")

	// delete if not used
	node.Clientbackward = hs.NewTokenServiceClient(connBackward)

	return connforward, connBackward, nil
}

func parseInput(stream hs.TokenService_TokenChatClient) {
	reader := bufio.NewReader(os.Stdin)

	//Infinite loop to listen for clients input.
	for {

		//Read input into var input and any errors into err
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("%v \n", err)
			log.Fatal(err)
		}
		input = strings.TrimSpace(input) //Trim input

		SendMessage(stream)
	}
}

func SendMessage(stream hs.TokenService_TokenChatClient) {
	message := &hs.TokenRequest{
		Token: "token",
	}
	stream.Send(message)

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
			clientServerTokenStream.Send(&hs.TokenRequest{Token: "token"})
		}
	}

	return nil
}

// this is for client listening
func ListenInternal(stream hs.TokenService_TokenChatClient, forwardStream hs.TokenService_TokenChatClient, backwardStream hs.TokenService_TokenChatClient) {
	for {
		time.Sleep(1 * time.Second)
		if stream != nil {
			msg, err := backwardStream.Recv()
			if err == io.EOF {
				fmt.Printf("Error: io.EOF in listenForMessages \n")
				log.Printf("Error: io.EOF in listenForMessages")
				break
			}
			if err != nil {
				fmt.Printf("%v \n", err)
			}

			if msg.Token == "token" {
				fmt.Printf("%s recieved token", *name)
				log.Printf("%s recieved token", *name)

				fmt.Printf("%s is writing to the critical section", *name)
				log.Printf("%s is writing to the critical section", *name)

				fmt.Printf("%s is sending the token to the next node", *name)
				log.Printf("%s is sending the token to the next node", *name)

				token = false
				forwardStream.Send(&hs.TokenRequest{Token: "token"})
			}
		} else if token {
			fmt.Printf("%s recieved token", *name)
			log.Printf("%s recieved token", *name)

			fmt.Printf("%s is writing to the critical section", *name)
			log.Printf("%s is writing to the critical section", *name)

			fmt.Printf("%s is sending the token to the next node", *name)
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
