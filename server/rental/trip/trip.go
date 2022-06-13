package trip

import (
	"context"
	"coolcar/rental/api/gen/v1"
	"coolcar/rental/trip/dao"
	"coolcar/shared/auth"
	"coolcar/shared/id"
	"coolcar/shared/mongo/objid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

type Service struct {
	rentalpb.UnimplementedTripServiceServer
	Logger         *zap.Logger
	Mongo          *dao.Mongo
	ProfileManager ProfileManager
	CarManager     CarManager
	POIManager     POIManager
	DistanceCalc   DistanceCalc
}

// ProfileManager : ACL : Anti Corruption Layer
type ProfileManager interface {
	Verify(context.Context, id.AccountID) (id.IdentityID, error)
}

// CarManager : ACL : Anti Corruption Layer
type CarManager interface {
	Verify(context.Context, id.CarID, *rentalpb.Location) error
	Unlock(context.Context, id.CarID, id.AccountID, id.TripID) error
	Lock(context.Context, id.CarID) error
}

// POIManager : resolve point of interest
type POIManager interface {
	Resolve(context.Context, *rentalpb.Location) (string, error)
}

func (s *Service) CreateTrip(ctx context.Context, request *rentalpb.CreateTripRequest) (*rentalpb.TripEntity, error) {
	aid, err := auth.AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Verify driver identity
	iID, err := s.ProfileManager.Verify(ctx, aid)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, err.Error())
	}

	// check vehicle status
	carID := id.CarID(request.CarId)
	if err = s.CarManager.Verify(ctx, carID, request.Start); err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	// Create itinerary: write to database, unlock and charge
	ls := s.calcCurrentStatus(ctx, &rentalpb.LocationStatus{
		Location:     request.Start,
		TimestampSec: nowFunc(),
	}, request.Start)

	tripRecord, err := s.Mongo.CreateTrip(ctx, &rentalpb.Trip{
		AccountId:  aid.String(),
		CarId:      carID.String(),
		Start:      ls,
		Current:    ls,
		Status:     rentalpb.TripStatus_IN_PROGRESS,
		IdentityId: iID.String(),
	})
	if err != nil {
		s.Logger.Warn("cannot create trip", zap.Error(err))
		return nil, status.Error(codes.Internal, "")
	}

	go func() {
		if err = s.CarManager.Unlock(context.Background(), carID, aid, objid.ToTripID(tripRecord.ID)); err != nil {
			s.Logger.Error("cannot unlock car", zap.Error(err))
		}
	}()

	return &rentalpb.TripEntity{
		Id:   tripRecord.ID.Hex(),
		Trip: tripRecord.Trip,
	}, nil
}

func (s *Service) GetTrip(ctx context.Context, request *rentalpb.GetTripRequest) (*rentalpb.Trip, error) {
	aid, err := auth.AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	tripRecord, err := s.Mongo.GetTrip(ctx, id.TripID(request.Id), aid)
	if err != nil {
		return nil, status.Error(codes.NotFound, "")
	}
	return tripRecord.Trip, nil
}

func (s *Service) GetTrips(ctx context.Context, request *rentalpb.GetTripsRequest) (*rentalpb.GetTripsResponse, error) {
	aid, err := auth.AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	tripRecords, err := s.Mongo.GetTrips(ctx, aid, request.Status)
	if err != nil {
		s.Logger.Error("cannot get trips", zap.String("error", err.Error()))
		return nil, status.Error(codes.Internal, "")
	}

	var trips []*rentalpb.TripEntity
	for _, tripRecord := range tripRecords {
		trips = append(trips, &rentalpb.TripEntity{
			Id:   tripRecord.ID.Hex(),
			Trip: tripRecord.Trip,
		})
	}
	return &rentalpb.GetTripsResponse{Trips: trips}, nil
}

func (s *Service) UpdateTrip(ctx context.Context, request *rentalpb.UpdateTripRequest) (*rentalpb.Trip, error) {
	aid, err := auth.AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if request.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "")
	}

	tid := id.TripID(request.Id)
	trip, err := s.Mongo.GetTrip(ctx, tid, aid)
	if err != nil {
		return nil, status.Error(codes.NotFound, "")
	}
	if trip.Trip.Status == rentalpb.TripStatus_FINISHED {
		return nil, status.Error(codes.FailedPrecondition, "cannot update a finished trip")
	}
	if trip.Trip.Current == nil {
		s.Logger.Error("trip without current location", zap.String("id", tid.String()))
		return nil, status.Error(codes.Internal, "")
	}

	// start update value
	curLocation := trip.Trip.Current.Location
	if request.Current != nil {
		curLocation = request.Current
	}
	trip.Trip.Current = s.calcCurrentStatus(ctx, trip.Trip.Current, curLocation)

	if request.EndTrip {
		trip.Trip.Status = rentalpb.TripStatus_FINISHED
		trip.Trip.End = trip.Trip.Current
		if err := s.CarManager.Lock(ctx, id.CarID(trip.Trip.CarId)); err != nil {
			return nil, status.Errorf(codes.FailedPrecondition, "cannot lock car: %v", err)
		}
	}

	// trip.UpdatedAt   optimistic locking
	err = s.Mongo.UpdateTrip(ctx, tid, aid, trip.UpdatedAt, trip.Trip)
	return trip.Trip, nil
}

const centsPerSec = 0.7

var nowFunc = func() int64 {
	return time.Now().Unix()
}

type DistanceCalc interface {
	DistanceKm(context.Context, *rentalpb.Location, *rentalpb.Location) (float64, error)
}

func (s *Service) calcCurrentStatus(c context.Context, last *rentalpb.LocationStatus, cur *rentalpb.Location) *rentalpb.LocationStatus {
	now := nowFunc()
	i := float64(now - last.TimestampSec)

	km, err := s.DistanceCalc.DistanceKm(c, last.Location, cur)
	if err != nil {
		s.Logger.Warn("cannot calculate distance", zap.Error(err))
	}

	poi, err := s.POIManager.Resolve(c, cur)
	if err != nil {
		s.Logger.Info("cannot resolve poi", zap.Stringer("location", cur), zap.Error(err))
	}

	return &rentalpb.LocationStatus{
		Location:     cur,
		FeeCent:      last.FeeCent + int32(centsPerSec*i),
		KmDriven:     last.KmDriven + km,
		PoiName:      poi,
		TimestampSec: now,
	}
}
