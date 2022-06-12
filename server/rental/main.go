package main

import (
	"context"
	blobpb "coolcar/blob/api/gen/v1"
	"coolcar/rental/ai"
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/rental/profile"
	profiledao "coolcar/rental/profile/dao"
	"coolcar/rental/trip"
	"coolcar/rental/trip/client/car"
	"coolcar/rental/trip/client/poi"
	pf "coolcar/rental/trip/client/profile"
	tripdao "coolcar/rental/trip/dao"
	coolenvpb "coolcar/shared/carenv"
	"coolcar/shared/server"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("cannot create logger: %v", err)
	}

	c := context.Background()
	mongoClient, err := mongo.Connect(c, options.Client().ApplyURI("mongodb://localhost:27017/coolcar?readPreference=primary&ssl=false"))
	if err != nil {
		logger.Fatal("cannot connect mongodb", zap.Error(err))
	}

	aconn, err := grpc.Dial("localhost:18001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("cannot connect aiservice", zap.Error(err))
	}

	bconn, err := grpc.Dial("localhost:8083", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("cannot connect blob service", zap.Error(err))
	}

	db := mongoClient.Database("coolcar")
	logger.Sugar().Fatal(server.RunGRPCServer(&server.GRPCConfig{
		Name:              "rental",
		Logger:            logger,
		Addr:              ":8082",
		AuthPublicKeyFile: "shared/auth/public.key",
		RegisterFunc: func(s *grpc.Server) {
			profileService := &profile.Service{
				Mongo:          profiledao.NewMongo(db),
				Logger:         logger,
				BlobClient:     blobpb.NewBlobServiceClient(bconn),
				PhotoGetExpire: 500 * time.Second,
				PhotoPutExpire: 500 * time.Second,
			}
			rentalpb.RegisterTripServiceServer(s, &trip.Service{
				Logger:         logger,
				ProfileManager: &pf.Manager{Fetcher: profileService},
				CarManager:     &car.Manager{},
				POIManager:     &poi.Manager{},
				Mongo:          tripdao.NewMongo(db),
				DistanceCalc:   ai.Client{AIClient: coolenvpb.NewAIServiceClient(aconn)},
			})
			rentalpb.RegisterProfileServiceServer(s, profileService)
		},
	}))
}
