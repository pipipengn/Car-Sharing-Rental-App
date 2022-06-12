package server

import (
	"coolcar/shared/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
)

type GRPCConfig struct {
	Name              string
	Logger            *zap.Logger
	Addr              string
	AuthPublicKeyFile string
	RegisterFunc      func(*grpc.Server)
}

func RunGRPCServer(c *GRPCConfig) error {
	nameField := zap.String("name", c.Name)

	var opts []grpc.ServerOption
	if c.AuthPublicKeyFile != "" {
		interceptor, err := auth.NewInterceptor(c.AuthPublicKeyFile)
		if err != nil {
			c.Logger.Fatal("cannot create auth intercepter", nameField)
		}
		opts = append(opts, grpc.UnaryInterceptor(interceptor))
	}

	s := grpc.NewServer(opts...)
	c.RegisterFunc(s)

	lis, err := net.Listen("tcp", c.Addr)
	if err != nil {
		c.Logger.Fatal("cannot listen", nameField, zap.Error(err))
	}

	c.Logger.Info("server started", nameField, zap.String("addr", c.Addr))
	return s.Serve(lis)
}
