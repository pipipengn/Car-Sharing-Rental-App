package main

import (
	"context"
	blobpb "coolcar/blob/api/gen/v1"
	carpb "coolcar/car/api/gen/v1"
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
	"github.com/namsral/flag"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"
)

var addr = flag.String("addr", ":8082", "address to listen")
var mongoURI = flag.String("mongo_uri", "mongodb://localhost:27017/coolcar?readPreference=primary&ssl=false", "mongo uri")
var blobAddr = flag.String("blob_addr", "localhost:8083", "address for blob service")
var aiAddr = flag.String("ai_addr", "localhost:18001", "address for ai service")
var carAddr = flag.String("car_addr", "localhost:8084", "address for car service")
var authPublicKeyFile = flag.String("auth_public_key_file", "shared/auth/public.key", "public key file for auth")

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("cannot create logger: %v", err)
	}

	c := context.Background()
	mongoClient, err := mongo.Connect(c, options.Client().ApplyURI(*mongoURI))
	if err != nil {
		logger.Fatal("cannot connect mongodb", zap.Error(err))
	}

	aconn, err := grpc.Dial(*aiAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("cannot connect aiservice", zap.Error(err))
	}

	bconn, err := grpc.Dial(*blobAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("cannot connect blob service", zap.Error(err))
	}

	cconn, err := grpc.Dial(*carAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("cannot connect car service", zap.Error(err))
	}

	db := mongoClient.Database("coolcar")
	logger.Sugar().Fatal(server.RunGRPCServer(&server.GRPCConfig{
		Name:              "rental",
		Logger:            logger,
		Addr:              *addr,
		AuthPublicKeyFile: *authPublicKeyFile,
		RegisterFunc: func(s *grpc.Server) {
			profileService := &profile.Service{
				Mongo:          profiledao.NewMongo(db),
				Logger:         logger,
				BlobClient:     blobpb.NewBlobServiceClient(bconn),
				PhotoGetExpire: 500 * time.Second,
				PhotoPutExpire: 500 * time.Second,
			}
			tripService := &trip.Service{
				Logger:         logger,
				ProfileManager: &pf.Manager{Fetcher: profileService},
				CarManager:     &car.Manager{CarService: carpb.NewCarServiceClient(cconn)},
				POIManager:     &poi.Manager{},
				Mongo:          tripdao.NewMongo(db),
				DistanceCalc:   ai.Client{AIClient: coolenvpb.NewAIServiceClient(aconn)},
			}
			rentalpb.RegisterTripServiceServer(s, tripService)
			rentalpb.RegisterProfileServiceServer(s, profileService)
		},
	}))
}
