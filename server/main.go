package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	auction "github.com/shhoitu/distributed-auction/grpc"
	"google.golang.org/grpc"
	glog "google.golang.org/grpc/grpclog"
	"google.golang.org/protobuf/types/known/durationpb"
)

var grpcLog glog.LoggerV2

func init() {
	grpcLog = glog.NewLoggerV2(os.Stdout, os.Stdout, os.Stdout)
}

type ReplicationManager struct {
	auction.UnimplementedAuctionServer
	endTime    time.Time
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
	if replicationManager.endTime.IsZero() {
		auctionEndTime := bid.Time.AsTime().Add(time.Minute)
		replicationManager.endTime = auctionEndTime
		grpcLog.Infof("Setting end time to: %v", auctionEndTime)
	}

	if replicationManager.endTime.Before(time.Now()) {
		grpcLog.Infof("Bidder %d is trying to bid after time ended", bid.BidderId)
		return &auction.Ack{}, nil
	}

	if replicationManager.setHighestBid(bid) {
		grpcLog.Infof("Bidder %d has highest bid of: %d", bid.BidderId, bid.Amount)
	} else {
		grpcLog.Infof("The bid %d from bidder %d was too low", bid.Amount, bid.BidderId)
	}
	return &auction.Ack{}, nil
}

func (replicationManager *ReplicationManager) GetStatus(ctx context.Context, request *auction.StatusRequest) (*auction.Status, error) {
	grpcLog.Infof("Bidder: %d is requesting status", request.BidderId)

	highestBid := replicationManager.getHighestBid()
	timeLeft := time.Until(replicationManager.endTime)

	status := &auction.Status{
		TimeLeft:   durationpb.New(timeLeft),
		HighestBid: highestBid.Amount,
		BidderId:   highestBid.BidderId,
	}

	return status, nil
}

func (replicationManager *ReplicationManager) setHighestBid(bid *auction.Bid) bool {
	replicationManager.bidLock.Lock()
	defer replicationManager.bidLock.Unlock()
	if bid.Amount < replicationManager.highestBid.Amount {
		return false
	}
	replicationManager.highestBid = bid
	return true
}

func (replicationManager *ReplicationManager) getHighestBid() *auction.Bid {
	replicationManager.bidLock.Lock()
	defer replicationManager.bidLock.Unlock()
	return replicationManager.highestBid
}
