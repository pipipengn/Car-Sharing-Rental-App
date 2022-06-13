package simulation

import (
	"context"
	carpb "coolcar/car/api/gen/v1"
	"go.uber.org/zap"
	"time"
)

type Contorller struct {
	Logger     *zap.Logger
	CarService carpb.CarServiceClient
	Subscriber Subscriber
}

type Subscriber interface {
	Subscribe(ctx context.Context) (chan *carpb.CarEntity, func(), error)
}

func (c *Contorller) RunSimulations(ctx context.Context) {
	var cars []*carpb.CarEntity
	for {
		time.Sleep(time.Second)
		res, err := c.CarService.GetCars(ctx, &carpb.GetCarsRequest{})
		if err != nil {
			c.Logger.Error("cannot get cars", zap.Error(err))
			continue
		}
		cars = res.Cars
		break
	}
	c.Logger.Info("Running car simulations.", zap.Int("car count", len(cars)))

	msgCh, closeFunc, err := c.Subscriber.Subscribe(ctx)
	defer closeFunc()
	if err != nil {
		c.Logger.Error("cannot subscribe", zap.Error(err))
		return
	}

	carChans := make(map[string]chan *carpb.Car)
	for _, car := range cars {
		ch := make(chan *carpb.Car)
		carChans[car.Id] = ch
		go c.SimulateCar(context.Background(), car, ch)
	}

	for carUpdate := range msgCh {
		ch := carChans[carUpdate.Id]
		if ch != nil {
			ch <- carUpdate.Car
		}
	}
}

func (c *Contorller) SimulateCar(ctx context.Context, initial *carpb.CarEntity, ch chan *carpb.Car) {
	carID := initial.Id
	c.Logger.Info("Simulating car.", zap.String("id", carID))

	for carUpdate := range ch {
		if carUpdate.Status == carpb.CarStatus_UNLOCKING {
			// Unlock car need hardware
			if _, err := c.CarService.UpdateCar(ctx, &carpb.UpdateCarRequest{
				Id:     carID,
				Status: carpb.CarStatus_UNLOCKED,
			}); err != nil {
				c.Logger.Error("cannot unlock car", zap.Error(err))
			} else {
				c.Logger.Info("unlock car successful", zap.String("car: ", carID))
			}
		} else if carUpdate.Status == carpb.CarStatus_LOCKING {
			// lock car need hardware
			if _, err := c.CarService.UpdateCar(ctx, &carpb.UpdateCarRequest{
				Id:     carID,
				Status: carpb.CarStatus_LOCKED,
			}); err != nil {
				c.Logger.Error("cannot lock car", zap.Error(err))
			}
		}
	}
}
