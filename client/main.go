package main

import (
	"bufio"
	"context"
	"os"
	"strconv"

	auction "github.com/shhoitu/distributed-auction/grpc"
	"google.golang.org/grpc"
	glog "google.golang.org/grpc/grpclog"
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
	grpcLog.Info("Trying to dial 8080")
	conn, err := grpc.Dial(":8080", grpc.WithInsecure(), grpc.WithBlock())
	grpcLog.Info("Dialed 8080")
	if err != nil {
		grpcLog.Fatalf("Couldn't connect to service: %v", err)
	}

	arg1, _ := strconv.ParseInt(os.Args[1], 10, 32)
	bidder := &Bidder{id: int32(arg1), client: auction.NewAuctionClient(conn)}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		amount, err := strconv.ParseInt(scanner.Text(), 10, 32)
		if err != nil {
			continue
		}

		if amount == 0 {
			status, err := bidder.client.GetStatus(context.Background(), &auction.Empty{})
			if err != nil {
				grpcLog.Errorf(err.Error())
				continue
			}
			grpcLog.Infof("HighestBid thing: %d", status.HighestBid)
			continue
		}

		grpcLog.Info("Sending bid!")
		bidder.client.MakeBid(context.Background(), &auction.Bid{BidderId: bidder.id, Amount: int32(amount)})
	}
}
