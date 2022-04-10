package controller

import (
	"GraduationProjection/fsm"
	"time"

	"go.uber.org/zap"
)

type bootstrappedController struct {
	fsm *bootstrappedFsm
}
type bootstrappedFsm struct {
	logger *zap.Logger

	heartbeat time.Duration // fsmNode的相关配置

	config *fsm.Config // fsmNode中fsmCore的相关配置
}

func bootstrap(cfg *ControllerConfig) (b *bootstrappedController, err error) {

	return &bootstrappedController{
		fsm: bootstrapFsm(cfg),
	}, nil
}

func (b *bootstrappedFsm) newFsmNode() *fsmNode {
	// 创建FSM
	f := fsm.StartFsm(b.config)

	// 创建fsmNodeConfig
	fsmNodeCfg := fsmNodeConfig{
		logger:    b.logger,
		FSM:       f,
		heartbeat: b.heartbeat,
	}
	// 将使用fsmNodeConfig创建fsmNode
	return newFsmNode(fsmNodeCfg)
}

func bootstrapFsm(cfg *ControllerConfig) *bootstrappedFsm {
	return &bootstrappedFsm{
		logger:    cfg.Logger,
		heartbeat: time.Duration(1000) * time.Millisecond,
		config:    fsmConfig(cfg),
	}
}
func fsmConfig(cfg *ControllerConfig) *fsm.Config {
	return &fsm.Config{
		HeartbeatSendTick:    1,
		HeartbeatReceiveTick: 10,
		ReqArbitSendTick:     5,
	}
}
