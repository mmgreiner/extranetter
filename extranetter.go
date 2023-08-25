package main

import (
	"flag"
	"fmt"

	"golang.org/x/exp/slog"

	//"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/gocolly/colly"
)

// http://go-colly.org/docs/introduction/start/

func main() {
	downloadDir := flag.String("download", "./downloads/", "folder where to download all the files to")
	var user, pwd, top, folder string
	flag.StringVar(&user, "user", "", "user name for extranet, example markus.greiner@pfarrei-aarau.ch")
	flag.StringVar(&pwd, "password", "", "password of extranet")
	flag.StringVar(&top, "top", "home",
		"top folder where to start scraping, like 'home', 'personal', 'sitzungen', 'archiv_sitzungen', 'budget', ...")
	flag.StringVar(&folder, "folder", "", "folder inside top-folder which to download")
	//tree := flag.Bool("index", false, "show the tree of the folders")
	silent := flag.Bool("silent", false, "print all files, but does not download them")
	debugging := flag.Bool("debug", false, "print debugging messages")

	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintln(w, "Downloads files from the extranet of Pastoralraum Aarau")
		fmt.Fprintln(w, "Copyright Pastoralraum Aarau, contact: markus.greiner@pfarrei-aarau.ch")
		fmt.Fprintln(w, "Options:")
		flag.PrintDefaults()
	}
	flag.Parse()

	if user == "" || pwd == "" {
		flag.Usage()
		println("username and password are required, aborting")
		os.Exit(1)
	}

	// handle logging
	level := slog.LevelInfo
	if *debugging {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)

	const (
		pastoralraum = "https://extranet.pastoralraum-aarau.ch"
		media        = pastoralraum + "/media/web/extranet.pastoralraum-aarau.ch/media/"
		itds         = "https://secure.itds.ch/itds-auth/login.jsp?redirectTo=https%3A%2F%2Fextranet.pastoralraum-aarau.ch%2F"
	)

	visited := make(map[string]bool)

	c := colly.NewCollector()

	err := c.Post(itds, map[string]string{"username": user, "password": pwd})
	if err != nil {
		panic(err)
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
		slog.Debug("folder", "requestUrl", requestUrl, "newUrl", newUrl)
		c.Visit(newUrl)
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if strings.HasPrefix(link, "/media/web") {
			c.Visit(e.Request.AbsoluteURL(link))
		}
	})

	/*
	 <td><a target="_blank" href="/media/web/extranet.pastoralraum-aarau.ch/media/Sitzungen/Sitzung%20vom%202023_07_04_baulicher%20Unterhalt/Baulicher%20Unterhalt%202024_Tischvorlage%20vom%2029.%20Juni"><i class="fa fa-file-pdf-o"></i> <span>Baulicher Unterhalt 2024_Tischvorlage vom 29. Juni</span></a></td>
	*/
	// https://github.com/gocolly/colly/issues/193
	c.OnResponse(func(r *colly.Response) {
		urlname := r.Request.URL.String()
		contains := strings.Contains(urlname, "media")
		slog.Debug("visiting", "request.URL", r.Request.URL, "media=", contains)
		if contains {
			relpath := strings.Replace(urlname, media, "", 1)
			relpath, _ = url.PathUnescape(relpath)

			// move to next node if the relpath does not contain the folder we are looking for
			if !strings.Contains(relpath, folder) {
				return
			}
			// if it doesnt have an extension or has a strange extension, add pdf
			ext := filepath.Ext(relpath)
			if ext == "" || strings.Contains(ext, " ") {
				relpath = relpath + ".pdf"
			}

			// make the directory if it does not exist
			fullpath := filepath.Join(*downloadDir, relpath)
			fulldir := filepath.Dir(fullpath)
			if *silent {
				fmt.Println("seeing:", relpath)
			} else {
				if err := os.MkdirAll(fulldir, 0777); err != nil {
					slog.Error("Error mkdir", "fulldir", fulldir, "err", err.Error())
				} else {
					slog.Info("saving", "fullpath", fullpath)
					if err := r.Save(fullpath); err != nil {
						slog.Error("Error saving", "fullpath", fullpath, "err", err.Error())
					}
				}
			}

		}
	})

	topUrl, err := url.JoinPath(pastoralraum, "de", top)
	if err != nil {
		panic(err)
	}
	slog.Info("starting", "topUrl", topUrl)
	c.Visit(topUrl)

	//c.Visit("https://extranet.pastoralraum-aarau.ch/de/adressen_behoerden?waxmlc_dbFolder_2338113=1636817")
	slog.Info("Done")
}

// https://extranet.pastoralraum-aarau.ch/de/archiv_sitzungen?waxmlc_dbFolder_2338127=1807677
