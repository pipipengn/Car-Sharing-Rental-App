package simulation

import (
	"context"
	carpb "coolcar/car/api/gen/v1"
	mq "coolcar/car/rabbitmq/mq_interface"
	coolenvpb "coolcar/shared/carenv"
	"fmt"
	"go.uber.org/zap"
	"time"
)

type Controller struct {
	Logger        *zap.Logger
	CarService    carpb.CarServiceClient
	CarSubscriber mq.CarSubscriber
	PosSubscriber mq.PosSubscriber
	AIService     coolenvpb.AIServiceClient
}

func (c *Controller) RunSimulations(ctx context.Context) {
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

	carCh, closeFunc, err := c.CarSubscriber.Subscribe(ctx)
	defer closeFunc()
	if err != nil {
		c.Logger.Error("cannot subscribe car", zap.Error(err))
		return
	}
	posCh, closeFunc2, err := c.PosSubscriber.Subscribe(ctx)
	defer closeFunc2()
	if err != nil {
		c.Logger.Error("cannot subscribe position", zap.Error(err))
		return
	}

	carChans := make(map[string]chan *carpb.Car)
	posChans := make(map[string]chan *carpb.Location)
	for _, car := range cars {
		ch := make(chan *carpb.Car)
		carChans[car.Id] = ch
		ch2 := make(chan *carpb.Location)
		posChans[car.Id] = ch2
		go c.SimulateCar(context.Background(), car, ch, ch2)
	}

	for {
		select {
		case carUpdate := <-carCh:
			ch := carChans[carUpdate.Id]
			if ch != nil {
				ch <- carUpdate.Car
			}
		case posUpdate := <-posCh:
			ch := posChans[posUpdate.CarId]
			if ch != nil {
				ch <- &carpb.Location{
					Latitude:  posUpdate.Pos.Latitude,
					Longitude: posUpdate.Pos.Longitude,
				}
			}
		}
	}
}

func (c *Controller) SimulateCar(ctx context.Context, initial *carpb.CarEntity, carCh chan *carpb.Car, posCh chan *carpb.Location) {
	car := initial
	c.Logger.Info("Simulating car.", zap.String("id", car.Id))

	for {
		select {
		case carUpdate := <-carCh:
			if carUpdate.Status == carpb.CarStatus_UNLOCKING {
				updatedCarEntity, err := c.unlockCar(ctx, car)
				if err != nil {
					c.Logger.Error("cannot unlock car", zap.String("id", car.Id), zap.Error(err))
					break
				}
				car = updatedCarEntity
			} else if carUpdate.Status == carpb.CarStatus_LOCKING {
				updatedCarEntity, err := c.lockcar(ctx, car)
				if err != nil {
					c.Logger.Error("cannot lock car", zap.String("id", car.Id), zap.Error(err))
					break
				}
				car = updatedCarEntity
			}
		case posUpdate := <-posCh:
			updatedCarEntity, err := c.moveCar(ctx, car, posUpdate)
			if err != nil {
				c.Logger.Error("cannot move car", zap.String("id", car.Id), zap.Error(err))
				break
			}
			car = updatedCarEntity
		}
	}
}

func (c *Controller) lockcar(ctx context.Context, car *carpb.CarEntity) (*carpb.CarEntity, error) {
	car.Car.Status = carpb.CarStatus_LOCKED
	// lock car need hardware
	if _, err := c.CarService.UpdateCar(ctx, &carpb.UpdateCarRequest{
		Id:     car.Id,
		Status: carpb.CarStatus_LOCKED,
	}); err != nil {
		return nil, fmt.Errorf("cannot lock car: %v", err)
	}

	// stop position simulation
	if _, err := c.AIService.EndSimulateCarPos(ctx, &coolenvpb.EndSimulateCarPosRequest{CarId: car.Id}); err != nil {
		return nil, fmt.Errorf("cannot end simulate: %v", err)
	}
	return car, nil
}

func (c *Controller) unlockCar(ctx context.Context, car *carpb.CarEntity) (*carpb.CarEntity, error) {
	car.Car.Status = carpb.CarStatus_UNLOCKED
	if _, err := c.CarService.UpdateCar(ctx, &carpb.UpdateCarRequest{
		Id:     car.Id,
		Status: carpb.CarStatus_UNLOCKED,
	}); err != nil {
		return nil, fmt.Errorf("cannot update car state: %v", err)
	}

	if _, err := c.AIService.SimulateCarPos(ctx, &coolenvpb.SimulateCarPosRequest{
		CarId: car.Id,
		Type:  coolenvpb.PosType_NINGBO,
		InitialPos: &coolenvpb.Location{
			Latitude:  car.Car.Position.Latitude,
			Longitude: car.Car.Position.Longitude,
		},
	}); err != nil {
		return nil, fmt.Errorf("cannot simulate car position: %v", err)
	}
	return car, nil
}

func (c *Controller) moveCar(ctx context.Context, car *carpb.CarEntity, pos *carpb.Location) (*carpb.CarEntity, error) {
	car.Car.Position = pos
	if _, err := c.CarService.UpdateCar(ctx, &carpb.UpdateCarRequest{
		Id:       car.Id,
		Position: pos,
	}); err != nil {
		return nil, fmt.Errorf("cannot update car: %v", err)
	}
	return car, nil
}
