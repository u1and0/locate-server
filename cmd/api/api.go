package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// IntQuery parse query as int
func IntQuery(c *gin.Context, q string) int {
	s, ok := c.GetQuery(q)
	if !ok {
		return 0
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}
