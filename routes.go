package main

import (
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/karlseguin/ccache"
)

type IpResponse struct {
	City    string `json:"city"`
	Country string `json:"country"`
	Area    string `json:"area"`
}

// HanlderRoutes 加载发访问路由
func HanlderRoutes(baseURI string) http.Handler {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	router := r.Group(baseURI)
	// 获取ip信息
	ccConfigure := ccache.Configure()
	ccConfigure.MaxSize(6000)
	ccCache := ccache.New(ccConfigure)
	router.GET("/location/:ip", func(c *gin.Context) {
		cc := ccCache.Get(c.Param("ip"))
		if cc != nil {
			defer cc.Release()
			if cc.Expired() == false {
				r := cc.Value().(IpResponse)
				c.JSON(200, r)
				return
			}
		}
		r := IPDB.Find(c.Param("ip"))
		if r.Area == "" && r.Country == "" {
			c.JSON(404, &gin.H{
				"message": "ip不存在",
			})
			return
		}
		matched, _ := regexp.MatchString(RegString, r.Country)
		country := r.Country
		if matched == true {
			country = "中国"
		} else {
			matched, _ := regexp.MatchString(r.Country, RegString)
			if matched == true {
				country = "中国"
			} else if r.Country != "局域网" && r.Country != "IANA" {
				country = "外国"
			}
		}
		resp := IpResponse{
			Area:    r.Area,
			Country: country,
			City:    r.Country,
		}
		ccCache.Set(c.Param("ip"), resp, 30*time.Second)
		c.JSON(200, resp)
	})

	r.NoRoute(func(c *gin.Context) {
		c.String(404, `
         \\\///
        / _  _ \
      (| (.)(.) |)
.---.OOOo--()--oOOO.---.
|                      |
|     404 NOT FOUND    |
|                      |
'---.oooO--------------'
    (   )   Oooo.
     \ (    (   )
      \_)    ) /
            (_/
	`)
	})
	return r
}
