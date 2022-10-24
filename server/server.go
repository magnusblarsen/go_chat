package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	grpcChat "github.com/magnusblarsen/go_chat/proto"
	"google.golang.org/grpc"
)

type message struct {
    MessageBody string
    SenderID string
}

type messages struct {
    messageQue []message
    mutex sync.Mutex
}

type Server struct {
    grpcChat.UnimplementedServicesServer // an interface that the server needs to have

	name             string
	port             string
}


//TODO: perhaps better name
//TODO: Should this be stored in server
var messagesObject = messages{}
var connectedClients = []string{}

var serverName = flag.String("name", "default", "Senders name")
var port = flag.String("port", "4500", "Server port")

func main() {
	flag.Parse()
	fmt.Println(".:server is starting:.")
	launchServer()
}

func launchServer() {
	list, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", *port))
	if err != nil {
		log.Printf("Server %s: Failed to listen on port %s: %v", *serverName, *port, err)
	}
	grpcServer := grpc.NewServer()

	server := &Server{
		name:             *serverName,
		port:             *port,
	}

	grpcChat.RegisterServicesServer(grpcServer, server)
	log.Printf("Server %s: Listening at %v\n", *serverName, list.Addr())
	if err := grpcServer.Serve(list); err != nil {
		log.Fatalf("failed to serve %v", err)
	}
}

func (s *Server) ChatService(msgStream grpcChat.Services_ChatServiceServer) error {
    errorChannel := make(chan error)

    go receiveStream(msgStream, errorChannel)
    go messagesListener(msgStream, errorChannel)

    return <- errorChannel
}

func receiveStream(msgStream grpcChat.Services_ChatServiceServer, errorChannel chan error) {
    //TODO: Add the stream of client to list of streams
    for {
        msg, err := msgStream.Recv()
        if err != nil {
            fmt.Printf("err er ikke nil: %v", err)
            errorChannel <- err
            return
        }
        if msg.Message == "bye" {
            //TODO: remove from client list
            ack := &grpcChat.ServerMessage{
                Message: fmt.Sprintf("GoodBye: %s", msg.SenderID),
                SenderID: *serverName,
            }
            msgStream.Send(ack) //TODO: SendAndClose() ???? way to close stream
            errorChannel <- err
            return
        }

        messagesObject.mutex.Lock()

        messagesObject.messageQue = append(messagesObject.messageQue, message{
            MessageBody: msg.Message,
            SenderID: msg.SenderID,
        })
        messagesObject.mutex.Unlock()
        objectBodyReceived := messagesObject.messageQue[len(messagesObject.messageQue)-1]
        fmt.Printf("Message recieved as: %s\nfrom: %s\n", objectBodyReceived.MessageBody, objectBodyReceived.SenderID)
        //TODO: notify all that message has been sent
    }
}

func messagesListener(msgStream grpcChat.Services_ChatServiceServer, errorChannel chan error){
    for {
        time.Sleep(500 * time.Millisecond)

        messagesObject.mutex.Lock()

        if len(messagesObject.messageQue) == 0 {
            messagesObject.mutex.Unlock()
            continue
        }

        //TODO: move below to a sendToClientStream()
        //TODO: for loop to loop through all stream of all clients
        senderID := messagesObject.messageQue[0].SenderID
        newMessage := messagesObject.messageQue[0].MessageBody

        messagesObject.mutex.Unlock()
        
        err := msgStream.Send(&grpcChat.ServerMessage{
            SenderID: senderID,
            Message: newMessage,
        })
        if err != nil {
            errorChannel <- err
        }

        messagesObject.mutex.Lock()

        if len(messagesObject.messageQue) > 1 {
            messagesObject.messageQue = messagesObject.messageQue[1:]
        } else {
            messagesObject.messageQue = []message{}
        }

        messagesObject.mutex.Unlock()
    }
}
