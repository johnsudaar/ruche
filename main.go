package main

import (
	"context"

	"github.com/Scalingo/go-utils/logger"
	"github.com/johnsudaar/ruche/config"
	"github.com/johnsudaar/ruche/webserver"
	"github.com/pkg/errors"
)

func main() {
	err := config.Init()
	if err != nil {
		panic(errors.Wrap(err, "fail to init config"))
	}
	// Logger init
	log := logger.Default()
	ctx := logger.ToCtx(context.Background(), log)
	log.Info("Config initialized")

	webserver.Start(ctx)
}
