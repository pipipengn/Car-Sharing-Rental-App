package main

import (
	"context"
	carpb "coolcar/car/api/gen/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.Dial("localhost:8084", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	cs := carpb.NewCarServiceClient(conn)
	c := context.Background()

	//add 5 cars
	//for i := 0; i < 5; i++ {
	//	carEntity, err := cs.CreateCar(context.Background(), &carpb.CreateCarRequest{})
	//	if err != nil {
	//		panic(err)
	//	}
	//	fmt.Printf("create car: %s", carEntity.Id)
	//}

	// reset
	res, _ := cs.GetCars(c, &carpb.GetCarsRequest{})
	for _, car := range res.Cars {
		_, _ = cs.UpdateCar(c, &carpb.UpdateCarRequest{
			Id:     car.Id,
			Status: carpb.CarStatus_LOCKED,
		})
	}
}
