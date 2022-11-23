package main

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	glog "google.golang.org/grpc/grpclog"

	auction "github.com/shhoitu/distributed-auction/grpc"
)

var grpcLog glog.LoggerV2

type FrontEnd struct {
	replicationClients map[int]auction.AuctionClient // Map port to replication server
	auction.UnimplementedAuctionServer
}

func main() {
	// Call all the replications (first replication has port 5010)

	// grpcServer := grpc.NewServer()
	// beer.RegisterDistributedMutexServer(grpcServer, p)
	numberOfReplicationServers := 3

	frontEnd := &FrontEnd{}

	// Setup frontEnd client side
	for i := 0; i < numberOfReplicationServers; i++ {
		replicationServerPort := 5010 + (10 * i)

		conn, err := grpc.Dial(fmt.Sprintf(":%v", replicationServerPort), grpc.WithInsecure(), grpc.WithBlock())
		defer conn.Close()

		if err != nil {
			grpcLog.Errorf("Failed to listen on port: %v", err)
		}

		// Client for the replication server at that specific port
		replicationClient := auction.NewAuctionClient(conn)

		// Add the client to frontEnds slice of clients
		frontEnd.replicationClients[replicationServerPort] = replicationClient
	}

	// Setup frontEnd server side
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		grpcLog.Fatalf("Error creating the server %v", err)
	}

	grpcLog.Info("Starting server at port :8080")

	grpcServer := grpc.NewServer()
	auction.RegisterAuctionServer(grpcServer, frontEnd)
	grpcServer.Serve(listener)
}

func (frontEnd *FrontEnd) MakeBid(ctx context.Context, bid *auction.Bid) (*auction.Ack, error) {
	ackChan := make(chan *auction.Ack) // Add buffers if crashing :)

	for _, replicationClient := range frontEnd.replicationClients {
		go func(client auction.AuctionClient) {
			ack, _ := client.MakeBid(context.Background(), bid)
			ackChan <- ack
		}(replicationClient)
	}

	return <-ackChan, nil
}

func (frontEnd *FrontEnd) GetStatus(ctx context.Context, empty *auction.Empty) (*auction.Status, error) {
	statusChan := make(chan *auction.Status) // Add buffers if crashing :)

	for _, replicationClient := range frontEnd.replicationClients {
		go func(client auction.AuctionClient) {
			status, _ := client.GetStatus(context.Background(), empty)
			statusChan <- status
		}(replicationClient)
	}

	return <-statusChan, nil
}
