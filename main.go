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

// Args command line flag options
type Args struct {
	Dbpath       string
	Limit        int
	PathSplitWin bool
	Root         string
	Trim         string
	Debug        bool
	ShowVersion  bool
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
		q := c.Query("q")
		l := cmd.Locater{
			SearchWords:  strings.Fields(q),
			Dbpath:       args.Dbpath,
			PathSplitWin: args.PathSplitWin,
			Root:         args.Root,
			Trim:         args.Trim,
			Debug:        args.Debug,
		}
		result, err := l.Locate()
		if err != nil {
			log.Error(err)
			l.Status = 404
			c.JSON(404, l)
		} else {
			l.Paths = result
			l.Status = http.StatusOK
			c.JSON(http.StatusOK, l)
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
		ShowVersion:  showVersion,
		Limit:        limit,
		Dbpath:       dbpath,
		PathSplitWin: pathSplitWin,
		Root:         root,
		Trim:         trim,
		Debug:        debug,
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
