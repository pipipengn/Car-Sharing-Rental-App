package main

import (
	"context"
	blobpb "coolcar/blob/api/gen/v1"
	"coolcar/blob/blob"
	"coolcar/blob/dao"
	s3client "coolcar/blob/s3"
	"coolcar/shared/server"
	"github.com/namsral/flag"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"log"
)

var addr = flag.String("addr", ":8083", "address to listen")
var mongoURI = flag.String("mongo_uri", "mongodb://localhost:27017/coolcar?readPreference=primary&ssl=false", "mongo uri")
var accessKeyId = flag.String("aws_sec_id", "<SEC_ID>", "aws secret id")
var secretAccessKey = flag.String("aws_sec_key", "<SEC_KEY>", "aws secret key")
var region = flag.String("aws_region", "us-west-2", "aws region")
var bucketName = flag.String("s3_bucket_name", "coolcar", "s3 bucket name")

func main() {
	flag.Parse()

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("cannot create logger: %v", err)
	}

	c := context.Background()
	mongoClient, err := mongo.Connect(c, options.Client().ApplyURI(*mongoURI))
	if err != nil {
		logger.Fatal("cannot connect mongodb", zap.Error(err))
	}

	logger.Sugar().Fatal(server.RunGRPCServer(&server.GRPCConfig{
		Name:   "blob",
		Logger: logger,
		Addr:   *addr,
		RegisterFunc: func(s *grpc.Server) {
			blobpb.RegisterBlobServiceServer(s, &blob.Service{
				Logger:  logger,
				Mongo:   dao.NewMongo(mongoClient.Database("coolcar")),
				Storage: s3client.NewS3Service(*accessKeyId, *secretAccessKey, *region, *bucketName),
			})
		},
	}))
}
