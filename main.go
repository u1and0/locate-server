package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	api "github.com/u1and0/locate-server/cmd/api"
	cache "github.com/u1and0/locate-server/cmd/cache"
	cmd "github.com/u1and0/locate-server/cmd/locater"

	"github.com/gin-gonic/gin"
)

const (
	// VERSION : version
	VERSION = "4.0.0"
	// LOGFILE : 検索条件 / 検索結果 / 検索時間を記録するファイル
	LOGFILE = "/var/lib/plocate/locate.log"
	// LOCATEDIR : locate (gocate) search db path
	LOCATEDIR = "/var/lib/plocate"
	// PORT : default open server port
	PORT = 8080
)

type (
	usageText struct {
		dir,
		port,
		root,
		windowsPathSeparate,
		trim,
		debug,
		locateCmd,
		showVersion string
	}

	// App holds server state
	App struct {
		locater cmd.Locater
		caches  *cache.Map
		mu      sync.RWMutex
		locateS int64
		port    int
	}
)

// NewApp creates a new App instance
func NewApp(locater cmd.Locater, port int) *App {
	return &App{
		locater: locater,
		caches:  cache.New(),
		port:    port,
	}
}

func main() {
	locater, port := parseCmdlineOption()

	if !locater.Args.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	logfile, err := os.OpenFile(LOGFILE, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		slog.Error("Cannot open logfile", "err", err)
		panic("Cannot open logfile")
	}
	defer logfile.Close()
	setLogger(logfile)

	slog.Info("Set dbpath", "path", locater.Args.Dbpath)
	slog.Info("locate command", "cmd", locater.Args.LocateCmd)

	app := NewApp(locater, port)
	if err := app.run(); err != nil {
		slog.Error("shutdown error", "err", err)
		os.Exit(1)
	}
}

func (a *App) run() error {
	route := gin.Default()
	route.Static("/static", "./static")
	route.LoadHTMLGlob("templates/*")

	route.GET("/", a.topPage)
	route.GET("/search", a.searchPage)
	route.GET("/history", a.fetchHistory)
	route.GET("/json", a.fetchJSON)
	route.GET("/status", a.fetchStatus)

	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(a.port),
		Handler: route,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("server starting", "port", a.port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("ListenAndServe", "err", err)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

func (a *App) topPage(c *gin.Context) {
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"title":          "",
		"lastUpdateTime": a.locater.Stats.LastUpdateTime,
		"query":          "",
	})
}

func (a *App) searchPage(c *gin.Context) {
	/* LocateStats()の結果が前と異なっていたら
	locateS更新
	cacheを初期化 */
	if l, err := cmd.DBSize(a.locater.Dbpath); l != a.locateS {
		a.mu.Lock()
		a.locateS = l
		a.caches = cache.New()
		a.locater.Stats.LastUpdateTime = cmd.DBLastUpdateTime(a.locater.Dbpath)
		a.mu.Unlock()
		if err != nil {
			slog.Error("DBSize", "err", err)
		}
	}
	q := c.Query("q")
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"title":          q,
		"lastUpdateTime": a.locater.Stats.LastUpdateTime,
		"query":          q,
	})
}

