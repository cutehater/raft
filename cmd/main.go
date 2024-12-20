package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"

	"hw2/internal/handlers"
	"hw2/internal/raft"
	"hw2/internal/rpc"
)

var (
	GRPCNodesAddress = []string{
		"localhost:5030",
		"localhost:5031",
		"localhost:5032",
		"localhost:5033",
		"localhost:5034",
	}
	HTTPNodesAddress = []string{
		"localhost:5035",
		"localhost:5036",
		"localhost:5037",
		"localhost:5038",
		"localhost:5039",
	}
)

func main() {
	isLeaderFlag := flag.Bool("leader", false, "Specify if the node is a leader")
	flag.Parse()
	isLeader := isLeaderFlag != nil && *isLeaderFlag

	id, err := strconv.Atoi(os.Args[len(os.Args)-1])
	if err != nil || id >= len(GRPCNodesAddress) || id < 0 {
		fmt.Println("Usage: <command> [--leader] <node index>")
		os.Exit(1)
	}

	node := raft.NewNode(int64(id), isLeader, GRPCNodesAddress, HTTPNodesAddress)

	go startGRPCServer(id, node)
	go startHTTPServer(id, node)
	select {}
}

func startGRPCServer(id int, node *raft.Node) {
	server := grpc.NewServer()
	address := GRPCNodesAddress[id]

	rpc.RegisterRaftNodeServer(server, node)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", address, err)
	}

	log.Printf("gRPC server listening on %s", address)

	if err := server.Serve(listener); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}

func startHTTPServer(id int, node *raft.Node) {
	address := HTTPNodesAddress[id]
	r := mux.NewRouter()

	r.HandleFunc("/create", handlers.MakePostHandler(node))
	r.HandleFunc("/read/{id}", handlers.MakeMasterReadHandler(node))
	r.HandleFunc("/readReplica/{id}", handlers.MakeReplicaReadHandler(node))
	r.HandleFunc("/update", handlers.MakeUpdateHandler(node))

	log.Printf("HTTP server listening on %s", address)
	if err := http.ListenAndServe(address, r); err != nil {
		log.Fatalf("Failed to serve HTTP server: %v", err)
	}
}
