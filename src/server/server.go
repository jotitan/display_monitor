package server

import (
	"net/http"
	"path/filepath"
	"xplanet"
	"os"
	"io"
	"logger"
	"strings"
)
//

type Server struct {
	webFolder string
	galleryManager xplanet.GalleryManager
	xPlanetManagerEarth xplanet.XPlanet
	xPlanetManagerMoon xplanet.XPlanet
}

func (s Server)getPlanetManager(planet string)*xplanet.XPlanet{
	switch planet {
	case "earth" : return &s.xPlanetManagerEarth
	case "moon":return &s.xPlanetManagerMoon
	default : return nil
	}
}


func (s Server)change(response http.ResponseWriter, request *http.Request) {
	s.galleryManager.Change(request.FormValue("folder"))
}

func (s Server)getFolderName(response http.ResponseWriter, request *http.Request) {
	response.Write([]byte(s.galleryManager.GetFolderName()))
}

func extractRequesterHost(request *http.Request)string{
	if !strings.Contains(request.Host,":"){
		return request.Host
	}
	return request.Host[:strings.Index(request.Host,":")]
}

// TurnOn turn some things on only if requester is same as server (ip)
func (s Server)turnOn(response http.ResponseWriter, request *http.Request) {
	url := "http://" + extractRequesterHost(request) + ":9098/turnOn"
	logger.GetLogger().Info("Wake up :",url)
	if _,e := http.Get(url) ; e != nil {
		logger.GetLogger().Error(e)
	}
}

// TurnOff turn some things off only if requester is same as server (ip)
func (s Server)turnOff(response http.ResponseWriter, request *http.Request) {
	url := "http://" + extractRequesterHost(request) + ":9098/turnOff"
	logger.GetLogger().Info("Run sleep mode :",url)
	if _,e := http.Get(url) ; e != nil {
		logger.GetLogger().Error(e)
	}
}

//getImage return an image of xplanet
// Precise 3 parameter : planet (earth, moon), format (gif, jpeg), hour
func (s Server)getImage(response http.ResponseWriter, request *http.Request) {
	if xPlanet := s.getPlanetManager(request.FormValue("planet")); xPlanet != nil {
		format := request.FormValue("format")
		filename := xPlanet.GenerateName(format, request.FormValue("date"))
		if f, err := os.Open(filename); err != nil {
			http.Error(response, "Image not found", 404)
		}else {
			defer f.Close()
			response.Header().Add("Content-Type", "image/" + format)
			io.Copy(response, f)
		}
	}else{
		// Random image case, return image in folder
		if filename,err := s.galleryManager.Get() ; err == nil {
			f,_:= os.Open(filename)
			defer f.Close()
			io.Copy(response,f)
		}
	}
}

func (s Server)root(response http.ResponseWriter, request *http.Request){
	logger.GetLogger().Info("Serve file",request.RequestURI)
	if url := request.RequestURI ; url == "/"{
		http.ServeFile(response,request,filepath.Join(s.webFolder,"index.html"))
	}else{
		http.ServeFile(response,request,filepath.Join(s.webFolder,url[1:]))
	}
}

func (s Server)Launch(port string){
	smux := http.NewServeMux()
	// create paths
	if strings.EqualFold(port,""){
		port = "8000"
	}
	smux.HandleFunc("/image",s.getImage)
	smux.HandleFunc("/change",s.change)
	smux.HandleFunc("/getFolderName",s.getFolderName)
	smux.HandleFunc("/turnOn",s.turnOn)
	smux.HandleFunc("/turnOff",s.turnOff)
	smux.HandleFunc("/",s.root)

	logger.GetLogger().Info("server is well loaded on " + port)
	// launch server
	http.ListenAndServe(":" + port,smux)
}

func New(webFolder string,xpEarth xplanet.XPlanet,xpMoon xplanet.XPlanet,gallery string)Server{
	s := Server{webFolder:webFolder,xPlanetManagerEarth:xpEarth,xPlanetManagerMoon:xpMoon}
	if strings.HasPrefix(gallery,"http"){
		// Rest case
	}else{
		s.galleryManager = xplanet.NewRollingFoldersGallery(gallery)
		//s.galleryManager = xplanet.NewFolderGallery(gallery)
	}
	return s
}
