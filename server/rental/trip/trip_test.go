package trip

import (
	"context"
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/rental/trip/client/poi"
	"coolcar/rental/trip/dao"
	"coolcar/shared/auth"
	"coolcar/shared/id"
	mgo "coolcar/shared/mongo"
	mongotesting "coolcar/shared/mongo/testing"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"math/rand"
	"os"
	"testing"
)

func TestCreateTrip(t *testing.T) {
	c := context.Background()
	car := &carManager{}
	profile := &profileManager{iID: "identity1"}
	s := NewService(c, t, car, profile)

	nowFunc = func() int64 {
		return 1654839573
	}

	req := &rentalpb.CreateTripRequest{
		CarId: "car1",
		Start: &rentalpb.Location{
			Latitude:  32.1231,
			Longitude: 114.623,
		},
	}

	golden := `{"account_id":%q,"car_id":"car1","start":{"location":{"latitude":32.1231,"longitude":114.623},"poi_name":"San Diago","timestamp_sec":1654839573},"current":{"location":{"latitude":32.1231,"longitude":114.623},"poi_name":"San Diago","timestamp_sec":1654839573},"status":1,"identity_id":"identity1"}`
	cases := []struct {
		name         string
		accountID    string
		tripID       string
		profileErr   error
		carVerifyErr error
		carUnlockErr error
		want         string
		wantErr      bool
	}{
		{
			name:      "normal_create",
			accountID: "account1",
			tripID:    "5f8132eb12714bf629489054",
			want:      fmt.Sprintf(golden, "account1"),
		},
		{
			name:       "profile_err",
			accountID:  "account2",
			tripID:     "5f8132eb12714bf629489055",
			profileErr: fmt.Errorf("profile"),
			wantErr:    true,
		},
		{
			name:         "car_verify_err",
			accountID:    "account3",
			tripID:       "5f8132eb12714bf629489056",
			carVerifyErr: fmt.Errorf("verify"),
			wantErr:      true,
		},
		{
			name:         "car_unlock_err",
			accountID:    "account4",
			tripID:       "5f8132eb12714bf629489057",
			carUnlockErr: fmt.Errorf("unlock"),
			want:         fmt.Sprintf(golden, "account4"),
		},
	}

	for _, cc := range cases {
		t.Run(cc.name, func(t *testing.T) {
			mgo.NewObjIDWithValue(id.TripID(cc.tripID))
			profile.err = cc.profileErr
			car.verifyErr = cc.carVerifyErr
			car.unlockErr = cc.carUnlockErr
			c := auth.ContextWithAccountID(context.Background(), id.AccountID(cc.accountID))
			tripEntity, err := s.CreateTrip(c, req)
			if cc.wantErr {
				if err == nil {
					t.Errorf("want error; got none")
				} else {
					return
				}
			}
			if err != nil {
				t.Errorf("got an error: %v", err)
				return
			}
			if tripEntity.Id != cc.tripID {
				t.Errorf("incorrect id; want %q, got %q", cc.tripID, tripEntity.Id)
			}
			b, err := json.Marshal(tripEntity.Trip)
			if err != nil {
				t.Errorf("cannot marshall response: %v", err)
			}
			got := string(b)
			if cc.want != got {
				t.Errorf("incorrect response: want %s, got %s", cc.want, got)
			}
		})
	}
}

