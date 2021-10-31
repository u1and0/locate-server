package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
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
	// REQUIRE : required commands. Separate by space.
	REQUIRE = "locate gocate"
)

var (
	log         = logging.MustGetLogger("main")
	showVersion bool
	port        string
)

func main() {
	var (
		stats   = cmd.Stats{}
		locater = parse()
	)

	if showVersion {
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

	// Command check
	for _, r := range strings.Fields(REQUIRE) {
		if _, err := exec.LookPath(r); err != nil {
			log.Panicf("%s", err.Error())
		}
	}

	// Open server
	route := gin.Default()
	route.Static("/static", "./static")
	route.LoadHTMLGlob("templates/*")
	stats.LastUpdateTime = cmd.DBLastUpdateTime(locater.Dbpath)

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
		locater.SearchWords = strings.Fields(q)
		result, err := locater.Locate()
		if err != nil {
			log.Error(err)
			locater.Status = 404
			c.JSON(404, locater)
		} else {
			locater.Paths = result
			locater.Status = http.StatusOK
			c.JSON(http.StatusOK, locater)
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
	route.Run(":" + port)
}

// Parse command line option
func parse() (l cmd.Locater) {
	flag.StringVar(&l.Dbpath, "d", LOCATEDIR, "Path of locate database directory")
	flag.IntVar(&l.Limit, "l", 1000, "Maximum limit for results")
	flag.BoolVar(&l.PathSplitWin, "s", false, "OS path split windows backslash")
	flag.StringVar(&l.Root, "r", "", "DB insert prefix for directory path")
	flag.StringVar(&l.Trim, "t", "", "DB trim prefix for directory path")
	flag.BoolVar(&l.Debug, "debug", false, "Debug mode")
	flag.StringVar(&port, "p", "8080", "Server port number. Default access to http://localhost:8080/")
	flag.StringVar(&port, "port", "8080", "Server port number. Default access to http://localhost:8080/")
	flag.BoolVar(&showVersion, "v", false, "show version")
	flag.BoolVar(&showVersion, "version", false, "show version")
	flag.Parse()
	return
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
