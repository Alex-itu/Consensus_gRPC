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
	ID     int
	Client hs.HelloServiceClient
}

// startServer starts the passed in gRPC server
func startServer(server *grpc.Server, lis net.Listener) {
	go func() {
		if err := server.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
}

var nodeID = flag.Int("id", 10, "The id for the node")
var connectedPort = flag.String("port", "20", "port to another node")

func main() {
	//args := os.Args[1:]

	//id, err := strconv.Atoi(args[0])
	//address := args[1]
	flag.Parse()

	node := &Node{ID: *nodeID, Client: nil}

	fmt.Println("--- CLIENT APP ---")
	fmt.Printf("nodeID: " + strconv.Itoa(*nodeID) + " and port to connect to: " + *connectedPort)

	//"Discorver" other nodes (We just need hard coded values)
	// if other nodes exsist, the join the connection
	//connectToOtherNode(node, *connectedPort) //TODO: fix

	time.Sleep(10 * time.Second)

	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	conn, err := grpc.Dial(*connectedPort, opts...)
	if err != nil {
		fmt.Printf("failed on Dial: %v", err)
	}
	//defer conn.Close()

	node.Client = hs.NewHelloServiceClient(conn)

	//fmt.Println("the connection is: ", conn.GetState().String())

	var yoyo hs.HelloServiceClient // maybe set global
	ChatStream, err := yoyo.SayHello(context.Background())
	if err != nil {
		fmt.Printf("Error on receive: %v \n", err)
		log.Fatalf("Error on receive: %v", err)
	}
	// finally when done, simply wait for for access with either token og agaadasdlasd
	go listenForMessages(ChatStream)
	parseInput(ChatStream)

}

// connectToOtherNode establishes a connection with the other node and performs a greeting
func connectToOtherNode(node *Node, address string) error {
	time.Sleep(10 * time.Second)

	conn, err := grpc.Dial(address /*, grpc.WithInsecure() // this is deprecated*/)
	if err != nil {
		return err
	}
	defer conn.Close()

	node.Client = hs.NewHelloServiceClient(conn)

	fmt.Println("the connection is: ", conn.GetState().String())
	return nil
}

func parseInput(stream hs.HelloService_SayHelloClient) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("--------------------")

	//Infinite loop to listen for clients input.
	for {
		fmt.Print("-> ")

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

func listenForMessages(stream hs.HelloService_SayHelloClient) {
	for {
		time.Sleep(1 * time.Second)
		if stream != nil {
			msg, err := stream.Recv()
			if err == io.EOF {
				fmt.Printf("Error: io.EOF in listenForMessages in client.go \n")
				log.Printf("Error: io.EOF in listenForMessages in client.go")
				break
			}
			if err != nil {
				fmt.Printf("%v \n", err)
				log.Fatalf("%v", err)

				//delete later
				fmt.Printf(msg.String())
			}

			/*if strings.Contains("msg.Content", "*clientsName"+" Connected") {
				// Updates the clientID
				//NodeID = int(msg.ClientID)

			}*/
		}
	}
}
