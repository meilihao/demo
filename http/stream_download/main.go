package main

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/meilihao/golib/v2/cmd"
)

func main() {
	r := gin.Default()
	r.GET("/file", func(c *gin.Context) {

		c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", "demo.zip"))
		c.Writer.Header().Set("Content-Type", "application/octet-stream")
		c.Writer.Header().Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))

		ctl, err := cmd.CmdStdoutStreamWithBash("zip -9r - /home/chen/test/zstack-repos") // test 2g: use 4m
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		defer ctl.Close()

		io.Copy(c.Writer, ctl.StdoutReader)
	})
	r.Run(":9090")
}
