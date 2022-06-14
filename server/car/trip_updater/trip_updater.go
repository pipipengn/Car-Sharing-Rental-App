package tripupdater

import (
	"context"
	carpb "coolcar/car/api/gen/v1"
	mq "coolcar/car/rabbitmq/mq_interface"
	r "coolcar/car/redis"
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/shared/auth"
	"coolcar/shared/id"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"time"
)

const redisLockField = "lock"

type Options struct {
	Sub             mq.CarSubscriber
	TripService     rentalpb.TripServiceClient
	Logger          *zap.Logger
	Redis           *r.RedisService
	RedisExpireTime time.Duration
}

func NewService(o *Options) *Service {
	return &Service{
		sub:             o.Sub,
		tripService:     o.TripService,
		logger:          o.Logger,
		redis:           o.Redis,
		redisExpireTime: o.RedisExpireTime,
	}
}

type Service struct {
	sub             mq.CarSubscriber
	tripService     rentalpb.TripServiceClient
	logger          *zap.Logger
	redis           *r.RedisService
	redisExpireTime time.Duration
}

func (s *Service) RunUpdator() {
	ch, closeFunc, err := s.sub.Subscribe(context.Background())
	defer closeFunc()
	if err != nil {
		s.logger.Error("cannot subscribe", zap.Error(err))
		return
	}

	for carEntity := range ch {
		tripID := carEntity.Car.TripId
		carPos := carEntity.Car.Position
		if carEntity.Car.Status == carpb.CarStatus_UNLOCKED && tripID != "" && carEntity.Car.Driver.Id != "" {
			if _, err := s.redis.Get(redisLockField); err != redis.Nil {
				continue
			}

			if _, err := s.tripService.UpdateTrip(context.Background(), &rentalpb.UpdateTripRequest{
				Id: tripID,
				Current: &rentalpb.Location{
					Latitude:  carPos.Latitude,
					Longitude: carPos.Longitude,
				},
			}, grpc.PerRPCCredentials(&impersonation{AccountID: id.AccountID(carEntity.Car.Driver.Id)})); err != nil {
				s.logger.Error("cannot update trip", zap.String("trip_id", tripID), zap.Error(err))
			}

			for {
				if err := s.redis.Set(redisLockField, true, s.redisExpireTime); err != nil {
					continue
				}
				break
			}
		}
	}
}

type impersonation struct {
	AccountID id.AccountID
}

func (i *impersonation) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		auth.ImpersonateAccountHeader: i.AccountID.String(),
	}, nil
}

func (i *impersonation) RequireTransportSecurity() bool {
	return false
}
