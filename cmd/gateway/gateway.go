package main

import (
	"harnsgateway/cmd/gateway/app"
	"k8s.io/component-base/logs"
	_ "k8s.io/component-base/logs/json/register"
	"os"
)

func main() {
	cmd := app.NewGatewayCmd()
	logs.InitLogs()
	defer logs.FlushLogs()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
