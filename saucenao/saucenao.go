package saucenao

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"

	. "github.com/ahmetb/go-linq/v3"
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
	Database string
	URL      string
}

const apiURL string = "https://saucenao.com/search.php?api_key=%s&db=999&output_type=2&numres=9&url=%s"

var apiKey string

func Init() {
	apiKey = os.Getenv("SAUCENAO_API_KEY")
	if apiKey == "" {
		log.Fatal("环境变量「SAUCENAO_API_KEY」缺失")
	}
}

func Search(fileURL string) (Header, []Result, error) {
	escapedURL := url.PathEscape(fileURL)
	resp, err := http.Get(fmt.Sprintf(apiURL, apiKey, escapedURL))

	if err != nil {
		return Header{}, []Result{}, err
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
	searchResultData := make(map[string]string)
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

		if source != "" && strings.Contains(source, "https://") {
			urls = append(urls, source)
		}

		for _, u := range urls {

			// 从P站多图Artwork的单图链接中提取Artwork链接
			if strings.Contains(u, "https://i.pximg.net") {
				fileName := path.Base(u)
				noExt := strings.Replace(fileName, path.Ext(fileName), "", 1)
				re := regexp.MustCompile(`_p[0-9]+`)
				pixivID := re.ReplaceAllString(noExt, "")
				u = fmt.Sprintf("https://www.pixiv.net/artworks/%s", pixivID)
			}

			// 将旧格式P站链接转换为新格式P站链接
			re := regexp.MustCompile(`www.pixiv.net/member_illust.php\?mode=medium&illust_id=([0-9]+)`)
			u = re.ReplaceAllString(u, "www.pixiv.net/artworks/${1}")

			searchResultData[u] = GetDatabaseFromURL(u)
		}
	}

	var results []Result
	From(searchResultData).Select(func(i interface{}) interface{} {
		return Result{
			Database: i.(KeyValue).Value.(string),
			URL:      i.(KeyValue).Key.(string),
		}
	}).ToSlice(&results)

	return searchResultHeader, results, err
}

func GetDatabaseFromURL(url string) string {
	if strings.Contains(url, "www.pixiv.net") {
		return "Pixiv"
	} else if strings.Contains(url, "danbooru.donmai.us") {
		return "Danbooru"
	} else if strings.Contains(url, "gelbooru.com") {
		return "Gelbooru"
	} else if strings.Contains(url, "chan.sankakucomplex.com") {
		return "Sankaku"
	} else if strings.Contains(url, "anime-pictures.net") {
		return "Anime Pictures"
	} else if strings.Contains(url, "i.redd.it") {
		return "Reddit"
	} else if strings.Contains(url, "yande.re") {
		return "Yandere"
	} else if strings.Contains(url, "www.imdb.com") {
		return "IMDB"
	} else if strings.Contains(url, "deviantart.com") {
		return "Deviantart"
	} else if strings.Contains(url, "twitter.com") {
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
	} else if strings.Contains(url, "bcy.net") {
		return "半次元"
	} else if strings.Contains(url, "konachan.com") {
		return "Konachan"
	} else if strings.Contains(url, "fanbox.cc") {
		return "Pixiv Fanbox"
	} else if strings.Contains(url, "e621.net") {
		return "e621"
	} else if strings.Contains(url, "exhentai.org") {
		return "exhentai"
	} else {
		return "Unknown"
	}
}
