package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	grpcChat "github.com/magnusblarsen/go_chat/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var clientsName = flag.String("name", "default", "Senders name")
var serverPort = flag.String("server", "4500", "Tcp server")

var client grpcChat.ServicesClient //TODO: the client right??? 
var ServerConn *grpc.ClientConn //the server connection

type clientHandle struct {
	stream     grpcChat.Services_ChatServiceClient
	clientName string
}

func main() {
	flag.Parse()

	fmt.Println("--- CLIENT APP ---")

	fmt.Println("--- join Server ---")
	connectToServer()
	defer ServerConn.Close()

    shutDown := make(chan bool)
	msgStream, err := client.ChatService(context.Background())
	if err != nil {
		log.Fatalf("Failed to call ChatService: %v", err)
	}
    clientHandle := clientHandle{
        stream: msgStream,
        clientName: *clientsName,
    }
    go clientHandle.bindStdinToServerStream(shutDown)
    go clientHandle.streamListener(shutDown)

	<-shutDown
}

func connectToServer() {
	var opts []grpc.DialOption
	opts = append(
		opts, grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	log.Printf("client %s: Attempts to dial on port %s\n", *clientsName, *serverPort)
	conn, err := grpc.Dial(fmt.Sprintf("localhost:%s", *serverPort), opts...)
	if err != nil {
		log.Printf("Fail to Dial : %v", err)
		return
	}

	// makes a client from the server connection and saves the connection
	// and prints whether or not the connection is READY
	client = grpcChat.NewServicesClient(conn)
	ServerConn = conn
	log.Println("the connection is: ", conn.GetState().String())
}

func (clientHandle *clientHandle) bindStdinToServerStream(shutDown chan bool) {

    fmt.Println("Binding stdin to serverstream")
    for {
        reader := bufio.NewReader(os.Stdin)
        clientMessage, err := reader.ReadString('\n')
        if err != nil {
            log.Fatalf("couldn't read from console: %v", err)
        }
        clientMessage = strings.Trim(clientMessage, "\r\n")

        message := &grpcChat.ClientMessage{
            SenderID: "noget", //TODO: Sender ID
            Message: clientMessage,
        }

        err = clientHandle.stream.Send(message)
        if err != nil {
            log.Printf("Error when sending message to stream: %v", err)
        }
        if message.Message == "bye" {
            fmt.Println("Shutting down")
            shutDown <- true
        }
    }
}

func (clientHandle *clientHandle) streamListener(shutDown chan bool) {
	for {
		msg, err := clientHandle.stream.Recv()
		if err != nil {
            log.Printf("Couldn't receive message from server: %v", err)
		}
		fmt.Printf("%s : %s \n",msg.SenderID ,msg.Message)
	}
}
