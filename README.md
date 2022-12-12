# Distributed auction system

protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative grpc/interface.proto

# How to run the program
Make sure you are in the root project folder.
Use the following commands (in seperate shells) to setup the 3 replica managers:
- go run server/main.go 0
- go run server/main.go 1
- go run server/main.go 2

Then to setup a client/frontend pair, use the following commands (again in seperate shells):
- go run main.go 0
- go run client/main.go 0
To setup more client/frontend pairs replace the 0 in the two commands above with another number (1, 2, 3, and so on).

For a client to make a bid, enter the amount (e.g. 1000000000).
For a client to get the current status of the auction, enter 0.
