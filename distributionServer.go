package main

import (
	"fmt"
	"log"

	generatedCode "github.com/Adarsh-Kmt/DistributionServer/generatedCode"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type DistributionServer struct {
	generatedCode.UnimplementedDistributionServerMessageServiceServer

	endServerClientMap map[string]generatedCode.EndServerMessageServiceClient
	activeUserConn     map[string]string
}

func NewDistributionServerInstance() *DistributionServer {

	return &DistributionServer{
		endServerClientMap: make(map[string]generatedCode.EndServerMessageServiceClient),
		activeUserConn:     make(map[string]string),
	}
}
func (ds *DistributionServer) SendMessage(ctx context.Context, message *generatedCode.DistributionServerMessage) (*generatedCode.DistributionServerResponse, error) {

	log.Printf("received message %s from end node.", message.Body)

	receiverUserEndNodeAddress, exists := ds.activeUserConn[message.ReceiverId]

	if !exists {
		log.Fatal("user not found in distribution server activeUserConn.")
	}
	endNodeClient, exists := ds.endServerClientMap[receiverUserEndNodeAddress]

	if !exists {
		log.Fatal("user not found in distribution server endServerClientMap.")
	}

	endNodeMessage := generatedCode.EndServerMessage{ReceiverId: message.ReceiverId, SenderId: message.SenderId, Body: message.Body}
	_, err := endNodeClient.ReceiveMessage(context.Background(), &endNodeMessage)

	if err != nil {

		return &generatedCode.DistributionServerResponse{ResponseStatus: 500}, nil
	}
	return &generatedCode.DistributionServerResponse{ResponseStatus: 200}, nil

}

func (ds *DistributionServer) UserConnected(ctx context.Context, connectionRequest *generatedCode.DistributionServerConnectionRequest) (*generatedCode.DistributionServerResponse, error) {

	ds.activeUserConn[connectionRequest.UserId] = connectionRequest.EndServerAddress

	fmt.Println(ds.activeUserConn)
	if _, exists := ds.endServerClientMap[connectionRequest.EndServerAddress]; !exists {

		conn, err := grpc.NewClient(connectionRequest.EndServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))

		if err != nil {
			return &generatedCode.DistributionServerResponse{ResponseStatus: 500}, nil
		}

		endNodeClient := generatedCode.NewEndServerMessageServiceClient(conn)

		ds.endServerClientMap[connectionRequest.EndServerAddress] = endNodeClient

	}
	log.Printf("user successfully connected according to distribution server.")
	return &generatedCode.DistributionServerResponse{ResponseStatus: 200}, nil

}

func (ds *DistributionServer) UserDisconnected(ctx context.Context, disconnectionRequest *generatedCode.DistributionServerConnectionRequest) (*generatedCode.DistributionServerResponse, error) {

	delete(ds.activeUserConn, disconnectionRequest.UserId)

	return &generatedCode.DistributionServerResponse{ResponseStatus: 200}, nil
}
