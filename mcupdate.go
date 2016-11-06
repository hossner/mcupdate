package main

import(
	"log"
	"golang.org/x/net/html"
	"net/http"
	"strings"
	"regexp"
	"io"
	"io/ioutil"
	"os"
)

const (
	dirPath = "/var/minecraft/"
	//dirPath = "/home/carl/Dev/go/src/github.com/chossner/mcupdate/"
	dlDir = "_MinecraftVersions/"
	linkName = "current_version"
	mcURL = "https://minecraft.net/en/download/server"
	latest = "latest.txt"
)

func downloadFromURL(url string) (err error){
	tokens := strings.Split(url, "/")
	fileName := tokens[len(tokens)-1]
	log.Println("Downloading", url, "to", dirPath+dlDir+fileName)

	// TODO: check file existence first with io.IsExist
	output, err := os.Create(dirPath+dlDir+fileName)
	if err != nil {
		log.Println("Error while creating", dirPath+dlDir+fileName, "-", err)
		return err
	}
	defer output.Close()

	response, err := http.Get(url)
	if err != nil {
		log.Println("Error while downloading", url, "-", err)
		return err
	}
	defer response.Body.Close()

	n, err := io.Copy(output, response.Body)
	if err != nil {
		log.Println("Error while downloading", url, "-", err)
		return err
	}
	log.Println(n, "bytes downloaded.")
	return nil
}


func getCurrent() (err error, current string){
	res, err := ioutil.ReadFile(dirPath+latest)
	if err == nil {
		return nil, strings.TrimSpace(string(res))
	}
	return err, ""
}

func putCurrent(cv string) (err error) {
	err = ioutil.WriteFile(dirPath+latest, []byte(cv), 644)
	return
}

func getHref(t html.Token) (ok bool, href string) {
	for _, a := range t.Attr {
		if a.Key == "href"{
			href = a.Val
			ok = true
		}
	}
	return
}



func main(){
	log.Println("Using URL "+mcURL)
	resp, err := http.Get(mcURL)
	if err != nil {
		log.Println("Error fetching URL "+mcURL)
		return
	}
	b := resp.Body
	defer b.Close()
	z := html.NewTokenizer(b)
	var mcLink = regexp.MustCompile(`minecraft_server\.([0-9]+\.[0-9]{1,2}(\.[0-9]{1,2})?)\.jar$`)
	for {
		tt := z.Next()
		
		switch {
		case tt == html.ErrorToken:
			return
		case tt == html.StartTagToken:
			t := z.Token()
			if t.Data != "a" {
				continue
			}
			ok, url := getHref(t)
			if !ok {
				continue
			}
			if strings.Index(url, "https") == 0 {
				//matched, _ := regexp.MatchString("minecraft_server..*.jar", url)
				//matched := mcLink.MatchString(url)
				matched := mcLink.FindStringSubmatch(url)
				if matched != nil{
					if matched[0] == "" || matched[1] == "" {
						log.Println("Looked at "+url+ "but found nothing!")
						return
					}
					log.Println("Found "+matched[0]+" version "+matched[1]+" at "+url)
					err, running := getCurrent()
					if err == nil {
						if matched[1] == running {
							log.Println("We are running current!")
						} else {
							log.Println("We need to update from "+running+" to "+matched[1])
							if downloadFromURL(url) != nil {
								return
							}
							if _, err := os.Stat(dirPath+linkName); err == nil {
								log.Println("Removing old symlink...")
								os.Remove(dirPath+linkName)
							}
							os.Symlink(dirPath+dlDir+matched[0], dirPath+linkName)
							log.Println("Creating new symlink...")
							putCurrent(matched[1])
							// Update file with version number
						}
					}
				}
			}
		}
	}
}

