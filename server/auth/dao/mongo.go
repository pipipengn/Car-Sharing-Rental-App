package dao

import (
	"context"
	"coolcar/shared/id"
	mgo "coolcar/shared/mongo"
	"coolcar/shared/mongo/objid"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const openIDField = "open_id"

type Mongo struct {
	col *mongo.Collection
}

func NewMongo(db *mongo.Database) *Mongo {
	return &Mongo{col: db.Collection("account")}
}

func (m *Mongo) ResolveAccountID(ctx context.Context, openID string) (id.AccountID, error) {
	res := m.col.FindOneAndUpdate(ctx, bson.M{openIDField: openID}, mgo.SetOnInsert(bson.M{
		mgo.IDFieldName: mgo.NewObjID(),
		openIDField:     openID,
	}), options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After))

	if err := res.Err(); err != nil {
		return "", fmt.Errorf("cannot findOneAndUpdate: %v", err)
	}

	var row mgo.IDField
	if err := res.Decode(&row); err != nil {
		return "", fmt.Errorf("cannot decode result: %v", err)
	}

	return objid.ToAccountID(row.ID), nil
}
