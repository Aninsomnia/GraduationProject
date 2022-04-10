package servermain

import (
	"fmt"
	"os"

	"go.uber.org/zap"
)

type initialConfig struct {
	logger *zap.Logger
}

func newConfig() *initialConfig {
	cfg := &initialConfig{}

	//初始化logger
	lg, zapError := zap.NewProduction()
	if zapError != nil {
		fmt.Printf("error creating zap logger %v", zapError)
		os.Exit(1)
	}
	cfg.logger = lg

	return cfg
}
