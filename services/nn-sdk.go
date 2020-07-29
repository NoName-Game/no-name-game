package services

import (
	"log"

	"google.golang.org/grpc"

	_ "github.com/joho/godotenv/autoload" // Autload .env

	pb "bitbucket.org/no-name-game/nn-grpc/rpc"
)

var (
	// NnSDK - NoName SDK
	// NnSDK *nnsdk.NoNameSDK
	NnSDK pb.NoNameClient
)

// NnSDKUp - Metodo di verfica e connessione al WS principale di NoName
func NnSDKUp() (err error) {
	// Set up a connection to the server.
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Panicf("did not connect: %v", err)
	}

	NnSDK = pb.NewNoNameClient(conn)
	// defer conn.Close()

	// ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// defer cancel()
	// r, err := c.GetMoney(ctx, &pb.GetMoneyRequest{
	// 	Name: "LoL",
	// })
	// if err != nil {
	// 	log.Fatalf("could not greet: %v", err)
	// }
	//
	// log.Panicf("Greeting: %s", r.GetMessage())

	// Riporto a video stato servizio
	log.Println("************************************************")
	log.Println("NoName WS connection: OK!")
	log.Println("************************************************")

	return
}
