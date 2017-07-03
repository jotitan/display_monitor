package xplanet

import (
	"os/exec"
	"fmt"
	"math"
	"image/gif"
	"image/jpeg"
	"os"
	"image"
	"bytes"
	"sync"
	"time"
	"path/filepath"
	"logger"
	"net/http"
	"io"
	"syscall"
)

// TODO : get clouds image to add on earth

type XPlanet struct {
	AppPath string
	ConfigPath string
	DefaultParameter []string
	Coordinates []string
	// prefix of all generated images
	prefix string
	folder string
	// If true, generate gif and xplanet image in parallel
	GenerateAsync bool
}

const (
	MINUTES_TIME_GENERATION = 40
)

// appPath : path of xplanet application
// config : path of config to generate images
// folder : where to put generated images
// target : name of the target planet
func New(appPath,config,folder,target string, latitude,longitude string,width,height int, generateAsync bool)XPlanet{
	xp := XPlanet{AppPath:appPath,ConfigPath:config,prefix:target,folder:folder,GenerateAsync:generateAsync}
	xp.DefaultParameter = []string{"-radius","40","-range","1000","-num_times","1","-target",target,"-quality","95","-geometry",fmt.Sprintf("%dx%d",width,height)}
	xp.Coordinates = []string{"-latitude",latitude,"-longitude",longitude}
	return xp
}


// Run the automatic creation of photo
// Tips : formating date in GO : 20060102.15:04:05
// shift is used to generate photos some time before. Photos are generated at HH40
func (xp XPlanet)RunAutoGenerations(shift int){
	current := time.Now().Add(time.Duration(shift)*time.Minute)
	fDate := current.Format("20060102.15")
	go xp.GeneratesForOneHour(fDate)

	min := current.Minute()
	waitTime := time.Duration(0)
	switch  {
	case min < MINUTES_TIME_GENERATION :
		// wait less than one hour. Add 1s to avoid tricky case xx.39.59.9999
		waitTime = time.Date(current.Year(),current.Month(),current.Day(),current.Hour(),MINUTES_TIME_GENERATION,0,0,current.Location()).Sub(time.Now()) + time.Second
	case min == MINUTES_TIME_GENERATION :
		// Wait the next hour
		waitTime = time.Date(current.Year(),current.Month(),current.Day(),current.Hour()+1,MINUTES_TIME_GENERATION,0,0,current.Location()).Sub(time.Now())
	case min > MINUTES_TIME_GENERATION :
		// no wait, relaunch now
	}
	logger.GetLogger().Info("Wait next generation ",waitTime)
	time.Sleep(waitTime)
	xp.RunAutoGenerations(60 - MINUTES_TIME_GENERATION)
}

func (xp XPlanet)GenerateEvery(duration time.Duration){
	// Find date to launch now
	go xp.Generate(xp.GenerateName("jpg",""),"")
	// Sleep to next full hour
	time.Sleep(duration)
	xp.GenerateEvery(duration)
}

func IsFileYoungerThan(path string,duration time.Duration)bool{
	if f,err := os.Open(path) ; err == nil {
		defer f.Close()
		if info,err := f.Stat() ; err == nil {
			data := info.Sys().(*syscall.Win32FileAttributeData)
			d := time.Unix(0,data.LastWriteTime.Nanoseconds())
			return time.Now().Sub(d) < duration
		}
	}
	return false
}

func isFileUpperThan(path string, exceptedSize int)bool{
	if f,err := os.Open(path) ; err == nil {
		defer f.Close()
		if info,err := f.Stat() ; err == nil {
			return info.Size() > int64(exceptedSize)
		}
	}
	return false
}

// Load cloud
func (xp XPlanet)LoadCloud(duration time.Duration,ackFirstGeneration chan struct{}){
	// Check last date generation. If  < 12h, no relaunch
	if !IsFileYoungerThan(xp.GenerateName("cloud",""),time.Duration(12) * time.Hour) {
		cloudUrl := "http://xplanetclouds.com/free/local/clouds_2048.jpg"
		logger.GetLogger().Info("Launch cloud image save")
		if response, err := http.Get(cloudUrl); err == nil {
			defer response.Body.Close()
			if out, err := os.OpenFile(xp.GenerateName("cloud", ""), os.O_TRUNC | os.O_CREATE | os.O_RDWR, os.ModePerm); err == nil {
				defer out.Close()
				if _, err := io.Copy(out, response.Body); err == nil {
					logger.GetLogger().Info("Cloud image well loaded")
				}
			}else {
				logger.GetLogger().Error("Impossible to load cloud image", err)
			}
		}else {
			logger.GetLogger().Error("Impossible to load cloud image", err)
		}
	}else{
		logger.GetLogger().Info("Cloud image enough young")
	}
	// Check if size < 1Ko and use a default image if necessary
	if(!isFileUpperThan(xp.GenerateName("cloud", ""),1024)){
		defaultInput,_ := os.Open(filepath.Join(xp.ConfigPath,"clouds_default.jpg"))
		defer defaultInput.Close()
		if out, err := os.OpenFile(xp.GenerateName("cloud", ""), os.O_TRUNC | os.O_CREATE | os.O_RDWR, os.ModePerm); err == nil {
			defer out.Close()
			if _, err := io.Copy(out, defaultInput); err == nil {
				logger.GetLogger().Info("Default cloud image used instead")
			}
		}
	}

	// First generation ask,
	if ackFirstGeneration != nil {
		 ackFirstGeneration <- struct{}{}
		close(ackFirstGeneration)
	}
	time.Sleep(duration)
	xp.LoadCloud(duration,nil)
}

