package s3client

import (
	"context"
	blobpb "coolcar/blob/api/gen/v1"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"testing"
)

func TestCreateBlob(t *testing.T) {
	conn, err := grpc.Dial("localhost:8083", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	c := blobpb.NewBlobServiceClient(conn)
	ctx := context.Background()

	res, err := c.CreateBlob(ctx, &blobpb.CreateBlobRequest{
		AccountId:           "account1",
		UploadUrlTimeoutSec: 1000,
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v", res)
}

func TestGetBlob(t *testing.T) {
	conn, err := grpc.Dial("localhost:8083", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	c := blobpb.NewBlobServiceClient(conn)
	ctx := context.Background()

	res, err := c.GetBlob(ctx, &blobpb.GetBlobRequest{Id: "62a5368521800b019dc29309"})
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v", res)
}

func TestGetBlobURL(t *testing.T) {
	conn, err := grpc.Dial("localhost:8083", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	c := blobpb.NewBlobServiceClient(conn)
	ctx := context.Background()

	res, err := c.GetBlobURL(ctx, &blobpb.GetBlobURLRequest{
		Id:         "62a5368521800b019dc29309",
		TimeoutSec: 1000,
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v", res)
}
