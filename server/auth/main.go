package main

import (
	"context"
	authpb "coolcar/auth/api/gen/v1"
	"coolcar/auth/auth"
	"coolcar/auth/dao"
	"coolcar/auth/token"
	"coolcar/auth/wx"
	"coolcar/shared/server"
	"github.com/namsral/flag"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"log"
	"time"
)

var addr = flag.String("addr", ":8081", "address to listen")
var mongoURI = flag.String("mongo_uri", "mongodb://localhost:27017/coolcar?readPreference=primary&ssl=false", "mongo uri")
var privateKeyFile = flag.String("private_key_file", "auth/token/private.key", "private key file")
var wechatAppID = flag.String("wechat_app_id", "<APPID>", "wechat app id")
var wechatAppSecret = flag.String("wechat_app_secret", "<APPSERET>", "wechat app secret")

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
		Name:   "auth",
		Logger: logger,
		Addr:   *addr,
		RegisterFunc: func(s *grpc.Server) {
			authpb.RegisterAuthServiceServer(s, &auth.Service{
				Logger: logger,
				OpenIDResolver: &wx.Server{
					AppID:     *wechatAppID,
					AppSecret: *wechatAppSecret,
				},
				Mongo:          dao.NewMongo(mongoClient.Database("coolcar")),
				TokenExpire:    750 * time.Hour,
				TokenGenerator: token.NewJWTTokenGen("coolcar/auth", token.PrivateKey(*privateKeyFile)),
			})
		},
	}))
}
