package main

import (
	"context"
	v1 "emag-homework/api/v1"
	"emag-homework/pkg/env"
	"emag-homework/pkg/log"
	"google.golang.org/grpc"
)

func main() {
	logger := log.NewLogger()

	appAddress, err := env.Require("APP_ADDRESS")
	if err != nil {
		panic(err)
	}

	cc, err := grpc.Dial(appAddress, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer cc.Close()

	logger.Info("successfully connected to %q", appAddress)

	client := v1.NewAppClient(cc)
	ctx := context.Background()

	saveRes, err := client.Save(ctx, &v1.SaveRequest{
		Text: "Lorem ipsum typed text lorem",
	})
	if err != nil {
		logger.Error("save failed: %s", err)

		return
	}

	logger.Info("successfully saved: %v", saveRes.Keywords)

	findRes, err := client.Find(ctx, &v1.FindRequest{
		Keywords: []string{"lorem"},
	})
	if err != nil {
		logger.Error("save failed: %s", err)

		return
	}

	logger.Info("successfully find: %v", findRes)
}
