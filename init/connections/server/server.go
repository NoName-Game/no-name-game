package server

import (
	"fmt"
	"log"
	"os"

	"google.golang.org/grpc"

	_ "github.com/joho/godotenv/autoload" // Autload .env

	pb "bitbucket.org/no-name-game/nn-grpc/build/proto"
)

// Server
type Server struct {
	Connection pb.NoNameClient
}

// NnSDKUp - Metodo di verfica e connessione al WS principale di NoName
func (server *Server) Init() {
	var err error

	// Set up a connection to the server.
	var conn *grpc.ClientConn
	conn, err = grpc.Dial(
		fmt.Sprintf("%s:%s", os.Getenv("WS_HOST"), os.Getenv("WS_PORT")),
		grpc.WithInsecure(),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Panicf("did not connect: %v", err)
	}

	server.Connection = pb.NewNoNameClient(conn)

	// Riporto a video stato servizio
	log.Println("************************************************")
	log.Println("NoName WS connection: OK!")
	log.Println("************************************************")

	return
}
