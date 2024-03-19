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
	"github.com/subosito/gotenv"
)

// http://go-colly.org/docs/introduction/start/

const (
	pastoralraum = "https://extranet.pastoralraum-aarau.ch"
	media        = pastoralraum + "/media/web/extranet.pastoralraum-aarau.ch/media/"
	itds         = "https://secure.itds.ch/itds-auth/login.jsp?redirectTo=https%3A%2F%2Fextranet.pastoralraum-aarau.ch%2F"
	//pfarraarau = "https://aarau.pastoralraum-aarau.ch/de/home"
	helpText = `Downloads files from the extranet of Pastoralraum Aarau
Copyright Pastoralraum Aarau, contact: markus.greiner@pfarrei-aarau.ch
Two main commands:
- top: starts downloading at the top level of pastoralraum
- link: you have to give a direct link to a subfolder (copied from the browser)
`
)

type Options struct {
	// common flags
	download string
	user     string
	password string
	silent   bool
	debug    bool
	isTop    bool
	// only with isTop
	top    string
	folder string
	// only with link
	url      string
	startUrl string
}

func handleOptions() (Options, error) {
	linkCmd := flag.NewFlagSet("link", flag.ContinueOnError)
	topCmd := flag.NewFlagSet("top", flag.ContinueOnError)

	var downloadDir, user, pwd, top, folder, startUrl string
	var silent, debugging bool

	// options only for link
	sampleStartUrl := "https://extranet.pastoralraum-aarau.ch/de/sitzungen?waxmlc_dbFolder_2338099=2532705"
	linkCmd.StringVar(&startUrl, "url", "", "url where to start downloading, copied from browser, ex "+sampleStartUrl)

	// options only for top
	topCmd.StringVar(&top, "top", "home",
		"top folder where to start scraping, like 'home', 'personal', 'sitzungen', 'archiv_sitzungen', 'budget', ...")
	topCmd.StringVar(&folder, "folder", "", "folder inside top-folder which to download")

	// common sub commands
	for _, cmd := range []*flag.FlagSet{linkCmd, topCmd} {
		cmd.StringVar(&downloadDir, "download", "./downloads/", "folder where to download all the files to")
		cmd.StringVar(&user, "user", "", "user name for extranet, example markus.greiner@pfarrei-aarau.ch")
		cmd.StringVar(&pwd, "password", "", "password of extranet")

		cmd.BoolVar(&silent, "silent", false, "print all files, but does not download them")
		cmd.BoolVar(&debugging, "debug", false, "print debugging messages")
	}
	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprint(w, helpText)
		fmt.Fprintln(w, "Options:")
		topCmd.Usage()
		linkCmd.Usage()
		flag.PrintDefaults()
	}

	gotenv.Load(".env")

	checkUser := func() error {
		// allow to set via environment variables
		if user == "" {
			user = os.Getenv("user")
		}
		if pwd == "" {
			pwd = os.Getenv("password")
		}

		if user == "" || pwd == "" {
			return fmt.Errorf("user and password cannot be empty")
		} else {
			return nil
		}
	}

	if len(os.Args) < 2 {
		return Options{}, fmt.Errorf("no subcommands given")
	}
	switch os.Args[1] {
	case "top":
		topCmd.Parse(os.Args[2:])
		topUrl, err := url.JoinPath(pastoralraum, "de", top)
		if err != nil {
			return Options{}, err
		}
		if err := checkUser(); err != nil {
			return Options{}, err
		}
		return Options{
			download: downloadDir,
			user:     user,
			password: pwd,
			silent:   silent,
			debug:    debugging,
			top:      top,
			folder:   folder,
			startUrl: topUrl,
		}, nil
	case "link":
		linkCmd.Parse(os.Args[2:])
		if err := checkUser(); err != nil {
			return Options{}, err
		}
		return Options{
			download: downloadDir,
			user:     user,
			password: pwd,
			silent:   silent,
			debug:    debugging,
			startUrl: startUrl,
		}, nil
	case "-h", "--h", "-help", "--help":
		flag.Usage()
		os.Exit(0)
		return Options{}, nil
	default:
		return Options{}, fmt.Errorf("wrong subcommand %s", os.Args[1])
	}
}

// https://aarau.pastoralraum-aarau.ch/de/intern/okp

func main() {

	options, err := handleOptions()
	if err != nil {
		panic(err)
	}

	// handle logging
	level := slog.LevelInfo
	if options.debug {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)

	// define starting position

	visited := make(map[string]bool)

	c := colly.NewCollector()

	err = c.Post(itds, map[string]string{"username": options.user, "password": options.password})
	if err != nil {
		panic(err)
	}

	// only parse the menu when the top is home
	if options.top == "home" {
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
			if !strings.Contains(relpath, options.folder) {
				return
			}
			// if it doesnt have an extension or has a strange extension, add pdf
			ext := filepath.Ext(relpath)
			if ext == "" || strings.Contains(ext, " ") {
				relpath = relpath + ".pdf"
			}

			// make the directory if it does not exist
			fullpath := filepath.Join(options.download, relpath)
			fulldir := filepath.Dir(fullpath)
			if options.silent {
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

	slog.Info("starting", "starturl", options.startUrl)
	c.Visit(options.startUrl)

	//c.Visit("https://extranet.pastoralraum-aarau.ch/de/adressen_behoerden?waxmlc_dbFolder_2338113=1636817")
	slog.Info("Done")
}
