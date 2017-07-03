package main

import (
	"xplanet"
	"server"
	"time"
	"arguments"
	"path/filepath"
	"logger"
)


func main() {

	logger.InitLogger("display.log",true)

	args := arguments.NewArguments()
	app := args.GetMandatoryString("app", "need to specify the xplanet app folder with option -app")
	generateFolder := args.GetMandatoryString("generated","need to specify output folder with option -generated")
	configFolder := args.GetMandatoryString("config","need to specify config folder with option -config")
	asyncGeneration := ! (args.GetString("sync") == "true")
	galleryParam := args.GetString("gallery")
	webFolder := args.GetMandatoryString("resources","need to specify web resources folder with option -resources")

	xp := xplanet.New(app,filepath.Join(configFolder,"earth.conf"),generateFolder,"earth","48","3",600,600,asyncGeneration)
	xpMoon := xplanet.New(app,filepath.Join(configFolder,"moon.conf"),generateFolder,"moon","0","0",400,400,asyncGeneration)

	// Used to be sure to load cloud image before launching generation
	ackFirstGeneration := make(chan struct{}	,1)
	go xp.LoadCloud(time.Duration(12) * time.Hour,ackFirstGeneration)

	// Wait first cloud get
	<- ackFirstGeneration
	go xp.RunAutoGenerations(0)
	go xpMoon.GenerateEvery(time.Duration(6) * time.Hour)

	s := server.New(webFolder,xp,xpMoon,galleryParam)
	s.Launch(args.GetString("port"))
}
