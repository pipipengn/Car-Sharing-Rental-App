package main

import (
	"context"
	authpb "coolcar/auth/api/gen/v1"
	carpb "coolcar/car/api/gen/v1"
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/shared/auth"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/namsral/flag"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net/http"
	"net/textproto"
)

var addr = flag.String("addr", ":8080", "address to listen")
var authAddr = flag.String("auth_addr", "localhost:8081", "address for auth service")
var tripAddr = flag.String("trip_addr", "localhost:8082", "address for trip service")
var profileAddr = flag.String("profile_addr", "localhost:8082", "address for profile service")
var carAddr = flag.String("car_addr", "localhost:8084", "address for car service")

func main() {
	flag.Parse()

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("cannot create logger: %v", err)
	}

	c := context.Background()
	c, cancel := context.WithCancel(c)
	defer cancel()

	mux := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		OrigName:    true,
		EnumsAsInts: true,
	}), runtime.WithIncomingHeaderMatcher(func(key string) (string, bool) {
		if key == textproto.CanonicalMIMEHeaderKey(runtime.MetadataHeaderPrefix+auth.ImpersonateAccountHeader) {
			return "", false
		}
		return runtime.DefaultHeaderMatcher(key)
	}))

	serverCongig := []struct {
		name         string
		addr         string
		registerFunc func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error)
	}{
		{
			name:         "auth",
			addr:         *authAddr,
			registerFunc: authpb.RegisterAuthServiceHandlerFromEndpoint,
		},
		{
			name:         "trip",
			addr:         *tripAddr,
			registerFunc: rentalpb.RegisterTripServiceHandlerFromEndpoint,
		},
		{
			name:         "profile",
			addr:         *profileAddr,
			registerFunc: rentalpb.RegisterProfileServiceHandlerFromEndpoint,
		},
		{
			name:         "car",
			addr:         *carAddr,
			registerFunc: carpb.RegisterCarServiceHandlerFromEndpoint,
		},
	}

	credentials := grpc.WithTransportCredentials(insecure.NewCredentials())
	for _, s := range serverCongig {
		if err := s.registerFunc(c, mux, s.addr, []grpc.DialOption{credentials}); err != nil {
			logger.Sugar().Fatalf("cannot register %s service: %v", s.name, err)
		}
	}

	logger.Sugar().Infof("grpc gateway started at %s", *addr)
	logger.Sugar().Fatal(http.ListenAndServe(*addr, mux))
}