// Generate for one specific hour a picture to each 10 minutes and a gif for the hour
// Date format YYYMMDD.HH
func (xp XPlanet)GeneratesForOneHour(datePrefix string) {
	logger.GetLogger().Info("Launch generate for",datePrefix)
	hour := datePrefix[9:]
	// Generate image every 10 minutes
	for minutes := 0 ; minutes < 60 ; minutes+=10 {
		pad :=""
		if minutes == 0 {
			pad= "0"
		}
		filename := xp.GenerateName("jpg",fmt.Sprintf("%s%s%d",hour,pad,minutes))
		xp.Generate(filename,fmt.Sprintf("%s%s%d00",datePrefix,pad,minutes))
	}
	// Generate animate image
	xp.GenerateManyAngle("tmp_gif_" + hour,xp.GenerateName("gif",hour),5,true,fmt.Sprintf("%s%s0000",datePrefix[:9],hour) )
}

func (xp XPlanet)GenerateName(typeImage,date string)string{
	switch typeImage {
	case "jpg":return filepath.Join(xp.folder,fmt.Sprintf("%s_%s.jpg",xp.prefix,date))
	case "gif":return filepath.Join(xp.folder,fmt.Sprintf("%s_anime_%s.gif",xp.prefix,date))
	case "tmp":return filepath.Join(xp.folder,fmt.Sprintf("%s_tmp_%s.jpg",xp.prefix,date))
	case "cloud":return filepath.Join(xp.folder,fmt.Sprintf("clouds_live.jpg"))
	default:return ""
	}
}

// Generate a picture for a specific date (if specified)
func (xp XPlanet)Generate(filename,date string){
	params := append(xp.Coordinates,append(xp.DefaultParameter,"-output", filename,"-config",xp.ConfigPath,"-searchdir",xp.AppPath)...)
	if(date != ""){
		params = append(params,"-date",date)
	}
	cmd := exec.Command(xp.AppPath + ".exe",params...)
	cmd.Run()
	//xplanet.exe -config ../config.conf  -geometry 1600x1200 -output ../out.jpg -date 20161018.190000
}

func (xp XPlanet)GenerateManyAngle(prefix,filename string, step int,onlyGif bool,date string){
	angleFrom := 0
	angleTo := 360
	begin := time.Now()
	nbMax :=int(math.Abs(float64(angleTo - angleFrom))) / step
	nb := 0
	// Create gif at same time
	generateGif := NewGenerateGif(filename,nbMax)
	if onlyGif {
		go generateGif.Run()
	}
	for angle := angleFrom ; angle <= angleTo && nb < nbMax; angle=(step+angle)%360{
		filename := filepath.Join(xp.folder,fmt.Sprintf("%s_%d.jpg",prefix,angle))
		params := append([]string{"-latitude","48","-longitude",fmt.Sprintf("%d",angle)},append(xp.DefaultParameter,"-output",filename,"-config",xp.ConfigPath,"-searchdir",xp.AppPath)...)
		if(date != ""){
			params = append(params,"-date",date)
		}
		cmd := exec.Command(xp.AppPath + ".exe",params...)
		cmd.Run()
		if onlyGif {
			if xp.GenerateAsync {
				go func(name string, pos int) {generateGif.TreatGif(name, pos)}(filename, nb)
			}else{
				// If no async, launch gif integration and wait before launching new xplanet
				generateGif.TreatGif(filename, nb)
			}
		}
		nb++
	}
	if onlyGif {
		generateGif.waiter.Wait()
	}
	logger.GetLogger().Info("Gif time generation",time.Now().Sub(begin))
}

type GenerateGif struct{
	nb int
	outputGif *gif.GIF
	filename string
	chanel chan imgPos
	waiter *sync.WaitGroup
}

func NewGenerateGif(filename string,nb int)GenerateGif{
	waiter := sync.WaitGroup{}
	waiter.Add(nb+1)
	gifOutput := &gif.GIF{}
	gifOutput.Delay = make([]int,nb)
	gifOutput.Image = make([]*image.Paletted,nb)
	return GenerateGif{nb:nb,chanel:make(chan imgPos,15),outputGif:gifOutput,filename:filename,waiter:&waiter}
}

type imgPos struct{
	img image.Image
	pos int
}

//TreatGif read a jpeg file and encode to be used in gif generation, remove the jpeg file after reading
func (gg * GenerateGif)TreatGif(filename string, pos int){
	imgFile,_ := os.Open(filename)
	if img, err := jpeg.Decode(imgFile); err == nil {
		outBuffer := bytes.NewBuffer(nil)
		gif.Encode(outBuffer, img, &gif.Options{})
		imgGif, _ := gif.Decode(outBuffer)
		imgFile.Close()
		os.Remove(filename)
		gg.chanel <- imgPos{imgGif,pos}
	}
}

//Run generate the final gif output : read all generated images in chanel and build gif
func (gg * GenerateGif)Run(){
	for i := 0 ; i < gg.nb ; i++{
		b := <- gg.chanel
		gg.outputGif.Image[b.pos] = b.img.(*image.Paletted)
		gg.waiter.Done()
	}
	gifImg,_ := os.OpenFile(gg.filename,os.O_CREATE|os.O_RDWR|os.O_TRUNC,os.ModePerm)
	defer gifImg.Close()
	if err := gif.EncodeAll(gifImg,gg.outputGif) ; err == nil {
		logger.GetLogger().Info("Generate gif",gg.filename)
	}else{
		logger.GetLogger().Error("Error when generate gif",err)
	}
	gg.waiter.Done()
}