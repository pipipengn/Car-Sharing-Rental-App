package dao

import (
	"context"
	"coolcar/shared/id"
	mgo "coolcar/shared/mongo"
	"coolcar/shared/mongo/objid"
	mongotesting "coolcar/shared/mongo/testing"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"os"
	"testing"
)

func TestResolveAccountID(t *testing.T) {
	ctx := context.Background()
	mc, err := mongotesting.NewClient(ctx)
	if err != nil {
		t.Fatalf("cannot connect mongodb: %v", err)
	}

	m := NewMongo(mc.Database("coolcar"))
	if _, err = m.col.InsertMany(ctx, []interface{}{
		bson.M{
			mgo.IDFieldName: objid.MustFromID(id.AccountID("629ff133aaaabfc5e7d26f01")),
			openIDField:     "openid_1",
		},
		bson.M{
			mgo.IDFieldName: objid.MustFromID(id.AccountID("629ff133aaaabfc5e7d26f02")),
			openIDField:     "openid_2",
		},
	}); err != nil {
		t.Fatalf("cannot insert initial values")
	}
	mgo.NewObjID = func() primitive.ObjectID {
		return objid.MustFromID(id.AccountID("629ff133aaaabfc5e7d26f03"))
	}

	cases := []struct {
		name   string
		openID string
		want   string
	}{
		{
			name:   "existing_user",
			openID: "openid_1",
			want:   "629ff133aaaabfc5e7d26f01",
		},
		{
			name:   "another_existing_user",
			openID: "openid_2",
			want:   "629ff133aaaabfc5e7d26f02",
		},
		{
			name:   "new_user",
			openID: "openid_3",
			want:   "629ff133aaaabfc5e7d26f03",
		},
	}

	for _, cc := range cases {
		t.Run(cc.name, func(t *testing.T) {
			accountID, err := m.ResolveAccountID(context.Background(), cc.openID)
			if err != nil {
				t.Errorf("faild resolve account id for %q: %v", cc.openID, err)
			}
			if accountID.String() != cc.want {
				t.Errorf("resolve account id: want: %q; got: %q", cc.want, accountID)
			}
		})
	}
}

func TestMain(m *testing.M) {
	os.Exit(mongotesting.RunWithMongoInDocker(m))
}
