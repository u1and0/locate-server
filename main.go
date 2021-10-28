package main

import (
	"net/http"
	"strings"

	cmd "github.com/u1and0/locate-server/cmd"

	"github.com/gin-gonic/gin"
)

// Paths locate command result
type Paths []string

// Result return JSON struct
type Result struct {
	Paths  `json:"paths"`
	Status int    `json:"status"`
	Err    error  `json:"error"`
	Query  string `json:"query"`
}

func main() {
	var (
		q     string
		route = gin.Default()
	)
	route.Static("/static", "./static")
	route.LoadHTMLGlob("templates/*")

	route.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK,
			"index.tmpl",
			gin.H{
				"title":          q,
				"explain":        "説明",
				"lastUpdateTime": "2016-01-02 15:04:05",
				"datalist":       "datalist",
				"query":          q,
			})
	})

	route.GET("/search", func(c *gin.Context) {
		// result, _ := ResultPath(c)
		c.HTML(http.StatusOK,
			"index.tmpl",
			gin.H{
				"title":          q,
				"explain":        "説明",
				"lastUpdateTime": "2016-01-02 15:04:05",
				"datalist":       "datalist",
				"query":          q,
				// "paths":          result.Paths,
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
