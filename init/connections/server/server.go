package server

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"

	"google.golang.org/grpc"

	_ "github.com/joho/godotenv/autoload" // Autload .env

	pb "bitbucket.org/no-name-game/nn-grpc/build/proto"
)

// Server
type Server struct {
	Connection pb.NoNameClient
}

// Init - Metodo di verfica e connessione al WS principale di NoName
func (server *Server) Init() {
	var err error

	// Set up a connection to the server.
	var conn *grpc.ClientConn
	if conn, err = grpc.Dial(
		fmt.Sprintf("%s:%s", os.Getenv("WS_HOST"), os.Getenv("WS_PORT")),
		grpc.WithInsecure(),
		grpc.WithBlock(),
	); err != nil {
		logrus.WithField("error", err).Fatal("[*] Server connections: KO!")
	}

	server.Connection = pb.NewNoNameClient(conn)

	logrus.Info("[*] Server connections: OK!")
	return
}
