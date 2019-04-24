package http

import (
	"io/ioutil"
	"log"
	"net/http"

	raven "github.com/getsentry/raven-go"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sentry"
	"github.com/gin-gonic/gin"
	"github.com/marknguyen85/server-api/fetcher"
	persister "github.com/marknguyen85/server-api/persister"
)

//HTTPServer struct
type HTTPServer struct {
	fetcher   *fetcher.Fetcher
	persister persister.Persister
	host      string
	r         *gin.Engine
}

//GetRate func
func (httpServer *HTTPServer) GetRate(c *gin.Context) {
	isNewRate := httpServer.persister.GetIsNewRate()
	if isNewRate != true {
		c.JSON(
			http.StatusOK,
			gin.H{"success": false, "data": nil},
		)
		return
	}

	rates := httpServer.persister.GetRate()
	updateAt := httpServer.persister.GetTimeUpdateRate()
	c.JSON(
		http.StatusOK,
		gin.H{"success": true, "updateAt": updateAt, "data": rates},
	)
}

//GetRateUSD func
func (httpServer *HTTPServer) GetRateUSD(c *gin.Context) {
	if !httpServer.persister.GetIsNewRateUSD() {
		c.JSON(
			http.StatusOK,
			gin.H{"success": false},
		)
		return
	}

	rates := httpServer.persister.GetRateUSD()
	c.JSON(
		http.StatusOK,
		gin.H{"success": true, "data": rates},
	)
}

//GetRateTOMO func
func (httpServer *HTTPServer) GetRateTOMO(c *gin.Context) {
	if !httpServer.persister.GetIsNewRateUSD() {
		c.JSON(
			http.StatusOK,
			gin.H{"success": false},
		)
		return
	}

	tomoRate := httpServer.persister.GetRateTOMO()
	c.JSON(
		http.StatusOK,
		gin.H{"success": true, "data": tomoRate},
	)
}

//GetErrorLog func
func (httpServer *HTTPServer) GetErrorLog(c *gin.Context) {
	dat, err := ioutil.ReadFile("error.log")
	if err != nil {
		log.Print(err)
		c.JSON(
			http.StatusOK,
			gin.H{"success": false, "data": err},
		)
	}
	c.JSON(
		http.StatusOK,
		gin.H{"success": true, "data": string(dat[:])},
	)
}

//GetLast7D func
func (httpServer *HTTPServer) GetLast7D(c *gin.Context) {
	listTokens := c.Query("listToken")
	data := httpServer.persister.GetLast7D(listTokens)
	if httpServer.persister.GetIsNewTrackerData() {
		c.JSON(
			http.StatusOK,
			gin.H{"success": true, "data": data, "status": "latest"},
		)
		return
	}
	c.JSON(
		http.StatusOK,
		gin.H{"success": true, "data": data, "status": "old"},
	)
}

//getCacheVersion func
func (httpServer *HTTPServer) getCacheVersion(c *gin.Context) {
	timeRun := httpServer.persister.GetTimeVersion()
	c.JSON(
		http.StatusOK,
		gin.H{"success": true, "data": timeRun},
	)
}

//Run func
func (httpServer *HTTPServer) Run(chainTexENV string) {
	httpServer.r.GET("/getRate", httpServer.GetRate)
	httpServer.r.GET("/rate", httpServer.GetRate)

	httpServer.r.GET("/getRateUSD", httpServer.GetRateUSD)
	httpServer.r.GET("/rateUSD", httpServer.GetRateUSD)

	httpServer.r.GET("/getLast7D", httpServer.GetLast7D)
	httpServer.r.GET("/last7D", httpServer.GetLast7D)

	httpServer.r.GET("/getRateTOMO", httpServer.GetRateTOMO)
	httpServer.r.GET("/rateTOMO", httpServer.GetRateTOMO)

	httpServer.r.GET("/cacheVersion", httpServer.getCacheVersion)

	if chainTexENV != "production" {
		httpServer.r.GET("/9d74529bc6c25401a2f984ccc9b0b2b3", httpServer.GetErrorLog)
	}

	httpServer.r.Run(httpServer.host)
}

//NewHTTPServer contruct
func NewHTTPServer(host string, persister persister.Persister, fetcher *fetcher.Fetcher) *HTTPServer {
	r := gin.Default()
	r.Use(sentry.Recovery(raven.DefaultClient, false))
	r.Use(cors.Default())

	return &HTTPServer{
		fetcher, persister, host, r,
	}
}
