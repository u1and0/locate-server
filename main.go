package main

import (
	"net/http"

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
		Status int   `json:"status"`
		Err    error `json:"error"`
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
