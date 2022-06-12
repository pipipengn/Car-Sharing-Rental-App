package poi

import (
	"context"
	rentalpb "coolcar/rental/api/gen/v1"
	"github.com/golang/protobuf/proto"
	"hash/fnv"
)

var poi = []string{
	"New York",
	"Los Angeles",
	"Chicago",
	"San Diago",
}

type Manager struct {
}

func (m Manager) Resolve(c context.Context, req *rentalpb.Location) (string, error) {
	b, err := proto.Marshal(req)
	if err != nil {
		return "", err
	}

	h := fnv.New32()
	_, _ = h.Write(b)
	return poi[int(h.Sum32())%len(poi)], nil
}
