package position

import (
	"context"
	"coolcar/car/rabbitmq"
	coolenvpb "coolcar/shared/carenv"
	"encoding/json"
	"go.uber.org/zap"
)

type Subscriber struct {
	Sub    *rabbitmq.Subscriber
	Logger *zap.Logger
}

func (s *Subscriber) Subscribe(ctx context.Context) (chan *coolenvpb.CarPosUpdate, func(), error) {
	msgCh, closeFunc, err := s.Sub.SubscribeRaw(ctx)
	if err != nil {
		return nil, closeFunc, err
	}

	posCh := make(chan *coolenvpb.CarPosUpdate)
	go func() {
		for msg := range msgCh {
			var posUpdate coolenvpb.CarPosUpdate
			err = json.Unmarshal(msg.Body, &posUpdate)
			if err != nil {
				s.Logger.Error("cannot unmmarshal", zap.Error(err))
			}
			posCh <- &posUpdate
		}
		close(posCh)
	}()
	return posCh, closeFunc, nil
}
