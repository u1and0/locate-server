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
	cache "github.com/u1and0/locate-server/cmd/cache"
	cmd "github.com/u1and0/locate-server/cmd/locater"

	"github.com/gin-gonic/gin"
	"github.com/op/go-logging"
)

const (
	// VERSION : version
	VERSION = "3.1.0"
	// LOGFILE : 検索条件 / 検索結果 / 検索時間を記録するファイル
	LOGFILE = "/var/lib/mlocate/locate.log"
	// LOCATEDIR : locate (gocate) search db path
	LOCATEDIR = "/var/lib/mlocate"
	// REQUIRE : required commands. Separate by space.
	REQUIRE = "locate gocate"
	// PORT : default open server port
	PORT = 8080
)

var (
	log         = logging.MustGetLogger("main")
	showVersion bool
	port        int
	locater     = parseCmdlineOption()
	locateS     []byte
	caches      = cache.New()
)

type (
	usageText struct {
		dir,
		port,
		root,
		windowsPathSeparate,
		trim,
		debug,
		showVersion string
	}
)

func main() {
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
	route.GET("/", topPage)

	// Result view
	route.GET("/search", searchPage)

	// API
	route.GET("/history", fetchHistory)
	route.GET("/json", fetchJSON)
	route.GET("/status", fetchStatus)

	// Listen and serve on 0.0.0.0:8080
	route.Run(":" + strconv.Itoa(port)) // => :8080
}

func topPage(c *gin.Context) {
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"title":          "",
		"lastUpdateTime": locater.Stats.LastUpdateTime,
		"query":          "",
	})
}

func searchPage(c *gin.Context) {
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
		caches = cache.New() // Reset cache
		// Count number of search target files
		var n int64
		n, err = cmd.LocateStatsSum(locateS)
		if err != nil {
			log.Error(err)
		}
		locater.Stats.Items = cmd.Ambiguous(n)
		// Update LastUpdateTime for database
		locater.Stats.LastUpdateTime = cmd.DBLastUpdateTime(locater.Dbpath)
	}
	// Response
	q := c.Query("q")
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"title":          q,
		"lastUpdateTime": locater.Stats.LastUpdateTime,
		"query":          q,
	})
}

func fetchJSON(c *gin.Context) {
	// locater.Query initialize
	// Shallow copy locater to local
	// for blocking to rewrite
	// locater{} struct while searching
	local := locater

	// Parse query
	query, err := api.New(c)
	local.Query = api.Query{
		Q:       query.Q,
		Logging: query.Logging,
		Limit:   query.Limit,
	}
	if err != nil {
		log.Errorf("error: %s query: %#v", err, query)
		local.Error = fmt.Sprintf("%s", err)
		c.JSON(406, local)
		// 406 Not Acceptable:
		// サーバ側が受付不可能な値であり提供できない状態
		return
	}

	local.SearchWords, local.ExcludeWords, err = api.QueryParser(query.Q)
	if local.Args.Debug {
		log.Debugf("local locater: %#v", local)
	}
	if err != nil {
		log.Errorf("error %v", err)
		local.Error = fmt.Sprintf("%v", err)
		c.JSON(406, local)
		// 406 Not Acceptable:
		// サーバ側が受付不可能な値であり提供できない状態
		return
	}

	// Execute locate command
	start := time.Now()
	result, ok, err := caches.Traverse(&local) // err <- OS command error
	if local.Args.Debug {
		log.Debugf("gocate result %v", result)
	}
	end := (time.Since(start)).Nanoseconds()
	local.Stats.SearchTime = float64(end) / float64(time.Millisecond)

	// Response & Logging
	if err != nil {
		log.Errorf("%s [ %-50s ]", err, query.Q)
		c.JSON(500, local)
		// 500 Internal Server Error
		// 何らかのサーバ内で起きたエラー
		return
	}
	local.Paths = result
	getpushLog := "PUSH result to cache"
	if ok {
		getpushLog = "GET result from cache"
	}
	if !query.Logging {
		getpushLog = "NO LOGGING result"
	}
	l := []interface{}{len(local.Paths), local.Stats.SearchTime, getpushLog, query.Q}
	log.Noticef("%8dfiles %3.3fmsec %s [ %-50s ]", l...)
	if len(local.Paths) == 0 {
		local.Error = "no content"
		c.JSON(204, local)
		// 204 No Content
		// リクエストに対して送信するコンテンツは無いが
		// ヘッダは有用である
		return
	}
	c.JSON(http.StatusOK, local)
	// 200 OK
	// リクエストが正常に処理できた
}

func fetchHistory(c *gin.Context) {
	history, err := cmd.Datalist(LOGFILE)
	if err != nil {
		log.Error(err)
		c.JSON(404, history)
		return
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
}

func fetchStatus(c *gin.Context) {
	l, err := cmd.LocateStats(locater.Args.Dbpath) // err <- OS command error
	ss := strings.Split(string(l), "\n")
	if err != nil {
		log.Errorf("error: %s", err)
		c.JSON(500, gin.H{
			"locate-S": ss,
			"error":    err,
		})
		// 500 Internal Server Error
		// 何らかのサーバ内で起きたエラー
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"locate-S": ss,
		"error":    err,
	})
}

// Parse command line option
func parseCmdlineOption() (l cmd.Locater) {
	var (
		showVersion bool
		usage       = usageText{
			dir:                 `Path of locate database directory (default "/var/lib/mlocate")`,
			port:                `Server port number. Default access to http://localhost:8080/ (default 8080)`,
			root:                `DB insert prefix for directory path`,
			windowsPathSeparate: `Use path separate Windows backslash`,
			trim:                `DB trim prefix for directory path`,
			debug:               `Run debug mode`,
			showVersion:         `Show version`,
		}
	)
	flag.StringVar(&l.Args.Dbpath, "d", LOCATEDIR, usage.dir)
	flag.StringVar(&l.Args.Dbpath, "dir", LOCATEDIR, usage.dir)
	flag.BoolVar(&l.Args.PathSplitWin, "s", false, usage.windowsPathSeparate)
	flag.BoolVar(&l.Args.PathSplitWin, "windows-path-separate", false, usage.windowsPathSeparate)
	flag.StringVar(&l.Args.Root, "r", "", usage.root)
	flag.StringVar(&l.Args.Root, "root", "", usage.root)
	flag.StringVar(&l.Args.Trim, "t", "", usage.trim)
	flag.StringVar(&l.Args.Trim, "trim", "", usage.trim)
	flag.BoolVar(&l.Args.Debug, "debug", false, usage.debug)
	flag.IntVar(&port, "p", PORT, usage.port)
	flag.IntVar(&port, "port", PORT, usage.port)
	flag.BoolVar(&showVersion, "v", false, usage.showVersion)
	flag.BoolVar(&showVersion, "version", false, usage.showVersion)
	flag.Usage = func() {
		usageTxt := fmt.Sprintf(`Open file search server

Usage of locate-server
	locate-server [OPTION]...
-d, -dir
	%s
-p, -port
	%s
-r, -root
	%s
-s, -windows-path-separate
	%s
-t, -trim
	%s
-debug
	%s
-v, -version
	%s`,
			usage.dir,
			usage.port,
			usage.root,
			usage.windowsPathSeparate,
			usage.trim,
			usage.debug,
			usage.showVersion,
		)
		fmt.Fprintf(os.Stderr, "%s\n", usageTxt)
	}
	flag.Parse()
	if showVersion {
		fmt.Println("locate-server version", VERSION)
		os.Exit(0) // Exit with version info
	}
	return l
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
