package api

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

func HistoryQueryParser(c *gin.Context, q string) (int, error) {
	s, ok := c.GetQuery(q)
	if !ok {
		return 0, errors.New("error: no query " + q)
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return n, nil
}
