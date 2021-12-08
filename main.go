package main

import (
	"flag"
	"fmt"
	"math"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	api "github.com/u1and0/locate-server/cmd/api"
	cmd "github.com/u1and0/locate-server/cmd/locater"

	"github.com/gin-gonic/gin"
	"github.com/op/go-logging"
)

const (
	// VERSION : version
	VERSION = "3.0.0r"
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
	port        int
)

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
		if err != nil {
			log.Error(err)
		}
		c.HTML(http.StatusOK,
			"index.tmpl",
			gin.H{
				"title":          "",
				"lastUpdateTime": locater.Stats.LastUpdateTime,
				"query":          "",
			})
	})

	// Result view
	route.GET("/search", func(c *gin.Context) {
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
		// query := c.Request.URL.Query()
		// q := strings.Join(query["q"], " ")
		q := c.Query("q")
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
		history, err := cmd.Datalist(LOGFILE)
		if err != nil {
			log.Error(err)
			c.JSON(404, history)
		}
		gt := api.IntQuery(c, "gt") // history?gt=10 => gt==10
		lt := api.IntQuery(c, "lt") // history?lt=100 => lt==100
		// lt default value is infinity
		if lt == 0 {
			lt = math.MaxInt64
		}
		// if designated query gt & lt then filter the history
		// No query gt nor lt then nothing to do
		if gt != 0 || lt != math.MaxInt64 {
			history = history.Filter(gt, lt)
		}
		c.JSON(http.StatusOK, history)
	})

	route.GET("/json", func(c *gin.Context) {
		// locater.Query initialize
		// Shallow copy locater to local
		// for blocking to rewrite
		// locater{} struct while searching
		local := locater

		// Parse query
		var queryDefault api.Query
		query := queryDefault.New()
		if err := c.ShouldBind(&query); err != nil {
			log.Errorf("error: %s query: %v", err, query)
			local.Stats.Response = 404
			c.JSON(local.Stats.Response, local)
			return
		}
		sw, ew, err := api.QueryParser(query.Q)
		if err != nil {
			log.Errorf("error %v", err)
		}
		local.SearchWords = sw
		local.ExcludeWords = ew
		local.Query.Q = query.Q
		local.Query.Logging = query.Logging
		local.Query.Limit = query.Limit

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
			log.Errorf("%s [ %-50s ]", err, query.Q)
			c.JSON(local.Stats.Response, local)
		} else {
			getpushLog := "PUSH result to cache"
			if ok {
				getpushLog = "GET result from cache"
			}
			local.Paths = result
			local.Stats.Response = http.StatusOK
			l := []interface{}{len(local.Paths), local.Stats.SearchTime, getpushLog, query.Q}
			if query.Logging {
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
	route.Run(":" + strconv.Itoa(port)) // => :8080
}

// Parse command line option
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
	flag.IntVar(&port, "p", 8080, "Server port number. Default access to http://localhost:8080/")
	flag.IntVar(&port, "port", 8080, "Server port number. Default access to http://localhost:8080/")
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
