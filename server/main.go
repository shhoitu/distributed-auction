package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"

	auction "github.com/shhoitu/distributed-auction/grpc"
	"google.golang.org/grpc"
	glog "google.golang.org/grpc/grpclog"
)

var grpcLog glog.LoggerV2

func init() {
	grpcLog = glog.NewLoggerV2(os.Stdout, os.Stdout, os.Stdout)
}

type ReplicationManager struct {
	auction.UnimplementedAuctionServer
	highestBid *auction.Bid
	bidLock    sync.Mutex
}

func main() {
	arg1, _ := strconv.ParseInt(os.Args[1], 10, 32)
	ownPort := 5010 + (arg1 * 10)

	replicationManager := &ReplicationManager{highestBid: &auction.Bid{BidderId: -1, Amount: 0}}

	// Setup frontEnd server side
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", ownPort))
	if err != nil {
		grpcLog.Fatalf("Error creating the server %v", err)
	}

	grpcLog.Infof("Starting server at port :%v", ownPort)

	grpcServer := grpc.NewServer()
	auction.RegisterAuctionServer(grpcServer, replicationManager)
	grpcServer.Serve(listener)
}

func (replicationManager *ReplicationManager) MakeBid(ctx context.Context, bid *auction.Bid) (*auction.Ack, error) {
	grpcLog.Infof("Bid: %d", bid.Amount)

	replicationManager.setHighestBid(bid)
	return &auction.Ack{}, nil
}

func (replicationManager *ReplicationManager) GetStatus(ctx context.Context, empty *auction.Empty) (*auction.Status, error) {
	highestBid := replicationManager.getHighestBid()
	status := &auction.Status{
		Status:      "status",
		SecondsLeft: 0,
		HighestBid:  highestBid.Amount,
	}
	return status, nil
}

func (replicationManager *ReplicationManager) setHighestBid(bid *auction.Bid) {
	replicationManager.bidLock.Lock()
	defer replicationManager.bidLock.Unlock()
	if bid.Amount > replicationManager.highestBid.Amount {
		replicationManager.highestBid = bid
	}
	grpcLog.Infof("HighestBid set !!!!")
}

func (replicationManager *ReplicationManager) getHighestBid() *auction.Bid {
	replicationManager.bidLock.Lock()
	defer replicationManager.bidLock.Unlock()
	return replicationManager.highestBid
}
