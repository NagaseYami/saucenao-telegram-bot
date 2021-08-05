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
	ShortLimit        string
	LongLimit         string
	ShortRemain       string
	LongRemain        string
	MinimumSimilarity string
}
type Result struct {
	DataBaseName string
	URL          string
}

var apiKey string

func Init() {
	apiKey = os.Getenv("SAUCENAO_API_KEY")
	if apiKey == "" {
		log.Fatal("SauceNAO api key not found.")
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
	header := gResult.Get("header")
	urls := gResult.Get("results.#.data.ext_urls.0").Array()

	var results []Result

	for _, url := range urls {
		results = append(results, Result{DataBaseName: GetDatabaseFromURL(url.String()), URL: url.String()})
	}

	return Header{
		ShortLimit:        header.Get("short_limit").String(),
		LongLimit:         header.Get("long_limit").String(),
		ShortRemain:       header.Get("short_remaining").String(),
		LongRemain:        header.Get("long_remaining").String(),
		MinimumSimilarity: header.Get("minimum_similarity").String(),
	}, results
}

func GetDatabaseFromURL(url string) string {
	if strings.Contains(url, "i.pximg.net") ||
		strings.Contains(url, "pixiv") {
		return "Pixiv"
	} else if strings.Contains(url, "danbooru") {
		return "Danbooru"
	} else if strings.Contains(url, "gelbooru") {
		return "Gelbooru"
	} else if strings.Contains(url, "sankaku") {
		return "Sankaku"
	} else if strings.Contains(url, "anime-pictures.net") {
		return "Anime Pictures"
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
