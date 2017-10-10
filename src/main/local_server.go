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
	execCmd("monitor","off")
	execCmd("mutesysvolume","1")
}

func turnOn(response http.ResponseWriter, request *http.Request) {
	execCmd("monitor","on")
	execCmd("mutesysvolume","0")
}

func volumeUp(response http.ResponseWriter, request *http.Request) {
	execCmd("changesysvolume","6500")
}

func volumeDown(response http.ResponseWriter, request *http.Request) {
	execCmd("changesysvolume","-6500")
}

func execCmd(command,value string){
	logger.GetLogger2().Info("Exec",command,"with",value)
	exec.Command(nirCmdPath,command,value).Run()
}