func TestTripLifeCycle(t *testing.T) {
	c := auth.ContextWithAccountID(context.Background(), "account_for_lifecycle")
	car := &carManager{}
	profile := &profileManager{}
	s := NewService(c, t, car, profile)

	tid := id.TripID("5f8132eb22714bf629489056")
	mgo.NewObjIDWithValue(tid)
	cases := []struct {
		name    string
		now     int64
		op      func() (*rentalpb.Trip, error)
		want    string
		wantErr bool
	}{
		{
			name: "create_trip",
			now:  10000,
			op: func() (*rentalpb.Trip, error) {
				e, err := s.CreateTrip(c, &rentalpb.CreateTripRequest{
					CarId: "car1",
					Start: &rentalpb.Location{
						Latitude:  32.123,
						Longitude: 114.2525,
					},
				})
				if err != nil {
					return nil, err
				}
				return e.Trip, nil
			},
			want: `{"account_id":"account_for_lifecycle","car_id":"car1","start":{"location":{"latitude":32.123,"longitude":114.2525},"poi_name":"San Diago","timestamp_sec":10000},"current":{"location":{"latitude":32.123,"longitude":114.2525},"poi_name":"San Diago","timestamp_sec":10000},"status":1}`,
		},
		{
			name: "update_trip",
			now:  20000,
			op: func() (*rentalpb.Trip, error) {
				return s.UpdateTrip(c, &rentalpb.UpdateTripRequest{
					Id: tid.String(),
					Current: &rentalpb.Location{
						Latitude:  28.234234,
						Longitude: 123.243255,
					},
				})
			},
			want: `{"account_id":"account_for_lifecycle","car_id":"car1","start":{"location":{"latitude":32.123,"longitude":114.2525},"poi_name":"San Diago","timestamp_sec":10000},"current":{"location":{"latitude":28.234234,"longitude":123.243255},"fee_cent":7000,"km_driven":100,"poi_name":"Chicago","timestamp_sec":20000},"status":1}`,
		},
		{
			name: "finish_trip",
			now:  30000,
			op: func() (*rentalpb.Trip, error) {
				return s.UpdateTrip(c, &rentalpb.UpdateTripRequest{
					Id:      tid.String(),
					EndTrip: true,
				})
			},
			want: `{"account_id":"account_for_lifecycle","car_id":"car1","start":{"location":{"latitude":32.123,"longitude":114.2525},"poi_name":"San Diago","timestamp_sec":10000},"current":{"location":{"latitude":28.234234,"longitude":123.243255},"fee_cent":14000,"km_driven":100,"poi_name":"Chicago","timestamp_sec":30000},"end":{"location":{"latitude":28.234234,"longitude":123.243255},"fee_cent":14000,"km_driven":100,"poi_name":"Chicago","timestamp_sec":30000},"status":2}`,
		},
		{
			name: "query_trip",
			now:  40000,
			op: func() (*rentalpb.Trip, error) {
				return s.GetTrip(c, &rentalpb.GetTripRequest{
					Id: tid.String(),
				})
			},
			want: `{"account_id":"account_for_lifecycle","car_id":"car1","start":{"location":{"latitude":32.123,"longitude":114.2525},"poi_name":"San Diago","timestamp_sec":10000},"current":{"location":{"latitude":28.234234,"longitude":123.243255},"fee_cent":14000,"km_driven":100,"poi_name":"Chicago","timestamp_sec":30000},"end":{"location":{"latitude":28.234234,"longitude":123.243255},"fee_cent":14000,"km_driven":100,"poi_name":"Chicago","timestamp_sec":30000},"status":2}`,
		},
		{
			name: "update_after_finished",
			now:  50000,
			op: func() (*rentalpb.Trip, error) {
				return s.UpdateTrip(c, &rentalpb.UpdateTripRequest{
					Id: tid.String(),
				})
			},
			wantErr: true,
		},
	}

	rand.Seed(1345)
	for _, cc := range cases {
		nowFunc = func() int64 {
			return cc.now
		}
		trip, err := cc.op()
		if cc.wantErr {
			if err == nil {
				t.Errorf("%s: want error; got none", cc.name)
			} else {
				continue
			}
		}
		if err != nil {
			t.Errorf("%s: operation failed: %v", cc.name, err)
			continue
		}
		b, err := json.Marshal(trip)
		if err != nil {
			t.Errorf("%s: failed marshalling response: %v", cc.name, err)
		}
		got := string(b)
		if cc.want != got {
			t.Errorf("%s: incorrect response; want: %s, got: %s", cc.name, cc.want, got)
		}
	}
}

func NewService(c context.Context, t *testing.T, car *carManager, profile *profileManager) *Service {
	logger, _ := zap.NewDevelopment()

	mc, err := mongotesting.NewClient(c)
	if err != nil {
		t.Fatalf("cannot connect to mongo client: %v", err)
	}

	db := mc.Database("coolcar")
	_ = mongotesting.SetupIndexes(c, db)
	return &Service{
		Logger:         logger,
		Mongo:          dao.NewMongo(db),
		POIManager:     poi.Manager{},
		CarManager:     car,
		ProfileManager: profile,
		DistanceCalc:   &distCalc{},
	}
}

type distCalc struct {
}

func (d *distCalc) DistanceKm(c context.Context, from *rentalpb.Location, to *rentalpb.Location) (float64, error) {
	if from.Longitude == to.Longitude && from.Latitude == to.Latitude {
		return 0, nil
	}
	return 100, nil
}

type profileManager struct {
	iID id.IdentityID
	err error
}

func (p *profileManager) Verify(context.Context, id.AccountID) (id.IdentityID, error) {
	return p.iID, p.err
}

type carManager struct {
	verifyErr error
	unlockErr error
}

func (m *carManager) Verify(context.Context, id.CarID, *rentalpb.Location) error {
	return m.verifyErr
}

func (m *carManager) Unlock(context.Context, id.CarID, id.AccountID, id.TripID) error {
	return m.unlockErr
}

func (m *carManager) Lock(context.Context, id.CarID) error {
	return nil
}

func TestMain(m *testing.M) {
	os.Exit(mongotesting.RunWithMongoInDocker(m))
}
