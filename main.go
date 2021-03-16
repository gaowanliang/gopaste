package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
)

// Configuration read from config.json

var r *gin.Engine

func saveHand(c *gin.Context) {
	paste := c.PostForm("p")
	expiry := c.PostForm("expiry")
	lang := c.PostForm("lang")
	if paste == "" { // Empty paste, go back
		c.Request.URL.Path = "/"
		r.HandleContext(c)
		return
	}
	b, err := Save(paste, expiry, lang)
	if err != nil {
		switch err.(type) {
		case *InvalidPasteError:
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Not Found"})
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error"})
		}
		return
	}
	//c.Request.URL.Path = b.URL
	//r.HandleContext(c)
	c.Redirect(http.StatusMovedPermanently, b.URL)
}
func pasteHand(c *gin.Context) {
	paste := c.Param("id")
	s, lang, err := GetPaste(paste)
	if err != nil {
		switch err.(type) {
		case *InvalidPasteError:
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Not Found"})
		case *InvalidURLError:
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Invalid URL"})
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error"})
		}
		return
	}
	if lang == "url" {
		value, ok := c.Request.Header["Referer"]
		if ok {
			hostURL, _ := url.Parse(value[0])
			hostname := hostURL.Hostname()
			if hostURL.Port() != "" {
				hostname += ":" + hostURL.Port()
			}
			//log.Println(hostname, c.Request.Host)
			if hostname != c.Request.Host {
				c.Redirect(http.StatusMovedPermanently, s)
			}
		} else {
			c.Redirect(http.StatusMovedPermanently, s)
		}
	}

	link := "/" + paste + "/raw"
	download := "/" + paste + "/download"
	clone := "/" + paste + "/clone"
	// Page struct

	c.HTML(http.StatusOK, "paste.html", gin.H{
		"Title":    paste,
		"Body":     []byte(s),
		"Raw":      link,
		"Home":     configuration.Address,
		"Download": download,
		"Clone":    clone,
		"lang":     lang,
	})
}

func main() {
	// 创建一个默认的路由引擎
	LoadConfiguration()
	CheckDB()

	r = gin.Default()
	//r.Static("assets/static", "./assets/static")
	r.LoadHTMLGlob("assets/html/*")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"Title": "goPaste",
			"Body":  []byte(""),
		})
	})
	r.POST("/", saveHand)
	r.GET("/:id", pasteHand)
	// 启动HTTP服务，默认在0.0.0.0:8080启动服务
	r.Run(configuration.Port)
}
