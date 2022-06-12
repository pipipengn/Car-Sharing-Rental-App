package dao

import (
	"context"
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/shared/id"
	mgo "coolcar/shared/mongo"
	"coolcar/shared/mongo/objid"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	tripField      = "trip"
	accountIDField = tripField + ".accountid"
	statusField    = tripField + ".status"
)

type Mongo struct {
	col *mongo.Collection
}

func NewMongo(db *mongo.Database) *Mongo {
	return &Mongo{col: db.Collection(tripField)}
}

type TripRecord struct {
	mgo.IDField        `bson:"inline"`
	mgo.UpdatedAtField `bson:"inline"`
	Trip               *rentalpb.Trip `bson:"trip"`
}

func (m *Mongo) CreateTrip(ctx context.Context, trip *rentalpb.Trip) (*TripRecord, error) {
	r := &TripRecord{Trip: trip}
	r.ID = mgo.NewObjID()
	r.UpdatedAt = mgo.UpdatedAt()

	_, err := m.col.InsertOne(ctx, r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (m *Mongo) GetTrip(ctx context.Context, id id.TripID, accountID id.AccountID) (*TripRecord, error) {
	objID, err := objid.FromID(id)
	if err != nil {
		return nil, fmt.Errorf("invalid id: %v", err)
	}

	result := m.col.FindOne(ctx, bson.M{
		mgo.IDFieldName: objID,
		accountIDField:  accountID,
	})

	if err := result.Err(); err != nil {
		return nil, err
	}

	var tr TripRecord
	err = result.Decode(&tr)
	if err != nil {
		return nil, fmt.Errorf("cannot decode: %v", err)
	}
	return &tr, nil
}

func (m *Mongo) GetTrips(ctx context.Context, accountID id.AccountID, status rentalpb.TripStatus) ([]*TripRecord, error) {
	filter := bson.M{
		accountIDField: accountID.String(),
	}
	if status != rentalpb.TripStatus_TS_NOT_SPECIFIED {
		filter[statusField] = status
	}
	res, err := m.col.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var trips []*TripRecord
	for res.Next(ctx) {
		var trip TripRecord
		err := res.Decode(&trip)
		if err != nil {
			return nil, err
		}
		trips = append(trips, &trip)
	}
	return trips, nil
}

func (m *Mongo) UpdateTrip(c context.Context, tid id.TripID, aid id.AccountID, updatedAt int64, trip *rentalpb.Trip) error {
	objID, err := objid.FromID(tid)
	if err != nil {
		return fmt.Errorf("invalid id: %v", err)
	}

	res, err := m.col.UpdateOne(c, bson.M{
		mgo.IDFieldName:        objID,
		accountIDField:         aid.String(),
		mgo.UpdatedAtFieldName: updatedAt, // optimistic locking
	}, mgo.Set(bson.M{
		tripField:              trip,
		mgo.UpdatedAtFieldName: mgo.UpdatedAt(),
	}))
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}
