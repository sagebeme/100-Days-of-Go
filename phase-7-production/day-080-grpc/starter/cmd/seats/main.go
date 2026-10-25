// Command seats runs the Twende seat service on localhost:50051 with two trips.
// Server reflection is on, so grpcurl can list and call it:
//
//	grpcurl -plaintext localhost:50051 list
//	grpcurl -plaintext -d '{"trip_id":"46K-0730"}' localhost:50051 twende.seats.v1.SeatService/GetTrip
package main

import (
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/sagebeme/100-Days-of-Go/phase-7-production/day-080-grpc/seatpb"
	seats "github.com/sagebeme/100-Days-of-Go/phase-7-production/day-080-grpc/starter"
)

func main() {
	lis, err := net.Listen("tcp", "localhost:50051")
	if err != nil {
		log.Fatal(err)
	}
	gs := grpc.NewServer()
	seatpb.RegisterSeatServiceServer(gs, seats.NewServer(
		&seatpb.Trip{Id: "46K-0730", Route: "46K Kencom to Kawangware", Departs: "07:30", Seats: 14},
		&seatpb.Trip{Id: "111-0800", Route: "111 Railways to Ngong", Departs: "08:00", Seats: 33},
	))
	reflection.Register(gs)
	log.Println("seat service on localhost:50051")
	log.Fatal(gs.Serve(lis))
}
