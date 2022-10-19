package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	gRPC "github.com/magnusblarsen/go_chat/proto"
	"google.golang.org/grpc"
)

type Server struct {
	// an interface that the server needs to have
	gRPC.UnimplementedServicesServer

	name             string
	port             string
	numberOfRequests int
}

// flags are used to get arguments from the terminal. Flags take a value, a default value and a description of the flag.
// to use a flag then just add it as an argument when running the program.
var serverName = flag.String("name", "default", "Senders name") // set with "-name <name>" in terminal
var port = flag.String("port", "4500", "Server port")           // set with "-port <port>" in terminal

func serverMain() {
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
		numberOfRequests: 0,
	}

	gRPC.RegisterServicesServer(grpcServer, server)
	log.Printf("Server %s: Listening at %v\n", *serverName, list.Addr())
	if err := grpcServer.Serve(list); err != nil {
		log.Fatalf("failed to serve %v", err)
	}
}





