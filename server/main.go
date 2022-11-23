package main

import (
	"context"
	"fmt"
	"log"

	auction "github.com/shhoitu/distributed-auction/grpc"
	"google.golang.org/grpc"
	glog "google.golang.org/grpc/grpclog"
)

var grpcLog glog.LoggerV2

var replicationManager auction.AuctionServer

func connect(auctionServer *auction.AuctionServer) error {
	var streamerror error

	stream, err := replicationManager.Join(context.Background(), &auction.Connect{
		AuctionServer: auctionServer,
	})

	if err != nil {
		return fmt.Errorf("Connection failed :%v", err)
	}
}

func main() {
	conn, err := grpc.Dial("#TODOgetfrontendIP", grpc.WithInsecure())

	if err != nil {
		log.Fatalf("Couldn't connect to service: %v", err)
	}

	replicationManager = auction.newAuctionServer(conn)
	replication := &auction.AuctionServer{
		Id:   number,
		Name: "Server " + number,
	}

	connect(replication)

	wait.Add(1)

}
