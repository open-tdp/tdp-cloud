package worker

import (
	"time"

	"tdp-cloud/cmd/args"
	"tdp-cloud/module/worker"
)

func inlet() {

	defer timer()

	args.WriteConfig()

	if err := worker.Connect(); err != nil {
		svclog.Error(err)
	}

}

func timer() {

	svclog.Warning("Connection disconnected, retry in 15 seconds.")

	time.Sleep(15 * time.Second)
	inlet()

}