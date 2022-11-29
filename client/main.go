package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"

	auction "github.com/shhoitu/distributed-auction/grpc"
	"google.golang.org/grpc"
	glog "google.golang.org/grpc/grpclog"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var grpcLog glog.LoggerV2

func init() {
	grpcLog = glog.NewLoggerV2(os.Stdout, os.Stdout, os.Stdout)
}

type Bidder struct {
	client auction.AuctionClient
	id     int32
}

func main() {
	id, _ := strconv.ParseInt(os.Args[1], 10, 32)

	frontEndPort := 8010 + (id * 10)
	grpcLog.Infof("Trying to dial %d", frontEndPort)
	conn, err := grpc.Dial(fmt.Sprintf(":%v", frontEndPort), grpc.WithInsecure(), grpc.WithBlock())
	grpcLog.Infof("Dialed %d", frontEndPort)
	if err != nil {
		grpcLog.Fatalf("Couldn't connect to service: %v", err)
	}

	bidder := &Bidder{id: int32(id), client: auction.NewAuctionClient(conn)}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		amount, err := strconv.ParseInt(scanner.Text(), 10, 32)
		if err != nil {
			continue
		}

		if amount == 0 {
			request := auction.StatusRequest{BidderId: bidder.id}
			status, err := bidder.client.GetStatus(context.Background(), &request)
			if err != nil {
				grpcLog.Errorf(err.Error())
				continue
			}

			timeLeft := status.TimeLeft.AsDuration()

			if timeLeft <= 0 {
				grpcLog.Info("Auction is over.")
			} else {
				grpcLog.Infof("Time left: %v", timeLeft)
			}
			grpcLog.Infof("Bidder: %d has the highest bid: %d", status.BidderId, status.HighestBid)
			continue
		}

		bid := auction.Bid{
			BidderId: bidder.id,
			Amount:   int32(amount),
			Time:     timestamppb.Now(),
		}
		_, bidErr := bidder.client.MakeBid(context.Background(), &bid)

		if bidErr != nil {
			grpcLog.Errorf(err.Error())
			continue
		}

		grpcLog.Infof("Bid with amount: %d sent!", amount)
	}
}
