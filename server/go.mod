module GraduationProjection/server

go 1.17

replace GraduationProjection/fsm => ../fsm

require (
	GraduationProjection/fsm v0.0.0-00010101000000-000000000000
	go.uber.org/zap v1.21.0
)

require (
	github.com/gogo/protobuf v1.3.2 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
)
