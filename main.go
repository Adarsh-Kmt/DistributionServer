package main

import (
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"
)

func DistributionServerStartListening(GRPCDistributionServer *grpc.Server, waitGroup *sync.WaitGroup) {

	defer waitGroup.Done()
	DNLis, err := net.Listen("tcp", ":9000")

	if err != nil {
		log.Fatal("error")
	}

	if err := GRPCDistributionServer.Serve(DNLis); err != nil {

		log.Fatal("error again.")
	}

}
func main() {

	var wg sync.WaitGroup

	GRPCDistributionServer := NewGRPCDistributionServerInstance()
	wg.Add(1)
	go DistributionServerStartListening(GRPCDistributionServer, &wg)

	wg.Wait()

}
