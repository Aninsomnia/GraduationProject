package controller

import (
	"GraduationProjection/fsm"
	"time"
)

type bootstrappedController struct {
	fsm *bootstrappedFsm
}
type bootstrappedFsm struct {
	heartbeat time.Duration // fsmNode的相关配置

	config *fsm.Config // fsmNode中fsmCore的相关配置
}

func bootstrap() (b *bootstrappedController, err error) {

	return &bootstrappedController{
		fsm: bootstrapFsm(),
	}, nil
}

func (b *bootstrappedFsm) newFsmNode() *fsmNode {
	// 创建FSM
	f := fsm.StartFsm(b.config)

	// 创建fsmNodeConfig
	fsmNodeCfg := fsmNodeConfig{
		FSM:       f,
		heartbeat: b.heartbeat,
	}
	// 将使用fsmNodeConfig创建fsmNode
	return newFsmNode(fsmNodeCfg)
}

func bootstrapFsm() *bootstrappedFsm {
	return &bootstrappedFsm{
		heartbeat: time.Duration(1000) * time.Millisecond,
		config:    fsmConfig(),
	}
}
func fsmConfig() *fsm.Config {
	return &fsm.Config{
		HeartbeatSendTick:    1,
		HeartbeatReceiveTick: 10,
		ReqArbitSendTick:     5,
	}
}
