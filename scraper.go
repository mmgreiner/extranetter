package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/gocolly/colly"
)

// http://go-colly.org/docs/introduction/start/

func main() {
	downloadDir := flag.String("download", "./downloads/", "folder where to download all the files to")
	var user, pwd, top string
	flag.StringVar(&user, "user", "", "user name for extranet, example markus.greiner@pfarrei-aarau.ch")
	flag.StringVar(&pwd, "password", "", "password of extranet")
	flag.StringVar(&top, "top", "home",
		"top directory where to start scraping, like 'home', 'documents', 'personal', ...")
	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintln(w, "Scraps the extranet of pastoralraum Aarau and downloads all the files")
		flag.PrintDefaults()
	}
	flag.Parse()

	if user == "" || pwd == "" {
		flag.Usage()
		fmt.Fprintln(os.Stderr, "username and password are required, aborting")
		os.Exit(1)
	}

	const (
		pastoralraum = "https://extranet.pastoralraum-aarau.ch"
		media        = pastoralraum + "/media/web/extranet.pastoralraum-aarau.ch/media/"
		itds         = "https://secure.itds.ch/itds-auth/login.jsp?redirectTo=https%3A%2F%2Fextranet.pastoralraum-aarau.ch%2F"
	)

	visited := make(map[string]bool)

	c := colly.NewCollector()

	err := c.Post(itds, map[string]string{"username": user, "password": pwd})
	if err != nil {
		log.Fatal(err)
	}

	// only parse the menu when the top is home
	if top == "home" {
		c.OnHTML("li.jwa_menu_entry a[href]", func(e *colly.HTMLElement) {
			link := e.Attr("href")
			//fmt.Println("MENU", e.Text, link) // /de/sitzungen
			if _, exists := visited[link]; !exists {
				visited[link] = true
				c.Visit(pastoralraum + link)

			}
		})
	}

	// <td><a href="?waxmlc_dbFolder_2338099=2532705"><i class="fa fa-folder"></i> <span>Sitzung vom 2023_01_12</span></a></td>
	c.OnHTML("a[href]:has(i.fa-folder)", func(e *colly.HTMLElement) {
		link := e.Attr("href")

		requestUrl := e.Request.URL
		requestUrl.RawQuery = ""
		newUrl := requestUrl.String() + link
		log.Println("request", requestUrl, "->", newUrl)
		c.Visit(newUrl)
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if strings.HasPrefix(link, "/media/web") {
			c.Visit(e.Request.AbsoluteURL(link))
		}
	})

	// https://github.com/gocolly/colly/issues/193
	c.OnResponse(func(r *colly.Response) {
		urlname := r.Request.URL.String()
		contains := strings.Contains(urlname, "media")
		log.Println("visiting", r.Request.URL, "media=", contains)
		if contains {
			relpath := strings.Replace(urlname, media, "", 1)
			relpath, _ = url.PathUnescape(relpath)
			// if it doesnt have an extension, add pdf
			if filepath.Ext(relpath) == "" {
				relpath = relpath + ".pdf"
			}

			// make the directory if it does not exist
			fullpath := filepath.Join(*downloadDir, relpath)
			fulldir := filepath.Dir(fullpath)
			if err := os.MkdirAll(fulldir, 0777); err != nil {
				log.Println("!!! Error mkdir", fulldir, err.Error())
			} else {
				log.Println("saving", fullpath)
				if err := r.Save(fullpath); err != nil {
					log.Println("!!! Error saving", fullpath, err.Error())
				}
			}

		}
	})

	topUrl, err := url.JoinPath(pastoralraum, "de", top)
	if err != nil {
		panic(err)
	}
	log.Println("starting", topUrl)
	c.Visit(topUrl)

	//c.Visit("https://extranet.pastoralraum-aarau.ch/de/adressen_behoerden?waxmlc_dbFolder_2338113=1636817")
	log.Println("Done")
}

// https://extranet.pastoralraum-aarau.ch/de/archiv_sitzungen?waxmlc_dbFolder_2338127=1807677
