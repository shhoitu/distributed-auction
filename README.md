# Distributed auction system

protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative grpc/interface.proto

# How to run the program

Set up 3 replica managers first by runnig them in seperate windows using go run main.go in the server folder  
Then open a frontend calling og run main.go in root  
lastly connect as many clients as you want by calling go run main.go in the client folder  
