package s3client

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
	"time"
)

type Service struct {
	Client *s3.Client
	Bucket string
}

func NewS3Service(accessKeyId, secretAccessKey, region, bucket string) *Service {
	client := s3.New(s3.Options{
		Region:      region,
		Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(accessKeyId, secretAccessKey, "")),
	})
	return &Service{Client: client, Bucket: bucket}
}

func (s *Service) PutSignURL(c context.Context, key string, timeout time.Duration) (string, error) {
	psClient := s3.NewPresignClient(s.Client, s3.WithPresignExpires(timeout))
	object, err := psClient.PresignPutObject(c, &s3.PutObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", fmt.Errorf("got an error retrieving pre-signed put object: %v", err)
	}
	return object.URL, nil
}

func (s *Service) GetSignURL(c context.Context, key string, timeout time.Duration) (string, error) {
	psClient := s3.NewPresignClient(s.Client, s3.WithPresignExpires(timeout))
	object, err := psClient.PresignGetObject(c, &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", fmt.Errorf("got an error retrieving pre-signed get object: %v", err)
	}
	return object.URL, nil
}

func (s *Service) Get(c context.Context, key string) (io.ReadCloser, error) {
	response, err := s.Client.GetObject(c, &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("got an error get object: %v", err)
	}
	return response.Body, nil
}
