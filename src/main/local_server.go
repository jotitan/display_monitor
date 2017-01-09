package main

import (
	"arguments"
	"logger"
	"net/http"
	"strings"
	"os/exec"
)


// create a micro service rest to manage sleep and wake up

var nirCmdPath string

func main() {

	logger.InitLogger("local.log",true)

	args := arguments.NewArguments()
	port := args.GetString("port")
	nirCmdPath = args.GetMandatoryString("nircmd","Define nircmd path with -nircmd")
	if strings.EqualFold(port,""){
		port = "9098"
	}
	smux := http.NewServeMux()
	smux.HandleFunc("/turnOn",turnOn)
	smux.HandleFunc("/turnOff",turnOff)
	smux.HandleFunc("/volumeUp",volumeUp)
	smux.HandleFunc("/volumeDown",volumeDown)

	logger.GetLogger().Info("Start micro service to turn off / on  screen on",port)
	err := http.ListenAndServe(":" + port,smux)
	logger.GetLogger().Error(err)
}

func turnOff(response http.ResponseWriter, request *http.Request) {
	turnOff := exec.Command(nirCmdPath,"monitor","off")
	turnOff.Run()

	volumeOff := exec.Command(nirCmdPath,"mutesysvolume","1")
	volumeOff.Run()
}

func turnOn(response http.ResponseWriter, request *http.Request) {
	turnOff := exec.Command(nirCmdPath,"monitor","on")
	turnOff.Run()

	volumeOff := exec.Command(nirCmdPath,"mutesysvolume","0")
	volumeOff.Run()
}

func volumeUp(response http.ResponseWriter, request *http.Request) {
	exec.Command(nirCmdPath,"changesysvolume","6500").Run()
}

func volumeDown(response http.ResponseWriter, request *http.Request) {
	exec.Command(nirCmdPath,"changesysvolume","-6500").Run()
}