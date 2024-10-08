package main

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"os"

	database "github.com/Adarsh-Kmt/DistributionServer/database"
	generatedCode "github.com/Adarsh-Kmt/DistributionServer/generatedCode"
	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
)

type DistributionServer struct {
	generatedCode.UnimplementedDistributionServerMessageServiceServer

	endServerClientMap map[string]generatedCode.EndServerMessageServiceClient
	//activeUserConn     map[string]string
	RedisDBStore *database.RedisDBStore
}

func NewDistributionServerInstance() *DistributionServer {

	RedisDBStoreInstance := database.NewRedisDBInstance()
	return &DistributionServer{
		endServerClientMap: make(map[string]generatedCode.EndServerMessageServiceClient),
		//activeUserConn:     make(map[string]string),
		RedisDBStore: RedisDBStoreInstance,
	}
}

func GenerateTLSConfigObjectForDistributionServer() *tls.Config {

	DistributionServerCert, err := tls.LoadX509KeyPair("/prod/DistributionServer.pem", "/prod/DistributionServer-key.pem")

	if err != nil {

		log.Fatal("error while loading key pair of Distribution Server: " + err.Error())
	}

	RootCA := x509.NewCertPool()

	caBytes, err := os.ReadFile("/prod/root.pem")

	if err != nil {

		log.Fatal("error while reading root certificate from file in Distribution Server: " + err.Error())
	}

	if ok := RootCA.AppendCertsFromPEM(caBytes); !ok {

		log.Fatal("failed to load certificate of root CA into certificate poll in Distribution Server.")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{DistributionServerCert},
		ClientCAs:    RootCA,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}

	if err != nil {
		log.Fatal("error while loading TLS certificate of Distribution Server.")
	}

	return tlsConfig

}
func NewGRPCDistributionServerInstance() *grpc.Server {

	distributionServerInstance := NewDistributionServerInstance()

	tlsConfig := GenerateTLSConfigObjectForDistributionServer()

	GRPCDistributionServer := grpc.NewServer(grpc.Creds(credentials.NewTLS(tlsConfig)), grpc.UnaryInterceptor(MiddlewareHandler))

	generatedCode.RegisterDistributionServerMessageServiceServer(GRPCDistributionServer, distributionServerInstance)

	return GRPCDistributionServer
}

func MiddlewareHandler(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	// you can write your own code here to check client tls certificate
	if p, ok := peer.FromContext(ctx); ok {
		if mtls, ok := p.AuthInfo.(credentials.TLSInfo); ok {
			for _, item := range mtls.State.PeerCertificates {
				log.Println("client certificate subject:", item.Subject)
			}
		}
	}
	return handler(ctx, req)
}
func (ds *DistributionServer) SendMessage(ctx context.Context, message *generatedCode.DistributionServerMessage) (*generatedCode.DistributionServerResponse, error) {

	log.Printf("user %s sent message %s to %s end node", message.SenderUsername, message.Body, message.ReceiverUsername)

	receiverUserEndNodeAddress, err := ds.RedisDBStore.FindUserEndServerAddress(message.ReceiverUsername)

	response := &generatedCode.DistributionServerResponse{}
	if err == redis.Nil {
		log.Println("user not found in routing table in redis DB. User " + message.ReceiverUsername + " is offline: " + err.Error())
		response.ResponseStatus = 404
		log.Println(response)
		return response, nil
	}
	if err != nil {
		log.Println("internal server error: " + err.Error())
		return &generatedCode.DistributionServerResponse{ResponseStatus: 500}, err
	}
	endNodeClient, exists := ds.endServerClientMap[receiverUserEndNodeAddress]

	if !exists {
		log.Println("end server client not found in distribution server endServerClientMap.")
		return &generatedCode.DistributionServerResponse{ResponseStatus: 500}, nil
	}

	endNodeMessage := generatedCode.EndServerMessage{ReceiverUsername: message.ReceiverUsername, SenderUsername: message.SenderUsername, Body: message.Body}

	endServerResponse, _ := endNodeClient.ReceiveMessage(context.Background(), &endNodeMessage)

	if endServerResponse.Status == 404 {
		response.ResponseStatus = 404
		log.Println("user " + message.ReceiverUsername + " is not online right now.")
		return response, nil
	}
	log.Println("user " + message.ReceiverUsername + " has received message: " + message.Body + " from " + message.SenderUsername)
	return &generatedCode.DistributionServerResponse{ResponseStatus: 200}, nil

}

func (ds *DistributionServer) UserConnected(ctx context.Context, connectionRequest *generatedCode.DistributionServerConnectionRequest) (*generatedCode.DistributionServerResponse, error) {

	err := ds.RedisDBStore.UserConnected(connectionRequest.Username, connectionRequest.EndServerAddress)

	if err != nil {
		log.Println("error while logging user connection status in redis db." + err.Error())
		return &generatedCode.DistributionServerResponse{ResponseStatus: 500}, err
	} else {
		log.Println("received user connection request from " + connectionRequest.Username)
	}

	if _, exists := ds.endServerClientMap[connectionRequest.EndServerAddress]; !exists {

		tlsConfig := GenerateTLSConfigObjectForDistributionServer()

		conn, err := grpc.NewClient(connectionRequest.EndServerAddress, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))

		if err != nil {
			log.Println("error while establishing secure connection with end server with address " + connectionRequest.EndServerAddress + ": " + err.Error())
			return &generatedCode.DistributionServerResponse{ResponseStatus: 500}, err
		} else {

			log.Println("Distribution Server has successfully connected to End Server at address: " + connectionRequest.EndServerAddress)
		}

		endNodeClient := generatedCode.NewEndServerMessageServiceClient(conn)

		ds.endServerClientMap[connectionRequest.EndServerAddress] = endNodeClient

	}
	log.Printf("user successfully connected according to distribution server.")
	return &generatedCode.DistributionServerResponse{ResponseStatus: 200}, nil

}

func (ds *DistributionServer) UserDisconnected(ctx context.Context, disconnectionRequest *generatedCode.DistributionServerConnectionRequest) (*generatedCode.DistributionServerResponse, error) {

	err := ds.RedisDBStore.UserDisconnected(disconnectionRequest.Username)

	if err != nil {
		log.Println("error while logging user disconnection status." + err.Error())
		return &generatedCode.DistributionServerResponse{ResponseStatus: 500}, nil
	}
	return &generatedCode.DistributionServerResponse{ResponseStatus: 200}, nil
}
