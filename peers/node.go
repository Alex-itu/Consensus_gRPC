// how to call the peer to peer network
// have one terminal connect to
// go run peers/node.go :5000 localhost:8500
// have another terminal connect to
// go run peers/node.go :5001 localhost:8500
// the peers will then connect

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
	"strconv"

	"strings"
	"time"

	hs "github.com/Alex-itu/Consensus_gRPC/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	//"google.golang.org/grpc/reflection"
)

// Node structure
type Node struct {
	hs.UnimplementedHelloServiceServer
	ID             int
	Port           string
	Client         hs.HelloServiceClient
	Clientforward  hs.HelloServiceClient
	Clientbackward hs.HelloServiceClient
}

var clientservertokenstream hs.HelloServiceClient

var nodeServerconn *grpc.ClientConn

var nodeID = flag.Int("id", 10, "The id for the node")
var conPort = flag.String("port", "10", "port to another node")

func main() {
	flag.Parse()

	node := &Node{
		ID:             *nodeID,
		Port:           strconv.Itoa(*nodeID), //for now ID == Port
		Client:         nil,
		Clientforward:  nil,
		Clientbackward: nil,
	}

	fmt.Printf("nodeID: " + strconv.Itoa(node.ID) + " and port to connect to: " + node.Port)

	createServer(*node)

	time.Sleep(1 * time.Second)

	createClientServerConn(*node)
	defer nodeServerconn.Close()

	tokenStream, err := clientservertokenstream.SayHello(context.Background())
	if err != nil {
		fmt.Printf("Error on receive: %v \n", err)
	}

	// finally when done, simply wait for for access with either token og agaadasdlasd
	go listenForMessages(tokenStream)
	parseInput(tokenStream)

}

func createServer(node Node) {
	list, err := net.Listen("tcp", fmt.Sprintf(":%s", node.Port))
	if err != nil {
		fmt.Printf("Server : Failed to listen on port : %v \n", err)
		log.Printf("Server  Failed to listen on port : %v", err) //If it fails to listen on the port, run launchServer method again with the next value/port in ports array
		return
	}

	var opts []grpc.ServerOption
	clientServer := grpc.NewServer(opts...)

	hs.RegisterHelloServiceServer(clientServer, node)

	if err := clientServer.Serve(list); err != nil {
		fmt.Printf("failed to serve %v", err)
	}
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

	node.Client = hs.NewHelloServiceClient(conn)
	nodeServerconn = conn
}

// connectToOtherNode establishes a connection with the other node and performs a greeting
func connectToOtherNode(node Node) error {
	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	connforward, err := grpc.Dial(strconv.Itoa(*nodeID+10), opts...)
	if err != nil {
		return err
	}
	defer connforward.Close()

	node.Clientforward = hs.NewHelloServiceClient(connforward)

	//need to check for if nodeID is 10, since the backward node is 30 or the highst id
	connBackward, err := grpc.Dial(strconv.Itoa(*nodeID-10), opts...)
	if err != nil {
		return err
	}
	defer connBackward.Close()

	node.Clientbackward = hs.NewHelloServiceClient(connBackward)

	return nil
}

func parseInput(stream hs.HelloService_SayHelloClient) {
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

		SendMessage(input, stream)
	}
}

func SendMessage(content string, stream hs.HelloService_SayHelloClient) {
	message := &hs.HelloRequest{
		Name: "something",
	}
	stream.Send(message)
}

// this is for server listening
func (s *Node) SayHello(msgStream hs.HelloService_SayHelloServer) error {
	// get the next message from the stream
	for {
		msg, err := msgStream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		//placeholder
		if msg.Name == "the client name for this nodeserver" {
			//should send token to the next client ('forward' node's client
			//sendmessage(msg)
		} else if msg.Name == "the name for the 'backwards' node's client" {
			//should send token access to the client for this server
			//sendmessage(msg)
		}
	}

	return nil
}

// this is for client listening
func listenForMessages(stream hs.HelloService_SayHelloClient) {
	for {
		time.Sleep(1 * time.Second)
		if stream != nil {
			msg, err := stream.Recv()
			if err == io.EOF {
				fmt.Printf("Error: io.EOF in listenForMessages \n")
				log.Printf("Error: io.EOF in listenForMessages")
				break
			}
			if err != nil {
				fmt.Printf("%v \n", err)
			}

			//placeholder
			if msg.Message != "something" {
				fmt.Printf("Something something... you now have access")
				// set token to true
			}
		}
	}
}
