package main

import (
	"context"
	"coolcar/car/rabbitmq"
	coolenvpb "coolcar/shared/carenv"
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"time"
)

func main() {
	conn, err := grpc.Dial("localhost:18001", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	ac := coolenvpb.NewAIServiceClient(conn)
	c := context.Background()

	res, err := ac.MeasureDistance(c, &coolenvpb.MeasureDistanceRequest{
		From: &coolenvpb.Location{
			Latitude:  30,
			Longitude: 120,
		},
		To: &coolenvpb.Location{
			Latitude:  31,
			Longitude: 121,
		},
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", res)

	logger, _ := zap.NewDevelopment()
	rabbitconn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")

	_, err = ac.SimulateCarPos(c, &coolenvpb.SimulateCarPosRequest{
		CarId: "car123",
		Type:  coolenvpb.PosType_RANDOM,
		InitialPos: &coolenvpb.Location{
			Latitude:  30,
			Longitude: 120,
		},
	})
	if err != nil {
		panic(err)
	}

	subscriber, err := rabbitmq.NewSubscriber(rabbitconn, "pos_sim", logger)
	ch, f, _ := subscriber.SubscribeRaw(c)
	defer f()

	after := time.After(10 * time.Second)
	for {
		needBreak := false
		select {
		case msg := <-ch:
			var update coolenvpb.CarPosUpdate
			err := json.Unmarshal(msg.Body, &update)
			if err != nil {
				panic(err)
			}
			fmt.Printf("%+v\n", &update)
		case <-after:
			needBreak = true
		}
		if needBreak {
			break
		}
	}

	_, _ = ac.EndSimulateCarPos(c, &coolenvpb.EndSimulateCarPosRequest{CarId: "car123"})
}
