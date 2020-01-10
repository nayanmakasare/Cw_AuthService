package main

import (
	"Cw_authService/apihandler"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net"
	"time"

	pb "Cw_authService/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	grpc_port        = ":7764"
	rest_port		 = ":7765"
	srvCertFile = "certs/server.crt"
	srvKeyFile  = "certs/server.key"
	//defaultHost          = "mongodb://nayan:tlwn722n@cluster0-shard-00-00-8aov2.mongodb.net:27017,cluster0-shard-00-01-8aov2.mongodb.net:27017,cluster0-shard-00-02-8aov2.mongodb.net:27017/test?ssl=true&replicaSet=Cluster0-shard-0&authSource=admin&retryWrites=true&w=majority"
	developmentMongoHost = "mongodb://dev-uni.cloudwalker.tv:6592"
	schedularMongoHost   = "mongodb://192.168.1.143:27017"
	schedularRedisHost   = "redis:6379"
)

func getMongoCollection(dbName, collectionName, mongoHost string) *mongo.Collection {
	// Register custom codecs for protobuf Timestamp and wrapper types
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoHost),)
	if err != nil {
		log.Fatal(err)
	}
	return mongoClient.Database(dbName).Collection(collectionName)
}

func startGRPCServer(address, certFile, keyFile string) error {
	// create a listener on TCP port
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	} // create a server instance
	s := apihandler.AuthService{
		getMongoCollection("cloudwalker", "users", developmentMongoHost),
	}

	// Create the TLS credentials
	creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
	if err != nil {
		return fmt.Errorf("could not load TLS keys: %s", err)
	} // Create an array of gRPC options with the credentials
	_ = []grpc.ServerOption{grpc.Creds(creds),}

	// create a gRPC server object
	//grpcServer := grpc.NewServer(opts...)

	// attach the Ping service to the server
	grpcServer := grpc.NewServer()                    // attach the Ping service to the server
	pb.RegisterCw_AuthServiceServer(grpcServer, &s) // start the server
	log.Printf("starting HTTP/2 gRPC server on %s", address)
	if err := grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %s", err)
	}
	return nil
}

func credMatcher(headerName string) (mdName string, ok bool) {
	if headerName == "Login" || headerName == "Password" {
		return headerName, true
	}
	return "", false
}

//func startRESTServer(address, grpcAddress, certFile string) error {
//	ctx := context.Background()
//	ctx, cancel := context.WithCancel(ctx)
//	defer cancel()
//	mux := runtime.NewServeMux(runtime.WithIncomingHeaderMatcher(credMatcher))
//	//creds, err := credentials.NewClientTLSFromFile(certFile, "")
//	//if err != nil {
//	//	return fmt.Errorf("could not load TLS certificate: %s", err)
//	//}  // Setup the client gRPC options
//	//
//	//opts := []grpc.DialOption{grpc.WithTransportCredentials(creds)}  // Register ping
//
//	opts := []grpc.DialOption{grpc.WithInsecure()} // Register ping
//	err := pb.RegisterSchedularServiceHandlerFromEndpoint(ctx, mux, grpcAddress, opts)
//	if err != nil {
//		return fmt.Errorf("could not register service Ping: %s", err)
//	}
//
//	log.Printf("starting HTTP/1.1 REST server on %s", address)
//	http.ListenAndServe(address, mux)
//	return nil
//}


func main() {

	//grpcAddress := fmt.Sprintf("%s:%d", "cloudwalker.services.tv", 7775)
	//restAddress := fmt.Sprintf("%s:%d", "cloudwalker.services.tv", 7776)
	grpcAddress := fmt.Sprintf(":%d",  grpc_port)
	//restAddress := fmt.Sprintf(":%d",  7765)


	// fire the gRPC server in a goroutine
	go func() {
		err := startGRPCServer(grpcAddress, srvCertFile, srvKeyFile)
		if err != nil {
			log.Fatalf("failed to start gRPC server: %s", err)
		}
	}()

	// fire the REST server in a goroutine
	//go func() {
	//	err := startRESTServer(restAddress, grpcAddress, certFile)
	//	if err != nil {
	//		log.Fatalf("failed to start gRPC server: %s", err)
	//	}
	//}()

	log.Printf("Entering infinite loop")
	select {}
}
