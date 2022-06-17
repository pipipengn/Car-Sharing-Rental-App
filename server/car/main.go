package main

import (
	"context"
	carpb "coolcar/car/api/gen/v1"
	"coolcar/car/car"
	"coolcar/car/dao"
	"coolcar/car/rabbitmq"
	r2 "coolcar/car/redis"
	"coolcar/car/simulation"
	"coolcar/car/simulation/position"
	tripupdater "coolcar/car/trip_updater"
	"coolcar/car/ws"
	rentalpb "coolcar/rental/api/gen/v1"
	coolenvpb "coolcar/shared/carenv"
	"coolcar/shared/server"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/namsral/flag"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net/http"
	"time"
)

var addr = flag.String("addr", ":8084", "address to listen")
var wsAddr = flag.String("ws_addr", ":9090", "websocket address to listen")
var mongoURI = flag.String("mongo_uri", "mongodb://localhost:27017/coolcar?readPreference=primary&ssl=false", "mongo uri")
var amqpURL = flag.String("amqp_url", "amqp://guest:guest@localhost:5672/", "amqp url")
var carAddr = flag.String("car_addr", "localhost:8084", "address for car service")
var tripAddr = flag.String("trip_addr", "localhost:8082", "address for trip service")
var aiAddr = flag.String("ai_addr", "localhost:18001", "address for ai service")
var redisAddr = flag.String("redis_addr", "localhost:6379", "address for redis service")

func main() {
	flag.Parse()

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("cannot create logger: %v", err)
	}
	c := context.Background()

	// mongo
	mongoClient, err := mongo.Connect(c, options.Client().ApplyURI(*mongoURI))
	if err != nil {
		logger.Fatal("cannot connect mongodb", zap.Error(err))
	}

	// rabbitmq
	rabbitconn, err := amqp.Dial(*amqpURL)
	if err != nil {
		logger.Fatal("cannot dial rabbitmq", zap.Error(err))
	}
	exchange := "coolcar"
	publisher, err := rabbitmq.NewPublisher(rabbitconn, exchange)
	if err != nil {
		logger.Fatal("cannot create publisher", zap.Error(err))
	}

	// car simulation
	carConn, err := grpc.Dial(*carAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("cannot connect car service", zap.Error(err))
	}
	subscriber, err := rabbitmq.NewSubscriber(rabbitconn, exchange, logger)
	if err != nil {
		logger.Fatal("cannot create subscriber", zap.Error(err))
	}
	aiConn, err := grpc.Dial(*aiAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
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

	// trip updater
	tripConn, err := grpc.Dial(*tripAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("cannot connect trip service", zap.Error(err))
	}
	tuService := tripupdater.NewService(&tripupdater.Options{
		Sub:             subscriber,
		TripService:     rentalpb.NewTripServiceClient(tripConn),
		Logger:          logger,
		Redis:           r2.NewRedisService(*redisAddr),
		RedisExpireTime: 3 * time.Second,
	})
	go tuService.RunUpdator()

	// websocket
	r := gin.Default()
	r.GET("/ws", ws.NewHandler(&ws.Options{
		Logger:        logger,
		CarSubscriber: subscriber,
		Upgrader: &websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}))
	go func() {
		addr := *wsAddr
		logger.Info("Websocket Service started", zap.String("addr", addr))
		if err := r.Run(addr); err != nil {
			logger.Fatal("cannot create websocket", zap.Error(err))
		}
	}()

	// grpc
	logger.Sugar().Fatal(server.RunGRPCServer(&server.GRPCConfig{
		Name:   "car",
		Logger: logger,
		Addr:   *addr,
		RegisterFunc: func(s *grpc.Server) {
			carpb.RegisterCarServiceServer(s, &car.Service{
				Logger:       logger,
				Mongo:        dao.NewMongo(mongoClient.Database("coolcar")),
				CarPublisher: publisher,
			})
		},
	}))
}
