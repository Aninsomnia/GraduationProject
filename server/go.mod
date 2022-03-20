module GraduationProject/server

go 1.17

replace (
	GraduationProject/handle => ../handle
	GraduationProject/node => ../node
	GraduationProject/serverhttp => ../serverhttp
)

require (
	GraduationProject/handle v0.0.0-00010101000000-000000000000
	GraduationProject/node v0.0.0-00010101000000-000000000000
	GraduationProject/serverhttp v0.0.0-00010101000000-000000000000
)

require (
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.21.0 // indirect
)
