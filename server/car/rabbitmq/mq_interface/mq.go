package mq

import (
	"context"
	carpb "coolcar/car/api/gen/v1"
	coolenvpb "coolcar/shared/carenv"
)

type CarPublisher interface {
	Publish(ctx context.Context, entity *carpb.CarEntity) error
}

type CarSubscriber interface {
	Subscribe(ctx context.Context) (chan *carpb.CarEntity, func(), error)
}

type PosSubscriber interface {
	Subscribe(ctx context.Context) (chan *coolenvpb.CarPosUpdate, func(), error)
}
