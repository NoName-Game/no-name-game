package server

import (
	"fmt"
	"os"
	"time"

	"google.golang.org/grpc/keepalive"

	"github.com/sirupsen/logrus"

	"google.golang.org/grpc"

	_ "github.com/joho/godotenv/autoload" // Autload .env

	"bitbucket.org/no-name-game/nn-grpc/build/pb"
)

// Server
type Server struct {
	Connection pb.NoNameClient
}

var kacp = keepalive.ClientParameters{
	Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
	Timeout:             time.Second,      // wait 1 second for ping ack before considering the connection dead
	PermitWithoutStream: true,             // send pings even without active streams
}

// Init - Metodo di verfica e connessione al WS principale di NoName
func (server *Server) Init() {
	var err error

	// Set up a connection to the server.
	var conn *grpc.ClientConn
	if conn, err = grpc.Dial(
		fmt.Sprintf("%s:%s", os.Getenv("WS_HOST"), os.Getenv("WS_PORT")),
		grpc.WithInsecure(),
		// grpc.WithBlock(),
		grpc.WithKeepaliveParams(kacp),
	); err != nil {
		logrus.WithField("error", err).Fatal("[*] Server connections: KO!")
	}

	server.Connection = pb.NewNoNameClient(conn)

	logrus.Info("[*] Server connections: OK!")
	return
}
