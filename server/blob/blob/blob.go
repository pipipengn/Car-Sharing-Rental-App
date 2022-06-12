package blob

import (
	"context"
	"coolcar/blob/api/gen/v1"
	"coolcar/blob/dao"
	"coolcar/shared/id"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"io/ioutil"
	"time"
)

type Service struct {
	blobpb.UnimplementedBlobServiceServer
	Logger  *zap.Logger
	Mongo   *dao.Mongo
	Storage Storage
}

type Storage interface {
	PutSignURL(c context.Context, key string, timeout time.Duration) (string, error)
	GetSignURL(c context.Context, key string, timeout time.Duration) (string, error)
	Get(c context.Context, key string) (io.ReadCloser, error)
}

func (s *Service) CreateBlob(c context.Context, req *blobpb.CreateBlobRequest) (*blobpb.CreateBlobResponse, error) {
	aid := id.AccountID(req.AccountId)
	br, err := s.Mongo.CreateBlob(c, aid)
	if err != nil {
		s.Logger.Error("cannot create blob", zap.Error(err))
		return nil, status.Error(codes.Internal, "")
	}
	signURL, err := s.Storage.PutSignURL(c, br.Path, secToDuration(req.UploadUrlTimeoutSec))
	if err != nil {
		return nil, status.Errorf(codes.Aborted, "cannot sign url: %v", err)
	}
	return &blobpb.CreateBlobResponse{
		Id:        br.ID.Hex(),
		UploadUrl: signURL,
	}, nil
}

func (s *Service) GetBlob(c context.Context, req *blobpb.GetBlobRequest) (*blobpb.GetBlobResponse, error) {
	br, err := s.getBlobRecord(c, id.BlobID(req.Id))
	if err != nil {
		return nil, err
	}

	readCloser, err := s.Storage.Get(c, br.Path)
	if readCloser != nil {
		defer func(readCloser io.ReadCloser) {
			err := readCloser.Close()
			if err != nil {

			}
		}(readCloser)
	}
	if err != nil {
		return nil, status.Errorf(codes.Aborted, "cannot get storage: %v", err)
	}

	bytes, err := ioutil.ReadAll(readCloser)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, "cannot read from response: %v", err)
	}

	return &blobpb.GetBlobResponse{Data: bytes}, nil
}

func (s *Service) GetBlobURL(c context.Context, req *blobpb.GetBlobURLRequest) (*blobpb.GetBlobURLResponse, error) {
	br, err := s.getBlobRecord(c, id.BlobID(req.Id))
	if err != nil {
		return nil, err
	}

	signURL, err := s.Storage.GetSignURL(c, br.Path, secToDuration(req.TimeoutSec))
	if err != nil {
		return nil, status.Errorf(codes.Aborted, "cannot sign url: %v", err)
	}

	return &blobpb.GetBlobURLResponse{Url: signURL}, nil
}

func (s *Service) getBlobRecord(c context.Context, bid id.BlobID) (*dao.BlobRecord, error) {
	blobRecord, err := s.Mongo.GetBlob(c, bid)
	if err == mongo.ErrNoDocuments {
		return nil, status.Error(codes.NotFound, "")
	}
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return blobRecord, nil
}

func secToDuration(sec int32) time.Duration {
	return time.Duration(sec) * time.Second
}
