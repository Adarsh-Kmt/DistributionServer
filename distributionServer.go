package main

import (
	"fmt"
	"log"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type DistributionServer struct {
	UnimplementedDistributionServerMessageServiceServer

	endServerClientMap map[string]EndServerMessageServiceClient
	activeUserConn     map[string]string
}

func NewDistributionServerInstance() *DistributionServer {

	return &DistributionServer{
		endServerClientMap: make(map[string]EndServerMessageServiceClient),
		activeUserConn:     make(map[string]string),
	}
}
func (ds *DistributionServer) SendMessage(ctx context.Context, message *DistributionServerMessage) (*DistributionServerResponse, error) {

	log.Printf("received message %s from end node.", message.Body)

	receiverUserEndNodeAddress, exists := ds.activeUserConn[message.ReceiverId]

	if !exists {
		log.Fatal("user not found in distribution server activeUserConn.")
	}
	endNodeClient, exists := ds.endServerClientMap[receiverUserEndNodeAddress]

	if !exists {
		log.Fatal("user not found in distribution server endServerClientMap.")
	}

	endNodeMessage := EndServerMessage{ReceiverId: message.ReceiverId, SenderId: message.SenderId, Body: message.Body}
	_, err := endNodeClient.ReceiveMessage(context.Background(), &endNodeMessage)

	if err != nil {

		return &DistributionServerResponse{ResponseStatus: 500}, nil
	}
	return &DistributionServerResponse{ResponseStatus: 200}, nil

}

func (ds *DistributionServer) UserConnected(ctx context.Context, connectionRequest *DistributionServerConnectionRequest) (*DistributionServerResponse, error) {

	ds.activeUserConn[connectionRequest.UserId] = connectionRequest.EndServerAddress

	fmt.Println(ds.activeUserConn)
	if _, exists := ds.endServerClientMap[connectionRequest.EndServerAddress]; !exists {

		conn, err := grpc.NewClient(connectionRequest.EndServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))

		if err != nil {
			return &DistributionServerResponse{ResponseStatus: 500}, nil
		}

		endNodeClient := NewEndServerMessageServiceClient(conn)

		ds.endServerClientMap[connectionRequest.EndServerAddress] = endNodeClient

	}
	log.Printf("user successfully connected according to distribution server.")
	return &DistributionServerResponse{ResponseStatus: 200}, nil

}

func (ds *DistributionServer) UserDisconnected(ctx context.Context, disconnectionRequest *DistributionServerConnectionRequest) (*DistributionServerResponse, error) {

	delete(ds.activeUserConn, disconnectionRequest.UserId)

	return &DistributionServerResponse{ResponseStatus: 200}, nil
}
