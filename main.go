package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"google.golang.org/grpc"
	glog "google.golang.org/grpc/grpclog"

	auction "github.com/shhoitu/distributed-auction/grpc"
)

var grpcLog glog.LoggerV2

func init() {
	grpcLog = glog.NewLoggerV2(os.Stdout, os.Stdout, os.Stdout)
}

type FrontEnd struct {
	replicationClients []auction.AuctionClient // Map port to replication server
	auction.UnimplementedAuctionServer
}

func main() {
	frontEnd := &FrontEnd{}

	// Setup frontEnd client side
	numberOfReplicationServers := 3
	for i := 0; i < numberOfReplicationServers; i++ {
		// Call all the replications (first replication has port 5010)
		replicationServerPort := 5010 + (10 * i)

		conn, err := grpc.Dial(fmt.Sprintf(":%v", replicationServerPort), grpc.WithTimeout(3*time.Second), grpc.WithInsecure(), grpc.WithBlock())

		if err != nil {
			grpcLog.Errorf("Failed to listen on port: %v, error: %v", replicationServerPort, err)
			continue
		}

		defer conn.Close()

		// Client for the replication server at that specific port
		replicationClient := auction.NewAuctionClient(conn)

		// Add the client to frontEnds slice of clients
		frontEnd.replicationClients = append(frontEnd.replicationClients, replicationClient)
	}

	// Setup frontEnd server side
	arg1, _ := strconv.ParseInt(os.Args[1], 10, 32)
	ownPort := 8010 + (arg1 * 10)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", ownPort))
	if err != nil {
		grpcLog.Fatalf("Error creating the server %v", err)
	}

	grpcLog.Infof("Starting server at port :%d", ownPort)

	grpcServer := grpc.NewServer()
	auction.RegisterAuctionServer(grpcServer, frontEnd)
	grpcServer.Serve(listener)
}

func (frontEnd *FrontEnd) MakeBid(ctx context.Context, bid *auction.Bid) (*auction.Ack, error) {
	ackChan := make(chan *auction.Ack) // Add buffers if crashing :)

	grpcLog.Infof("Bid: %d from bidder: %d", bid.Amount, bid.BidderId)

	for index, replicationClient := range frontEnd.replicationClients {
		go func(client auction.AuctionClient, index int) {
			ack, err := client.MakeBid(context.Background(), bid)
			if err != nil {
				frontEnd.removeReplicationManager(index)
				return
			}
			ackChan <- ack
		}(replicationClient, index)
	}

	return <-ackChan, nil
}

func (frontEnd *FrontEnd) GetStatus(ctx context.Context, request *auction.StatusRequest) (*auction.Status, error) {
	statusChan := make(chan *auction.Status) // Add buffers if crashing :)

	grpcLog.Info("Status requested")

	for index, replicationClient := range frontEnd.replicationClients {
		go func(client auction.AuctionClient, index int) {
			status, err := client.GetStatus(context.Background(), request)
			if err != nil {
				frontEnd.removeReplicationManager(index)
				return
			}
			statusChan <- status
		}(replicationClient, index)
	}

	grpcLog.Info("Returning status")
	return <-statusChan, nil
}

func (frontEnd *FrontEnd) removeReplicationManager(index int) {
	grpcLog.Infof("A Replication Manager at index %d is not responding.", index)
	frontEnd.replicationClients = append(frontEnd.replicationClients[:index], frontEnd.replicationClients[index+1:]...)
}
