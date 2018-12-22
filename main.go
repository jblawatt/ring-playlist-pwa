package main

import (
	"crypto/sha1"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/gin-gonic/gin"
)

type Item struct {
	No     int    `json:"n"`
	Artist string `json:"a"`
	Title  string `json:"tt"`
	Time   string `json:"t"`
	Raw    string `json:"r"`
	Hash   string `json:"h"`
}

func main() {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./static")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})

	r.GET("/api/playlist", func(c *gin.Context) {
		doc, _ := htmlquery.LoadURL("http://www.konstantinkuehn.com/playlist/")
		list := htmlquery.Find(doc, "//body/blockquote/ol/li")
		r := regexp.MustCompile(`(.+) *- *(.+) \((\d{1,3}\:\d{2})\)`)
		var resp []Item
		for i, el := range list {
			text := htmlquery.InnerText(el)
			parsed := r.FindStringSubmatch(text)
			if len(parsed) < 4 {
				parsed = []string{"unbekannt", "unbekannt", "unbekannt", "??:??"}
			}
			hash := sha1.New()
			io.WriteString(hash, strconv.Itoa(i+1))
			io.WriteString(hash, strings.Trim(parsed[1], " "))
			io.WriteString(hash, strings.Trim(parsed[2], " "))
			io.WriteString(hash, strings.Trim(parsed[3], " "))
			io.WriteString(hash, strings.Trim(parsed[0], " "))
			item := Item{
				No:     i + 1,
				Artist: strings.Trim(parsed[1], " "),
				Title:  strings.Trim(parsed[2], " "),
				Time:   strings.Trim(parsed[3], " "),
				Raw:    strings.Trim(parsed[0], " "),
				Hash:   fmt.Sprintf("%x", hash.Sum(nil)),
			}
			resp = append(resp, item)
		}
		c.JSON(http.StatusOK, gin.H{"data": resp})
	})
	r.Run(fmt.Sprintf("127.0.0.1:%s", os.Getenv("RING_PWA_PORT")))
}
