package serverhttp

import (
	"go.uber.org/zap"
)

type Transporter interface {
	// Start 启动一个Transport.
	Start() error
	//请求发起仲裁
	Requestarbitration() error
}
type transporte struct {
	Logger    *zap.Logger
	LocalID   ID
	LocalURLs URLs

	WitnessURLs URLs // 仲裁介质所处地址
}
