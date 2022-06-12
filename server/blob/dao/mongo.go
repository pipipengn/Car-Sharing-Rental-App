package dao

import (
	"context"
	"coolcar/shared/id"
	mgo "coolcar/shared/mongo"
	"coolcar/shared/mongo/objid"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Mongo struct {
	col *mongo.Collection
}

func NewMongo(db *mongo.Database) *Mongo {
	return &Mongo{col: db.Collection("blob")}
}

type BlobRecord struct {
	mgo.IDField `bson:"inline"`
	AccountID   string `bson:"accountid"`
	Path        string `bson:"path"`
}

func (m *Mongo) CreateBlob(c context.Context, aid id.AccountID) (*BlobRecord, error) {
	br := &BlobRecord{
		AccountID: aid.String(),
	}
	objID := mgo.NewObjID()
	br.ID = objID
	br.Path = fmt.Sprintf("%s/%s", aid.String(), objID.Hex())

	_, err := m.col.InsertOne(c, br)
	if err != nil {
		return nil, err
	}
	return br, nil
}

func (m *Mongo) GetBlob(c context.Context, bid id.BlobID) (*BlobRecord, error) {
	objectID, err := objid.FromID(bid)
	if err != nil {
		return nil, fmt.Errorf("invalid object id: %v", err)
	}

	res := m.col.FindOne(c, bson.M{mgo.IDFieldName: objectID})
	if err := res.Err(); err != nil {
		return nil, err
	}

	var blobRecord BlobRecord
	err = res.Decode(&blobRecord)
	if err != nil {
		return nil, fmt.Errorf("cannot decode result: %v", err)
	}
	return &blobRecord, nil
}
