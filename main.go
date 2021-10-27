package main

import (
	"net/http"
	"strings"

	cmd "github.com/u1and0/locate-server/cmd"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	r.GET("/index", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title":          "Locate Server",
			"explain":        "説明",
			"LastUpdateTime": "2016-01-02 15:04:05",
			"datalist":       "datalist",
		})
	})

	type Paths []string
	type Result struct {
		Paths  `json:"paths"`
		Status int    `json:"status"`
		Err    error  `json:"error"`
		Query  string `json:"query"`
	}
	var err error
	result := Result{}
	result.Paths = Paths{
		"path/to/1",
		"path/to/2",
		"path/to/3",
	}
	result.Status = http.StatusOK
	result.Err = err

	r.GET("/search", func(c *gin.Context) {
		q := c.Query("query")
		l := cmd.Locater{
			SearchWords: strings.Fields(q),
			Dbpath:      "/var/lib/mlocate",
			Debug:       false,
		}
		path, err := l.Locate()
		result.Paths = path
		if err != nil {
			panic(err)
		}
		result.Query = q
		c.JSON(http.StatusOK, result)
	})

	r.GET("/moreJSON", func(c *gin.Context) {
		// gin.H is a shortcut for map[string]interface{}
		c.JSON(http.StatusOK, gin.H{
			"results": Paths{
				"path/to/4",
				"path/to/5",
				"path/to/6",
			},
			"status": http.StatusOK,
			"error":  err,
		})
	})

	// Listen and serve on 0.0.0.0:8080
	r.Run(":8080")
}
