package main

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
)

type Item struct {
	No     int    `json:"n"`
	Artist string `json:"a"`
	Title  string `json:"tt"`
	Time   string `json:"t"`
	Raw    string `json:"r"`
	Hash   string `json:"h"`
}

const C_ETAG_CACHE_KEY = "konsti-etag"
const C_NODE_CACHE_KEY = "playlist-node"

const NodeCacheTime = 60 * time.Minute

const MainViewCacheTime = 10 * time.Minute
const ApiViewCacheTime = 10 * time.Minute

const PlaylistURL = "http://www.konstantinkuehn.com/playlist/"

var BuildVersion = "2019.05.12-12:00:00"
var CacheVersion = "v1.0"
var ServiceWorkerVersion = "v1.0"

func LoadURL(url string, useETag bool) (*html.Node, error) {
	request, _ := http.NewRequest("GET", url, nil)
	etag := ""

	// ETag aus dem Cache laden und zum Request hinzufügen
	if useETag {
		if err := store.Get(C_ETAG_CACHE_KEY, &etag); err == nil {
			if etag != "" {
				request.Header.Add("if-none-match", etag)
			}
		}
	}

	// Den eigentlichen Request ausfürhren
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	var node *html.Node
	log.Printf("Got statuscode %d from konsti.\n", resp.StatusCode)

	// Wenn sich nichts geändert hat, holen wir das Ergebnis aus dem Cache.
	if resp.StatusCode == http.StatusNotModified {
		if err := store.Get(C_NODE_CACHE_KEY, &node); err != nil {
			return LoadURL(url, false)
		}
		return node, nil
	}

	// Ein geändertes ergebnis. Die Rückgabe parsen und
	// im Cache speichern.
	defer resp.Body.Close()
	etagResp := resp.Header.Get("etag")
	if etagResp != "" {
		store.Set(C_ETAG_CACHE_KEY, etagResp, 5*time.Minute)
	}
	r, err := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	if err != nil {
		return nil, err
	}
	node, _ = html.Parse(r)
	store.Set(C_NODE_CACHE_KEY, node, NodeCacheTime)
	return node, nil
}

type ErrorUnchanged struct{}

func (e *ErrorUnchanged) Error() string {
	return "Unchanged"
}

func getData() ([]Item, error) {

	doc, err := LoadURL(PlaylistURL, true)
	if err, ok := err.(*ErrorUnchanged); ok {
		return nil, errors.New(err.Error())
	}
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

	store.Set("items", resp, time.Minute*60)
	return resp, nil
}

var store = persistence.NewInMemoryStore(time.Second)

func mainHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{"Build": BuildVersion})
}

func manifestHandler(c *gin.Context) {
	// https://developers.google.com/web/fundamentals/web-app-manifest/
	c.JSON(http.StatusOK, gin.H{
		"short_name":                  "Suff am Ring",
		"name":                        "Suff am Ring - Die Playlist",
		"manifest_version":            "2",
		"description":                 "...",
		"start_url":                   "/",
		"background_color":            "#070a0a",
		"display":                     "fullscreen",
		"theme_color":                 "#b50e12",
		"orientation":                 "portrait",
		"prefer_related_applications": false,
		"icons": []gin.H{
			gin.H{
				"src":  "/static/images/192.png",
				"type": "image/png",
				"size": "192x192",
			},
			gin.H{
				"src":  "/static/images/512.png",
				"type": "image/png",
				"size": "512x512",
			},
		},
	})
}

func playlistApiHandler(c *gin.Context) {
	if resp, err := getData(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		ifNoneMatch := c.Request.Header.Get("If-None-Match")
		if b, err := json.Marshal(resp); err == nil {
			etag := fmt.Sprintf("\"W\\%x\"", md5.Sum(b))
			if ifNoneMatch == etag {
				c.String(http.StatusNotModified, "")
			} else {
				c.Header("etag", etag)
				c.JSON(http.StatusOK, gin.H{"data": resp})
			}
		}
	}
}

func serviceWorkerHandler(c *gin.Context) {
	c.Header("content-type", "application/javascript")
	c.HTML(http.StatusOK, "sw.js", gin.H{
		"CacheVersion":         CacheVersion,
		"ServiceWorkerVersion": ServiceWorkerVersion,
		"BuildVersion":         BuildVersion,
	})
}

func run(e *gin.Engine) {
	port := os.Getenv("RING_PWA_PORT")
	if port == "" {
		port = "8000"
	}
	host := os.Getenv("RING_PWA_HOST")
	if host == "" {
		host = "127.0.0.1"
		// host = "0.0.0.0"
	}
	e.Run(fmt.Sprintf("%s:%s", host, port))
}

func main() {
	r := gin.Default()

	BuildVersion = time.Now().Format("2006.01.02-15:04:05")
	ServiceWorkerVersion = BuildVersion
	CacheVersion = BuildVersion

	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/*")
	r.GET("/", mainHandler)
	r.GET("/manifest.json", manifestHandler)
	r.GET("/api/playlist", playlistApiHandler)
	r.GET("/sw.js", serviceWorkerHandler)

	run(r)
}
