package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	cmd "github.com/u1and0/locate-server/cmd"

	"github.com/gin-gonic/gin"
	"github.com/op/go-logging"
)

const (
	// VERSION : version
	VERSION = "2.3.2r"
	// LOGFILE : 検索条件 / 検索結果 / 検索時間を記録するファイル
	LOGFILE = "/var/lib/mlocate/locate.log"
	// LOCATEDIR : locate (gocate) search db path
	LOCATEDIR = "/var/lib/mlocate"
)

// Paths locate command result
type Paths []string

// Args command line flag options
type Args struct {
	Dbpath      string
	Limit       int
	Root        string
	Trim        string
	Debug       bool
	ShowVersion bool
}

// Result return JSON struct
type Result struct {
	Paths  `json:"paths"`
	Status int    `json:"status"`
	Err    error  `json:"error"`
	Query  string `json:"query"`
}

var log = logging.MustGetLogger("main")

func main() {
	var (
		stats = cmd.Stats{}
		args  = parse()
	)

	if args.ShowVersion {
		fmt.Println("locate-server version", VERSION)
		return // versionを表示して終了
	}

	// Log setting
	logfile, err := os.OpenFile(LOGFILE, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	defer logfile.Close()
	setLogger(logfile) // log.XXX()を使うものはここより後に書く
	if err != nil {
		log.Panicf("Cannot open logfile %v", err)
	}

	// Open server
	route := gin.Default()
	route.Static("/static", "./static")
	route.LoadHTMLGlob("templates/*")
	stats.LastUpdateTime = cmd.DBLastUpdateTime(args.Dbpath)

	route.GET("/", func(c *gin.Context) {
		datalist, err := cmd.Datalist(LOGFILE)
		if err != nil {
			log.Error(err)
		}
		c.HTML(http.StatusOK,
			"index.tmpl",
			gin.H{
				"title":          "",
				"lastUpdateTime": stats.LastUpdateTime,
				"datalist":       datalist,
				"query":          "",
			})
	})

	route.GET("/search", func(c *gin.Context) {
		datalist, err := cmd.Datalist(LOGFILE)
		if err != nil {
			log.Error(err)
		}
		query := c.Request.URL.Query()
		q := strings.Join(query["q"], " ")
		c.HTML(http.StatusOK,
			"index.tmpl",
			gin.H{
				"title":          q,
				"lastUpdateTime": stats.LastUpdateTime,
				"datalist":       datalist,
				"query":          q,
			})
	})

	route.GET("/json", func(c *gin.Context) {
		result, err := ResultPath(c)
		if err != nil {
			result.Err = err
			c.JSON(404, result)
		} else {
			result.Status = http.StatusOK
			c.JSON(http.StatusOK, result)
		}
	})

	route.GET("/status", func(c *gin.Context) {
		db := "/var/lib/mlocate"
		l, err := cmd.LocateStats(db)
		ss := strings.Split(string(l), "\n")
		c.JSON(http.StatusOK, gin.H{
			"result": ss,
			"status": http.StatusOK,
			"error":  err,
		})
	})

	// Listen and serve on 0.0.0.0:8080
	route.Run(":8080")
}

// ResultPath execute locate and return
func ResultPath(c *gin.Context) (Result, error) {
	q := c.Query("q")
	l := cmd.Locater{
		SearchWords: strings.Fields(q),
		Dbpath:      "/var/lib/mlocate",
		Debug:       false,
	}
	path, err := l.Locate()
	result := Result{
		Paths: path,
		Query: q,
	}
	return result, err
}

// Parse command line option
func parse() Args {
	var (
		showVersion  bool
		limit        int
		dbpath       string
		pathSplitWin bool
		root         string
		trim         string
		debug        bool
	)

	flag.StringVar(&dbpath, "d", LOCATEDIR, "Path of locate database directory")
	flag.IntVar(&limit, "l", 1000, "Maximum limit for results")
	flag.BoolVar(&pathSplitWin, "s", false, "OS path split windows backslash")
	flag.StringVar(&root, "r", "", "DB insert prefix for directory path")
	flag.StringVar(&trim, "t", "", "DB trim prefix for directory path")
	flag.BoolVar(&debug, "debug", false, "Debug mode")
	flag.BoolVar(&showVersion, "v", false, "show version")
	flag.BoolVar(&showVersion, "version", false, "show version")
	flag.Parse()

	args := Args{
		Dbpath:      dbpath,
		Limit:       limit,
		Root:        root,
		Trim:        trim,
		Debug:       debug,
		ShowVersion: showVersion,
	}
	return args
}

// setLogger is printing out log message to STDOUT and LOGFILE
func setLogger(f *os.File) {
	var format = logging.MustStringFormatter(
		`%{color}[%{level:.6s}] ▶ %{time:2006-01-02 15:04:05} %{shortfile} %{message} %{color:reset}`,
	)
	backend1 := logging.NewLogBackend(os.Stdout, "", 0)
	backend2 := logging.NewLogBackend(f, "", 0)
	backend1Formatter := logging.NewBackendFormatter(backend1, format)
	backend2Formatter := logging.NewBackendFormatter(backend2, format)
	logging.SetBackend(backend1Formatter, backend2Formatter)
}
