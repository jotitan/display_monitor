package xplanet
import (
    "strings"
    "net/http"
    "encoding/json"
    "io/ioutil"
    "fmt"
    "logger"
)

// Get randomly some flickr pictures
// Implement GalleryManager

var baseFlickrUrl = "http://api.flickr.com/v2/media/search?query={tag}&orderBy=interestingness&pageNumber={page}&pageSize=50&format=json"

var tags = []string{"water","nature","beach","sunset","sky","night","landscape","clouds","sun","lake","bridge"}

type FlickrGallery struct{
    // Tag actually display
    currentTag *int
    // current page
    page * int
    current * int
    // List of pictures
    pictures * []string
}

func NewFlickrGallery()FlickrGallery{
    pictures := make([]string,0)
    fg := FlickrGallery{newIntAddress(),newIntAddress(),newIntAddress(),&pictures}
    fg.FindFolders()
    return fg
}

func newIntAddress()*int{
    i := 0
    return &i
}

func (fg FlickrGallery) Get()(string,error){
    // If not enough photo, load next page
    if *fg.current >= len(*fg.pictures){
        // Load next page
        *fg.page++
        fg.FindFolders()
    }
    pictures := *fg.pictures
    picture :=pictures[*fg.current]
    *fg.current++;
    return picture,nil
}

func (fg FlickrGallery)Change(folder string){
    *fg.currentTag = (*fg.currentTag+1) % len(tags)
    // Restart at first page
    *fg.page = 0
    fg.FindFolders()
}

func (fg FlickrGallery)GetFolderName()string{
    return tags[*fg.currentTag]
}

// Load a new tag
func (fg FlickrGallery)FindFolders(){
    url := strings.Replace(strings.Replace(baseFlickrUrl,"{tag}",tags[*fg.currentTag],-1),"{page}",fmt.Sprintf("%d",*fg.page),-1)
    logger.GetLogger2().Info("Load",url)
    if resp,err := http.Get(url) ; err == nil {
        data,_ := ioutil.ReadAll(resp.Body)
        m := make([]map[string]interface{},0)
        json.Unmarshal(data,&m)
        pictures := make([]string,0,len(m))
        // Extract, for each picture, the correct url for image between 1400 and 2000
        for _,photo := range m {
            sizes := photo["sizes"].([]interface{})
            for _,size := range sizes {
                infoSize := size.(map[string]interface{})
                if width := infoSize["width"].(float64) ; width > 1400 && width < 2000 {
                    pictures = append(pictures,infoSize["location"].(string))
                    break
                }
            }
        }
        if len(pictures) == 0 {
            // load another tag
            fg.Change("")
        }
        *fg.pictures = pictures
    }
    // Restart at first image
    *fg.current = 0
}

func (fg FlickrGallery)Name()string{
    return "Flickr gallery"
}