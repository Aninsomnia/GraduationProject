package servermain

func Main() {
	initCfg := newConfig()
	startServer(initCfg)
}
