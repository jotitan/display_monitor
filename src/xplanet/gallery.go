package xplanet

import (
	"os"
	"math/rand"
	"errors"
	"path/filepath"
	"strings"
	"time"
	"logger"
)

var random = rand.New(rand.NewSource(time.Now().UnixNano()))

// interface to implement to create a new gallery manager
type GalleryManager interface {
	// return a random image
	Get()(string,error)
	// change folder if necessary
	Change(folder string)
	//
	GetFolderName()string
	FindFolders()
	Name()string
}

type FolderGallery struct{
	folder string
	nbPictures int
}

func NewFolderGallery(folder string)FolderGallery{
	gallery := FolderGallery{folder:folder}
	// Compute number of pictures inside folder
	if dir,err := os.Open(folder) ; err == nil {
		defer dir.Close()
		if stat,err := dir.Stat() ; err == nil && stat.IsDir() {
			list,_:= dir.Readdir(-1)
			gallery.nbPictures = len(list)
		}
	}
	return gallery
}

func (f FolderGallery)Change(folder string){
	// Change the folder
}

func (f FolderGallery)GetFolderName()string{
	return ""
}

func (f FolderGallery)Name()string{
	return "Folder"
}

func (f FolderGallery)FindFolders(){}

func (f FolderGallery)Get()(string,error){
	if f.nbPictures == 0 {
		return "",errors.New("Empty folder")
	}
	pos := int(random.Int31()) % f.nbPictures
	dd,_ := os.Open(f.folder)
	names,_ := dd.Readdirnames(-1)
	return filepath.Join(f.folder,names[pos]),nil
}

type currentSubGallery struct {
	folder folder
	files []string
	position int
}

type folder struct {
	path string
	name string
}

type RollingFoldersGallery struct{
	root string
	folders * []folder
	currentGallery * currentSubGallery
}

// Add option to reload folders sometimes
func NewRollingFoldersGallery(root string)RollingFoldersGallery {
	begin := time.Now()
	folders := make([]folder,0)
	rolling := RollingFoldersGallery{root:root,folders:&folders,currentGallery:&currentSubGallery{}}
	rolling.FindFolders()

	// Select randomly a folder
	rolling.loadFolder()
	logger.GetLogger().Info("Loading time :",time.Now().Sub(begin))
	logger.GetLogger().Info(rolling.folders)
	return rolling
}

func (rolling RollingFoldersGallery)Name()string{
	return "Rolling folders gallery"
}

/** Search all folders in root folder */
func (rolling RollingFoldersGallery)FindFolders(){
	allFolders := make([]folder,0)
	rootDir,err := os.Open(rolling.root)
	if err != nil {
		logger.GetLogger().Error("Wait for relaunch parse folder",err.Error())
		// Maybe mounting does not already exist...wait one minute before relaunch
		time.Sleep(time.Minute)
		// Relaunch
		if rootDir,err = os.Open(rolling.root) ; err != nil {
			return
		}
	}
	defer rootDir.Close()
	folders,_ := rootDir.Readdir(-1)
	for _,fol := range folders {
		if fol.IsDir() {
			// Parse sub folder if exist, only one level
			fDir,_ := os.Open(filepath.Join(rolling.root,fol.Name()))
			files,_ := fDir.Readdir(-1)
			haveImages := false
			for _,f := range files {
				// If also a folder,
				if f.IsDir() {
					name := formatFolderName(fol.Name()) + " - " + formatFolderName(f.Name())
					allFolders = append(allFolders,folder{filepath.Join(fol.Name(),f.Name()),name})
				} else{
					haveImages = true
				}
			}
			if haveImages {
				allFolders = append(allFolders, folder{fol.Name(),formatFolderName(fol.Name())})
			}
		}
	}
	*rolling.folders = allFolders
	logger.GetLogger().Info("Found",len(*rolling.folders),"folders")
}

func (rolling RollingFoldersGallery)Get()(string,error){
	defer func(){
		// Catch error
		if err := recover() ; err != nil {
			logger.GetLogger2().Error("Error when get image",err,rolling.currentGallery.folder.path,rolling.currentGallery.position,len(rolling.currentGallery.files))
		}
	}()
	if rolling.currentGallery == nil {
		return "",errors.New("Impossible, no photo here")
	}
	if rolling.currentGallery.position >= len(rolling.currentGallery.files) -1 {
		rolling.loadFolder()
	}
	file := filepath.Join(rolling.currentGallery.folder.path,rolling.currentGallery.files[rolling.currentGallery.position])
	rolling.currentGallery.position++
	return file,nil
}

func (rolling RollingFoldersGallery)Change(folder string){
	// Change the folder
	rolling.loadFolder()
}

func formatFolderName(name string)string{
	return strings.ToLower(strings.Replace(name,"_"," ",-1))
}

func (rolling RollingFoldersGallery)GetFolderName()string{
	// Change the folder
	return rolling.currentGallery.folder.name
}

func (rolling * RollingFoldersGallery)loadFolder(){
	if len(*rolling.folders) == 0 {
		return
	}
	folders := *rolling.folders
	position := int(random.Int31()) % len(folders)
	selectedFolder := filepath.Join(rolling.root,folders[position].path)
	fol,_ := os.Open(selectedFolder)
	files,_ := fol.Readdir(-1)
	filteredFiles := make([]string,0,len(files))
	for _,file := range files {
		if !file.IsDir() && (strings.HasSuffix(strings.ToLower(file.Name()),".jpg") || strings.HasSuffix(strings.ToLower(file.Name()),".jpeg")) {
			filteredFiles = append(filteredFiles,file.Name())
		}
	}
	if len(filteredFiles) == 0 {
		// Search an another one
		rolling.loadFolder()
		return
	}
	logger.GetLogger().Info("Select gallery folder",selectedFolder)
	*(rolling.currentGallery) = currentSubGallery{folder:folder{selectedFolder,folders[position].name},position:0,files:filteredFiles}
}