package server

import (
	"net/http"
	"path/filepath"
	"xplanet"
	"os"
	"io"
	"logger"
	"strings"
	"fmt"
)
//

type Server struct {
	webFolder string
	// store the current manager of gallery (rolling folder or flickr)
	currentGalleryManager * xplanet.GalleryManager
	flickrManager *xplanet.GalleryManager
	galleryManager *xplanet.GalleryManager
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

func (s Server)getCurrentGallery()xplanet.GalleryManager{
	g := *(s.currentGalleryManager)
	return g
}

func (s * Server)change(response http.ResponseWriter, request *http.Request) {
	s.getCurrentGallery().Change(request.FormValue("folder"))
}

func (s * Server)changeGallery(response http.ResponseWriter, request *http.Request) {
	if strings.Contains(strings.ToLower(s.getCurrentGallery().Name()),"flickr") {
		s.currentGalleryManager = s.galleryManager
	}else{
		s.currentGalleryManager = s.flickrManager
	}
	logger.GetLogger2().Info("Change to gallery",s.getCurrentGallery().Name())
}

func (s * Server)getFolderName(response http.ResponseWriter, request *http.Request) {
	response.Write([]byte(s.getCurrentGallery().GetFolderName()))
}

func (s * Server)findFolders(response http.ResponseWriter, request *http.Request) {
	s.getCurrentGallery().FindFolders()
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
func (s *Server)getImage(response http.ResponseWriter, request *http.Request) {
	if xPlanet := s.getPlanetManager(request.FormValue("planet")); xPlanet != nil {
		s.loadPlanet(response,xPlanet,request.FormValue("date"),request.FormValue("format"))
	}else{
		s.loadImage(response)
	}
}

func (s * Server)loadPlanet(response http.ResponseWriter,xPlanet *xplanet.XPlanet, date, format string){
	// Random image case, return image in folder
	filename := xPlanet.GenerateName(format, date)
	if f, err := os.Open(filename); err != nil {
		http.Error(response, "Image not found", 404)
	}else {
		defer f.Close()
		response.Header().Add("Content-Type", "image/" + format)
		io.Copy(response, f)
	}
}

// return image from gallery
func (s * Server)loadImage(response http.ResponseWriter){
	// Random image case, return image in folder
	if filename,err := s.getCurrentGallery().Get() ; err == nil {
		// Http case
		if strings.HasPrefix(filename,"http") {
			if resp,err := http.Get(filename) ; err == nil {
				io.Copy(response,resp.Body)
			}
		}else {
			f, _ := os.Open(filename)
			defer f.Close()
			io.Copy(response, f)
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
	smux.HandleFunc("/changeGallery",s.changeGallery)
	smux.HandleFunc("/getFolderName",s.getFolderName)
	smux.HandleFunc("/findFolders",s.findFolders)
	smux.HandleFunc("/turnOn",s.turnOn)
	smux.HandleFunc("/turnOff",s.turnOff)
	smux.HandleFunc("/",s.root)

	logger.GetLogger().Info(fmt.Sprintf("server is well loaded on %s (%v)",port,s.xPlanetManagerEarth.GenerateAsync))
	// launch server
	http.ListenAndServe(":" + port,smux)
}

func New(webFolder string,xpEarth xplanet.XPlanet,xpMoon xplanet.XPlanet,gallery string)Server{
	s := Server{webFolder:webFolder,xPlanetManagerEarth:xpEarth,xPlanetManagerMoon:xpMoon}
	if strings.HasPrefix(gallery,"http"){
		// Rest case
	}else{
		rfg := xplanet.GalleryManager(xplanet.NewRollingFoldersGallery(gallery))
		fg := xplanet.GalleryManager(xplanet.NewFlickrGallery())
		s.galleryManager = &rfg
		s.flickrManager = &fg
		// By default, gallery is rolling folders
		s.currentGalleryManager = s.galleryManager

	}
	return s
}
