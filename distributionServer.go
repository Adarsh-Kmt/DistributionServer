package main

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"os"

	database "github.com/Adarsh-Kmt/DistributionServer/database"
	generatedCode "github.com/Adarsh-Kmt/DistributionServer/generatedCode"
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

	DistributionServerCert, err := tls.LoadX509KeyPair("/app/DistributionServer.pem", "/app/DistributionServer-key.pem")

	if err != nil {

		log.Fatal("error while loading key pair of Distribution Server: " + err.Error())
	}

	RootCA := x509.NewCertPool()

	caBytes, err := os.ReadFile("/app/root.pem")

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

	log.Printf("received message %s from end node.", message.Body)

	//receiverUserEndNodeAddress, exists := ds.activeUserConn[message.ReceiverId]

	receiverUserEndNodeAddress, err := ds.RedisDBStore.FindUserEndServerAddress(message.ReceiverId)

	if err != nil {
		log.Println(err.Error())
		log.Fatal("user not found in distribution server activeUserConn.")
	}
	endNodeClient, exists := ds.endServerClientMap[receiverUserEndNodeAddress]

	if !exists {
		log.Fatal("end server client not found in distribution server endServerClientMap.")
	}

	endNodeMessage := generatedCode.EndServerMessage{ReceiverId: message.ReceiverId, SenderId: message.SenderId, Body: message.Body}
	_, err = endNodeClient.ReceiveMessage(context.Background(), &endNodeMessage)

	if err != nil {

		return &generatedCode.DistributionServerResponse{ResponseStatus: 500}, nil
	}
	return &generatedCode.DistributionServerResponse{ResponseStatus: 200}, nil

}

func (ds *DistributionServer) UserConnected(ctx context.Context, connectionRequest *generatedCode.DistributionServerConnectionRequest) (*generatedCode.DistributionServerResponse, error) {

	err := ds.RedisDBStore.UserConnected(connectionRequest.UserId, connectionRequest.EndServerAddress)
	//ds.activeUserConn[connectionRequest.UserId] = connectionRequest.EndServerAddress

	if err != nil {
		log.Println(err.Error())
		log.Fatal("error while logging user connection status in redis db.")
	} else {
		log.Println("received user connection request from " + connectionRequest.UserId)
	}
	//fmt.Println(ds.activeUserConn)
	if _, exists := ds.endServerClientMap[connectionRequest.EndServerAddress]; !exists {

		tlsConfig := GenerateTLSConfigObjectForDistributionServer()

		conn, err := grpc.NewClient(connectionRequest.EndServerAddress, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))

		if err != nil {
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

	err := ds.RedisDBStore.UserDisconnected(disconnectionRequest.UserId)

	if err != nil {
		log.Println(err.Error())
		log.Fatal("error while logging user disconnection status.")
	}
	return &generatedCode.DistributionServerResponse{ResponseStatus: 200}, nil
}
