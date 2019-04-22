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

type HTTPServer struct {
	fetcher   *fetcher.Fetcher
	persister persister.Persister
	host      string
	r         *gin.Engine
}

func (self *HTTPServer) GetRate(c *gin.Context) {
	isNewRate := self.persister.GetIsNewRate()
	if isNewRate != true {
		c.JSON(
			http.StatusOK,
			gin.H{"success": false, "data": nil},
		)
		return
	}

	rates := self.persister.GetRate()
	updateAt := self.persister.GetTimeUpdateRate()
	c.JSON(
		http.StatusOK,
		gin.H{"success": true, "updateAt": updateAt, "data": rates},
	)
}

func (self *HTTPServer) GetRateUSD(c *gin.Context) {
	if !self.persister.GetIsNewRateUSD() {
		c.JSON(
			http.StatusOK,
			gin.H{"success": false},
		)
		return
	}

	rates := self.persister.GetRateUSD()
	c.JSON(
		http.StatusOK,
		gin.H{"success": true, "data": rates},
	)
}

func (self *HTTPServer) GetRateTOMO(c *gin.Context) {
	if !self.persister.GetIsNewRateUSD() {
		c.JSON(
			http.StatusOK,
			gin.H{"success": false},
		)
		return
	}

	tomoRate := self.persister.GetRateTOMO()
	c.JSON(
		http.StatusOK,
		gin.H{"success": true, "data": tomoRate},
	)
}

func (self *HTTPServer) GetErrorLog(c *gin.Context) {
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

func (self *HTTPServer) GetLast7D(c *gin.Context) {
	listTokens := c.Query("listToken")
	data := self.persister.GetLast7D(listTokens)
	if self.persister.GetIsNewTrackerData() {
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

func (self *HTTPServer) getCacheVersion(c *gin.Context) {
	timeRun := self.persister.GetTimeVersion()
	c.JSON(
		http.StatusOK,
		gin.H{"success": true, "data": timeRun},
	)
}

func (self *HTTPServer) Run(chainTexENV string) {
	self.r.GET("/getRateUSD", self.GetRateUSD)
	self.r.GET("/rateUSD", self.GetRateUSD)

	self.r.GET("/getLast7D", self.GetLast7D)
	self.r.GET("/last7D", self.GetLast7D)

	self.r.GET("/getRateTOMO", self.GetRateTOMO)
	self.r.GET("/rateTOMO", self.GetRateTOMO)

	self.r.GET("/cacheVersion", self.getCacheVersion)

	if chainTexENV != "production" {
		self.r.GET("/9d74529bc6c25401a2f984ccc9b0b2b3", self.GetErrorLog)
	}

	self.r.Run(self.host)
}

func NewHTTPServer(host string, persister persister.Persister, fetcher *fetcher.Fetcher) *HTTPServer {
	r := gin.Default()
	r.Use(sentry.Recovery(raven.DefaultClient, false))
	r.Use(cors.Default())

	return &HTTPServer{
		fetcher, persister, host, r,
	}
}
