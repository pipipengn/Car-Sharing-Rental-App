package main

import (
	"context"
	carpb "coolcar/car/api/gen/v1"
	"coolcar/car/car"
	"coolcar/car/dao"
	"coolcar/car/rabbitmq"
	"coolcar/car/simulation"
	"coolcar/car/simulation/position"
	"coolcar/car/ws"
	coolenvpb "coolcar/shared/carenv"
	"coolcar/shared/server"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net/http"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("cannot create logger: %v", err)
	}
	c := context.Background()

	// mongo
	mongoClient, err := mongo.Connect(c, options.Client().ApplyURI("mongodb://localhost:27017/coolcar?readPreference=primary&ssl=false"))
	if err != nil {
		logger.Fatal("cannot connect mongodb", zap.Error(err))
	}

	// rabbitmq
	rabbitconn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		logger.Fatal("cannot dial rabbitmq", zap.Error(err))
	}
	exchange := "coolcar"
	publisher, err := rabbitmq.NewPublisher(rabbitconn, exchange)
	if err != nil {
		logger.Fatal("cannot create publisher", zap.Error(err))
	}

	// car simulation
	carConn, err := grpc.Dial("localhost:8084", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("cannot connect car service", zap.Error(err))
	}
	subscriber, err := rabbitmq.NewSubscriber(rabbitconn, exchange, logger)
	if err != nil {
		logger.Fatal("cannot create subscriber", zap.Error(err))
	}
	aiConn, err := grpc.Dial("localhost:18001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("cannot connect ai service", zap.Error(err))
	}

	posSub, err := rabbitmq.NewSubscriber(rabbitconn, "pos_sim", logger)
	simulatorContorller := &simulation.Controller{
		Logger:        logger,
		CarService:    carpb.NewCarServiceClient(carConn),
		CarSubscriber: subscriber,
		PosSubscriber: &position.Subscriber{
			Sub:    posSub,
			Logger: logger,
		},
		AIService: coolenvpb.NewAIServiceClient(aiConn),
	}
	go simulatorContorller.RunSimulations(context.Background())

	// websocket
	r := gin.Default()
	r.GET("/ws", ws.NewHandler(ws.Options{
		Logger:        logger,
		CarSubscriber: subscriber,
		Upgrader: &websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}))
	go func() {
		addr := ":9090"
		logger.Info("Websocket Service started", zap.String("addr", addr))
		if err := r.Run(addr); err != nil {
			logger.Fatal("cannot create websocket", zap.Error(err))
		}
	}()

	// grpc
	logger.Sugar().Fatal(server.RunGRPCServer(&server.GRPCConfig{
		Name:   "car",
		Logger: logger,
		Addr:   ":8084",
		RegisterFunc: func(s *grpc.Server) {
			carpb.RegisterCarServiceServer(s, &car.Service{
				Logger:       logger,
				Mongo:        dao.NewMongo(mongoClient.Database("coolcar")),
				CarPublisher: publisher,
			})
		},
	}))
}
