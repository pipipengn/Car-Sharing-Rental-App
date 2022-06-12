package profile

import (
	"context"
	blobpb "coolcar/blob/api/gen/v1"
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/rental/profile/dao"
	"coolcar/shared/auth"
	"coolcar/shared/id"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

type Service struct {
	rentalpb.UnimplementedProfileServiceServer
	Mongo          *dao.Mongo
	Logger         *zap.Logger
	BlobClient     blobpb.BlobServiceClient
	PhotoGetExpire time.Duration
	PhotoPutExpire time.Duration
}

func (s *Service) GetProfile(c context.Context, request *rentalpb.GetProfileRequest) (*rentalpb.Profile, error) {
	aid, err := auth.AccountIDFromContext(c)
	if err != nil {
		return nil, err
	}
	profileRecord, err := s.Mongo.GetProfile(c, aid)
	if err != nil {
		code := s.convertProfileErr(err)
		if code == codes.NotFound {
			return &rentalpb.Profile{}, nil
		}
		return nil, status.Error(code, "")
	}

	if profileRecord.Profile == nil {
		return &rentalpb.Profile{}, nil
	}
	return profileRecord.Profile, nil
}

func (s *Service) SubmitProfile(c context.Context, identity *rentalpb.Identity) (*rentalpb.Profile, error) {
	aid, err := auth.AccountIDFromContext(c)
	if err != nil {
		return nil, err
	}

	p := &rentalpb.Profile{
		Identity:       identity,
		IdentityStatus: rentalpb.IdentityStatus_PENDING,
	}
	err = s.Mongo.UpdateProfile(c, aid, p, rentalpb.IdentityStatus_UNSUBMITTED)
	if err != nil {
		s.Logger.Error("cannot create profile", zap.Error(err))
		return nil, status.Error(codes.Internal, "")
	}

	// verify
	go func() {
		time.Sleep(10 * time.Second)
		p := &rentalpb.Profile{
			Identity:       identity,
			IdentityStatus: rentalpb.IdentityStatus_VERIFIED,
		}
		err = s.Mongo.UpdateProfile(context.Background(), aid, p, rentalpb.IdentityStatus_PENDING)
		if err != nil {
			s.Logger.Error("cannot verify profile", zap.Error(err))
		}
	}()
	return p, nil
}

func (s *Service) ClearProfile(c context.Context, request *rentalpb.ClearProfileRequest) (*rentalpb.Profile, error) {
	aid, err := auth.AccountIDFromContext(c)
	if err != nil {
		return nil, err
	}

	p := &rentalpb.Profile{}
	err = s.Mongo.UpdateProfile(c, aid, p, rentalpb.IdentityStatus_VERIFIED)
	if err != nil {
		s.Logger.Error("cannot clear profile", zap.Error(err))
		return nil, status.Error(codes.Internal, "")
	}
	return p, nil
}

func (s *Service) GetProfilePhoto(c context.Context, request *rentalpb.GetProfilePhotoRequest) (*rentalpb.GetProfilePhotoResponse, error) {
	aid, err := auth.AccountIDFromContext(c)
	if err != nil {
		return nil, err
	}

	profileRecord, err := s.Mongo.GetProfile(c, aid)
	if err != nil {
		return nil, status.Error(s.convertProfileErr(err), "")
	}
	if profileRecord.PhotoBLobID == "" {
		return nil, status.Error(codes.NotFound, "")
	}

	blobResp, err := s.BlobClient.GetBlobURL(c, &blobpb.GetBlobURLRequest{
		Id:         profileRecord.PhotoBLobID,
		TimeoutSec: int32(s.PhotoGetExpire.Seconds()),
	})
	if err != nil {
		s.Logger.Error("cannot get blob", zap.Error(err))
		return nil, status.Error(codes.Internal, "")
	}

	return &rentalpb.GetProfilePhotoResponse{Url: blobResp.Url}, nil
}

func (s *Service) CreateProfilePhoto(c context.Context, request *rentalpb.CreateProfilePhotoRequest) (*rentalpb.CreateProfilePhotoResponse, error) {
	aid, err := auth.AccountIDFromContext(c)
	if err != nil {
		return nil, err
	}

	blobResp, err := s.BlobClient.CreateBlob(c, &blobpb.CreateBlobRequest{
		AccountId:           aid.String(),
		UploadUrlTimeoutSec: int32(s.PhotoPutExpire.Seconds()),
	})
	if err != nil {
		s.Logger.Error("cannot create blob", zap.Error(err))
		return nil, status.Error(codes.Aborted, "")
	}

	if err = s.Mongo.UpdateProfilePhoto(c, aid, id.BlobID(blobResp.Id)); err != nil {
		s.Logger.Error("cannot update profile photo", zap.Error(err))
		return nil, status.Error(codes.Aborted, "")
	}

	return &rentalpb.CreateProfilePhotoResponse{UploadUrl: blobResp.UploadUrl}, nil
}

func (s *Service) CompleteProfilePhoto(c context.Context, request *rentalpb.CompleteProfilePhotoRequest) (*rentalpb.Identity, error) {
	aid, err := auth.AccountIDFromContext(c)
	if err != nil {
		return nil, err
	}

	profileRecord, err := s.Mongo.GetProfile(c, aid)
	if err != nil {
		return nil, status.Error(s.convertProfileErr(err), "")
	}
	if profileRecord.PhotoBLobID == "" {
		return nil, status.Error(codes.NotFound, "")
	}

	blobResp, err := s.BlobClient.GetBlob(c, &blobpb.GetBlobRequest{Id: profileRecord.PhotoBLobID})
	if err != nil {
		s.Logger.Error("cannot get blob", zap.Error(err))
		return nil, status.Error(codes.Aborted, "")
	}

	// TODO use image recognition API and return image recognition result
	_ = blobResp.Data
	return &rentalpb.Identity{
		LicNumber:       "21321343535",
		Name:            "pipipengn",
		Gender:          rentalpb.Gender_MALE,
		BirthDateMillis: 631250000000,
	}, nil
}

func (s *Service) ClearProfilePhoto(c context.Context, request *rentalpb.ClearProfilePhotoRequest) (*rentalpb.ClearProfilePhotoResponse, error) {
	aid, err := auth.AccountIDFromContext(c)
	if err != nil {
		return nil, err
	}

	err = s.Mongo.UpdateProfilePhoto(c, aid, "")
	if err != nil {
		return nil, status.Error(codes.Internal, "")
	}
	return &rentalpb.ClearProfilePhotoResponse{}, nil
}

func (s *Service) convertProfileErr(err error) codes.Code {
	if err == mongo.ErrNoDocuments {
		return codes.NotFound
	}
	s.Logger.Error("cannot get profile", zap.Error(err))
	return codes.Internal
}
