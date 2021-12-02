package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	cmd "github.com/u1and0/locate-server/cmd"

	"github.com/gin-gonic/gin"
	"github.com/op/go-logging"
)

const (
	// VERSION : version
	VERSION = "3.0.0r"
	// APIVERSION : api version
	APIVERSION = "0.1.0"
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

func parseCmdlineOption() (l cmd.Locater) {
	flag.StringVar(&l.Args.Dbpath, "d", LOCATEDIR, "Path of locate database directory")
	flag.StringVar(&l.Args.Dbpath, "dir", LOCATEDIR, "Path of locate database directory")
	flag.BoolVar(&l.Args.PathSplitWin, "s", false, "OS path split windows backslash")
	flag.BoolVar(&l.Args.PathSplitWin, "windows-path-separate", false, "OS path separate windows backslash")
	flag.StringVar(&l.Args.Root, "r", "", "DB insert prefix for directory path")
	flag.StringVar(&l.Args.Root, "root", "", "DB insert prefix for directory path")
	flag.StringVar(&l.Args.Trim, "t", "", "DB trim prefix for directory path")
	flag.StringVar(&l.Args.Trim, "trim", "", "DB trim prefix for directory path")
	flag.BoolVar(&l.Args.Debug, "debug", false, "Debug mode")
	flag.StringVar(&port, "p", "8080", "Server port number. Default access to http://localhost:8080/")
	flag.StringVar(&port, "port", "8080", "Server port number. Default access to http://localhost:8080/")
	flag.BoolVar(&showVersion, "v", false, "show version")
	flag.BoolVar(&showVersion, "version", false, "show version")
	flag.Parse()
	l.Version = APIVERSION
	return
}

func main() {
	var (
		locateS []byte
		cache   = cmd.CacheMap{}
		locater = parseCmdlineOption()
	)

	if showVersion {
		fmt.Println("locate-server version", VERSION)
		return // versionを表示して終了
	}

	// Release mode
	fmt.Println("locater.Args.Debug", locater.Args.Debug)
	if !locater.Args.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	// Log setting
	logfile, err := os.OpenFile(LOGFILE, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	defer logfile.Close()
	setLogger(logfile) // log.XXX()を使うものはここより後に書く
	if err != nil {
		log.Panicf("Cannot open logfile %v", err)
	} else {
		// DB path flag parse
		log.Infof("Set dbpath: %s", locater.Dbpath)
	}

	// Directory check
	if _, err := os.Stat(LOCATEDIR); os.IsNotExist(err) {
		log.Panic(err) // /var/lib/mlocateがなければ終了
	}

	// Command check
	// スペース区切りされたconstをexec.LookPath()で実行可能ファイルであるかを調べる
	for _, r := range strings.Fields(REQUIRE) {
		if _, err := exec.LookPath(r); err != nil {
			log.Panicf("%s", err.Error())
		}
	}

	// Open server
	route := gin.Default()
	route.Static("/static", "./static")
	route.LoadHTMLGlob("templates/*")

	// Top page
	route.GET("/", func(c *gin.Context) {
		datalist, err := cmd.Datalist(LOGFILE)
		if err != nil {
			log.Error(err)
		}
		c.HTML(http.StatusOK,
			"index.tmpl",
			gin.H{
				"title":          "",
				"lastUpdateTime": locater.Stats.LastUpdateTime,
				"datalist":       datalist,
				"query":          "",
			})
	})

	// Result view
	route.GET("/search", func(c *gin.Context) {
		// HTML
		query := c.Request.URL.Query()
		q := strings.Join(query["q"], " ")
		// JSON
		// 検索文字数チェックOK
		/* LocateStats()の結果が前と異なっていたら
		locateS更新
		cacheを初期化 */
		if l, err := cmd.LocateStats(locater.Dbpath); string(l) != string(locateS) {
			// DB更新されていたら
			if err != nil {
				log.Error(err)
			}
			locateS = l // 保持するDB情報の更新
			// Initialize cache
			// nil map assignment errorを発生させないために必要
			cache = cmd.CacheMap{} // Reset cache
			// Count number of search target files
			var n uint64
			n, err = cmd.LocateStatsSum(locateS)
			if err != nil {
				log.Error(err)
			}
			locater.Stats.Items = cmd.Ambiguous(n)
			// Update LastUpdateTime for database
			locater.Stats.LastUpdateTime = cmd.DBLastUpdateTime(locater.Dbpath)
		}
		// Response
		c.HTML(http.StatusOK,
			"index.tmpl",
			gin.H{
				"title":          q,
				"lastUpdateTime": locater.Stats.LastUpdateTime,
				"query":          q,
			})
	})

	// API
	route.GET("/history", func(c *gin.Context) {
		searchHistory, err := cmd.Datalist(LOGFILE)
		if err != nil {
			log.Error(err)
			c.JSON(404, searchHistory)
		}
		if locater.Args.Debug {
			log.Debug(searchHistory)
		}
		c.JSON(http.StatusOK, searchHistory)
	})

	route.GET("/json", func(c *gin.Context) {
		// locater.Query initialize
		local := locater // Shallow copy
		local.Query.Logging = true
		local.Query.Limit = 0

		// Parse query
		q := c.Query("q")
		sw, ew, err := cmd.QueryParser(q)
		if err != nil {
			log.Errorf("%s [ %-50s ]", err, q)
			local.Stats.Response = 404
			c.JSON(local.Stats.Response, local)
			return
		}
		local.Query.SearchWords, local.Query.ExcludeWords = sw, ew
		lm := c.Query("limit")
		if lm != "" {
			ln, err := strconv.Atoi(lm)
			if err != nil {
				log.Errorf("Error in 'limit' API: %s", err)
				local.Stats.Response = 404
				c.JSON(local.Stats.Response, local)
				return
			}
			if ln > 0 { // 0未満のときは無視(initialize で既に0になっている)
				local.Query.Limit = ln
			}
		}

		// Execute locate command
		start := time.Now()
		result, ok, err := cache.Traverse(&local)
		if local.Args.Debug {
			log.Debugf("gocate result %v", result)
		}
		end := (time.Since(start)).Nanoseconds()
		local.Stats.SearchTime = float64(end) / float64(time.Millisecond)

		// Response & Logging
		if err != nil {
			local.Stats.Response = 404
			log.Errorf("%s [ %-50s ]", err, q)
			c.JSON(local.Stats.Response, local)
		} else {
			getpushLog := "PUSH result to cache"
			if ok {
				getpushLog = "GET result from cache"
			}
			local.Paths = result
			local.Stats.Response = http.StatusOK
			l := []interface{}{len(local.Paths), local.Stats.SearchTime, getpushLog, q}
			// 基本的にすべての検索はログに記録する
			// http:...&logging=falseのときだけ記録しない
			if c.Query("logging") == "false" {
				local.Query.Logging = false
			}
			if local.Query.Logging {
				log.Noticef("%8dfiles %3.3fmsec %s [ %-50s ]", l...)
			} else {
				fmt.Printf("[NO LOGGING NOTICE]\t%8dfiles %3.3fmsec %s [ %-50s ]\n", l...) // Printfで表示はする
			}
			c.JSON(http.StatusOK, local)
		}
	})

	route.GET("/status", func(c *gin.Context) {
		l, err := cmd.LocateStats(locater.Args.Dbpath)
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