func (a *App) fetchJSON(c *gin.Context) {
	// Shallow copy locater to local for blocking rewrite while searching
	local := a.locater

	// Parse query
	query, err := api.New(c)
	local.Query = api.Query{
		Q:       query.Q,
		Logging: query.Logging,
		Limit:   query.Limit,
	}
	if err != nil {
		slog.Error("query parse error", "err", err, "query", fmt.Sprintf("%#v", query))
		local.Error = fmt.Sprintf("%s", err)
		c.JSON(406, local)
		return
	}

	local.SearchWords, local.ExcludeWords, err = api.QueryParser(query.Q)
	if local.Args.Debug {
		slog.Debug("local locater", "locater", fmt.Sprintf("%#v", local))
	}
	if err != nil {
		slog.Error("QueryParser error", "err", err)
		local.Error = fmt.Sprintf("%v", err)
		c.JSON(406, local)
		return
	}

	// Execute locate command
	start := time.Now()
	a.mu.Lock()
	result, ok, err := a.caches.Traverse(&local)
	a.mu.Unlock()
	if local.Args.Debug {
		slog.Debug("locate result", "result", result)
	}
	end := (time.Since(start)).Nanoseconds()
	local.Stats.SearchTime = float64(end) / float64(time.Millisecond)

	// Response & Logging
	if err != nil {
		slog.Error("locate command error", "err", err, "query", query.Q)
		c.JSON(500, local)
		return
	}
	local.Paths = result
	cacheLog := "PUSH"
	if ok {
		cacheLog = "GET"
	}
	if !query.Logging {
		cacheLog = "NO LOGGING"
	}
	slog.Info("search",
		"files", len(local.Paths),
		"msec", local.Stats.SearchTime,
		"cache", cacheLog,
		"query", query.Q,
	)
	if len(local.Paths) == 0 {
		err = errors.New("no content")
		local.Error = fmt.Sprintf("%v", err)
		c.JSON(200, local)
		return
	}
	c.JSON(http.StatusOK, local)
}

func (a *App) fetchHistory(c *gin.Context) {
	history, err := cmd.Datalist(LOGFILE)
	if err != nil {
		slog.Error("fetchHistory", "err", err)
		c.JSON(404, history)
		return
	}
	gt := api.IntQuery(c, "gt")
	lt := api.IntQuery(c, "lt")
	if lt == 0 {
		lt = math.MaxInt64
	}
	if gt != 0 || lt != math.MaxInt64 {
		history = history.Filter(gt, lt)
	}
	c.JSON(http.StatusOK, history)
}

func (a *App) fetchStatus(c *gin.Context) {
	l, err := cmd.DBSize(a.locater.Args.Dbpath)
	if err != nil {
		slog.Error("DBSize", "err", err)
		c.JSON(500, gin.H{
			"locate-S": l,
			"error":    err,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"locate-S": l,
		"error":    err,
	})
}

// parseCmdlineOption parses command line flags and returns Locater and port
func parseCmdlineOption() (l cmd.Locater, port int) {
	var (
		showVersion bool
		usage       = usageText{
			dir:                 `Path of locate database directory (default "/var/lib/plocate")`,
			port:                `Server port number. Default access to http://localhost:8080/ (default 8080)`,
			root:                `DB insert prefix for directory path`,
			windowsPathSeparate: `Use path separate Windows backslash`,
			trim:                `DB trim prefix for directory path`,
			debug:               `Run debug mode`,
			locateCmd:           `locate command path (default: auto-detect gocate or locate)`,
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
	flag.StringVar(&l.Args.LocateCmd, "locate-cmd", "", usage.locateCmd)
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
-locate-cmd
	%s
-v, -version
	%s`,
			usage.dir,
			usage.port,
			usage.root,
			usage.windowsPathSeparate,
			usage.trim,
			usage.debug,
			usage.locateCmd,
			usage.showVersion,
		)
		fmt.Fprintf(os.Stderr, "%s\n", usageTxt)
	}
	flag.Parse()
	if showVersion {
		fmt.Println("locate-server version", VERSION)
		os.Exit(0)
	}

	// 環境変数フォールバック
	if l.Args.LocateCmd == "" {
		if envCmd := os.Getenv("LOCATE_CMD"); envCmd != "" {
			l.Args.LocateCmd = envCmd
		}
	}
	resolved, err := cmd.ResolveLocateCmd(l.Args.LocateCmd)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	l.Args.LocateCmd = resolved
	return
}

// setLogger is printing out log message to STDOUT and LOGFILE
func setLogger(f *os.File) {
	mw := io.MultiWriter(f, os.Stdout)
	h := slog.NewTextHandler(mw, &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(h))
}
