package saucenao

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

type Header struct {
	ShortLimit  string
	LongLimit   string
	ShortRemain string
	LongRemain  string
}
type Result struct {
	Databases []string
	URLs      []string
}

var apiKey string

func Init() {
	apiKey = os.Getenv("SAUCENAO_API_KEY")
	if apiKey == "" {
		log.Fatal("环境变量「SAUCENAO_API_KEY」缺失")
	}
}

func Search(fileURL string) (Header, []Result) {
	escapedURL := url.PathEscape(fileURL)
	resp, err := http.Get(fmt.Sprintf("https://saucenao.com/search.php?api_key=%s&db=999&output_type=2&numres=9&url=%s",
		apiKey, escapedURL))

	if err != nil {
		log.Fatal("")
	}

	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	io.Copy(buf, resp.Body)

	gResult := gjson.ParseBytes(buf.Bytes())

	jsonHeader := gResult.Get("header")

	searchResultHeader := Header{
		ShortLimit:  jsonHeader.Get("short_limit").String(),
		LongLimit:   jsonHeader.Get("long_limit").String(),
		ShortRemain: jsonHeader.Get("short_remaining").String(),
		LongRemain:  jsonHeader.Get("long_remaining").String(),
	}

	jsonResults := gResult.Get("results").Array()
	var searchResultData []Result
	for _, r := range jsonResults {

		if r.Get("header.similarity").Float() < 80 {
			continue
		}

		var urls []string
		for _, u := range r.Get("data.ext_urls").Array() {
			urls = append(urls, u.String())
		}

		if len(urls) == 0 {
			continue
		}

		source := r.Get("data.source").String()

		if source != "" && strings.Contains(source, "https://") && !strings.Contains(source, "i.pximg.net") {
			urls = append(urls, source)
		}

		var databases []string
		for _, u := range urls {
			databases = append(databases, GetDatabaseFromURL(u))
		}

		searchResultData = append(searchResultData, Result{
			Databases: databases,
			URLs:      urls,
		})
	}

	return searchResultHeader, searchResultData
}

func GetDatabaseFromURL(url string) string {
	if strings.Contains(url, "pixiv") {
		return "Pixiv"
	} else if strings.Contains(url, "danbooru") {
		return "Danbooru"
	} else if strings.Contains(url, "gelbooru") {
		return "Gelbooru"
	} else if strings.Contains(url, "sankaku") {
		return "Sankaku"
	} else if strings.Contains(url, "anime-pictures.net") {
		return "Anime Pictures"
	} else if strings.Contains(url, "i.redd.it") {
		return "Reddit"
	} else if strings.Contains(url, "yande.re") {
		return "Yandere"
	} else if strings.Contains(url, "imdb") {
		return "IMDB"
	} else if strings.Contains(url, "deviantart") {
		return "Deviantart"
	} else if strings.Contains(url, "twitter") {
		return "Twitter"
	} else if strings.Contains(url, "nijie.info") {
		return "Nijie"
	} else if strings.Contains(url, "pawoo.net") {
		return "Pawoo"
	} else if strings.Contains(url, "seiga.nicovideo.jp") {
		return "Seiga Nicovideo"
	} else if strings.Contains(url, "tumblr.com") {
		return "Tumblr"
	} else if strings.Contains(url, "anidb.net") {
		return "Anidb"
	} else if strings.Contains(url, "mangadex.org") {
		return "MangaDex"
	} else if strings.Contains(url, "mangaupdates.com") {
		return "MangaUpdates"
	} else if strings.Contains(url, "myanimelist.net") {
		return "MyAnimeList"
	} else if strings.Contains(url, "furaffinity.net") {
		return "FurAffinity"
	} else if strings.Contains(url, "artstation.com") {
		return "ArtStation"
	} else {
		return "Unknown"
	}
}
