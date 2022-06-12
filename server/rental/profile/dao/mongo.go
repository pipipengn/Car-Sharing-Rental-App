package dao

import (
	"context"
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/shared/id"
	mgo "coolcar/shared/mongo"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	accountIDField      = "accountid"
	profileField        = "profile"
	identityStatusField = profileField + ".identitystatus"
	photoBLobIDField    = "photoblobid"
)

type Mongo struct {
	col *mongo.Collection
}

func NewMongo(db *mongo.Database) *Mongo {
	return &Mongo{col: db.Collection("profile")}
}

type ProfileRecord struct {
	AccountID   string            `bson:"accountid"`
	Profile     *rentalpb.Profile `bson:"profile"`
	PhotoBLobID string            `bson:"photoblobid"`
}

func (m *Mongo) GetProfile(c context.Context, aid id.AccountID) (*ProfileRecord, error) {
	res := m.col.FindOne(c, byAccountID(aid))
	if err := res.Err(); err != nil {
		return nil, err
	}
	var profileRecord ProfileRecord
	err := res.Decode(&profileRecord)
	if err != nil {
		return nil, fmt.Errorf("cannot decode profile record: %v", err)
	}
	return &profileRecord, nil
}

func (m Mongo) UpdateProfile(c context.Context, aid id.AccountID, p *rentalpb.Profile, prevStatus rentalpb.IdentityStatus) error {
	filter := bson.M{
		identityStatusField: prevStatus,
	}
	if prevStatus == rentalpb.IdentityStatus_UNSUBMITTED {
		filter = mgo.ZeroOrDoesNotExist(identityStatusField, rentalpb.IdentityStatus_UNSUBMITTED)
	}
	filter[accountIDField] = aid.String()

	_, err := m.col.UpdateOne(c, filter, mgo.Set(bson.M{
		accountIDField: aid.String(),
		profileField:   p,
	}), options.Update().SetUpsert(true))
	return err
}

func (m *Mongo) UpdateProfilePhoto(c context.Context, aid id.AccountID, bid id.BlobID) error {
	_, err := m.col.UpdateOne(c, bson.M{
		accountIDField: aid.String(),
	}, mgo.Set(bson.M{
		accountIDField:   aid.String(),
		photoBLobIDField: bid.String(),
	}), options.Update().SetUpsert(true))
	return err
}

func byAccountID(aid id.AccountID) bson.M {
	return bson.M{accountIDField: aid.String()}
}
