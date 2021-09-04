package server

import (
	"context"
	v1 "emag-homework/internal/db/api/v1"
	"emag-homework/internal/db/node"
	"fmt"
	"math/rand"
	"time"
)

func GenerateID() string {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	return fmt.Sprint(rnd.Int63())
}

type NodeInfo struct {
	ID      string
	Address string
}

func Register(ctx context.Context, info NodeInfo, ctrlClient v1.ControllerClient, logger node.Logger) error {
	logger.Info("register to the controller...")

	_, err := ctrlClient.RegisterNode(ctx, &v1.RegisterNodeRequest{
		Id:      info.ID,
		Address: info.Address,
	})
	if err != nil {
		return fmt.Errorf("failed to register: %w", err)
	}

	logger.Info("successfully registered to controller")

	return nil
}

func Unregister(
	ctx context.Context, info NodeInfo, ctrlClient v1.ControllerClient, logger node.Logger,
) error {
	logger.Info("unregister from the controller...")

	_, err := ctrlClient.UnregisterNode(ctx, &v1.UnregisterNodeRequest{
		Id: info.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to unregister: %w", err)
	}

	logger.Info("successfully unregistered to controller")

	return nil
}
