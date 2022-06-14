package tripupdater

import (
	"context"
	carpb "coolcar/car/api/gen/v1"
	mq "coolcar/car/rabbitmq/mq_interface"
	rentalpb "coolcar/rental/api/gen/v1"
	"go.uber.org/zap"
)

type Options struct {
	Sub         mq.CarSubscriber
	TripService rentalpb.TripServiceClient
	Logger      *zap.Logger
}

func NewService(o Options) *Service {
	return &Service{
		sub:         o.Sub,
		tripService: o.TripService,
		logger:      o.Logger,
	}
}

type Service struct {
	sub         mq.CarSubscriber
	tripService rentalpb.TripServiceClient
	logger      *zap.Logger
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
			if _, err := s.tripService.UpdateTrip(context.Background(), &rentalpb.UpdateTripRequest{
				Id: tripID,
				Current: &rentalpb.Location{
					Latitude:  carPos.Latitude,
					Longitude: carPos.Longitude,
				},
			}); err != nil {
				s.logger.Error("cannot update trip", zap.String("trip_id", tripID), zap.Error(err))
			}
		}
	}
}
