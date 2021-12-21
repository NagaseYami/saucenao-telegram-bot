package saucenao

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/tidwall/gjson"
	tb "gopkg.in/tucnak/telebot.v2"
)

const apiURL string = "https://saucenao.com/search.php?api_key=%s&db=999&output_type=2&numres=9&url=%s"

type Config struct {
	Enable     bool    `yaml:"Enable"`
	ApiKey     string  `yaml:"ApiKey"`
	Similarity float64 `yaml:"Similarity"`
}

type Service struct {
	*Config
}

var Instance *Service

func (service *Service) Search(fileURL string) (*Result, error) {
	// 访问API
	resp, err := http.Get(fmt.Sprintf(apiURL, service.ApiKey, url.PathEscape(fileURL)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	io.Copy(buf, resp.Body)

	// 将Response转为json
	gResult := gjson.ParseBytes(buf.Bytes())

	// 从Body中获取搜索结果
	jsonResults := gResult.Get("results").Array()
	searchResultData := make(map[string]string)
	for _, r := range jsonResults {

		// 相似度低于一定程度则跳过
		if r.Get("header.similarity").Float() < service.Similarity {
			continue
		}

		// 从ext_urls中获取图片url
		var urls []string
		for _, u := range r.Get("data.ext_urls").Array() {
			urls = append(urls, u.String())
		}

		// 从sauce中获取图片url
		source := r.Get("data.source").String()
		if source != "" && strings.Contains(source, "https://") {
			urls = append(urls, source)
		}

		// 如果没有任何url，则跳过
		if len(urls) == 0 {
			continue
		}

		// 将上述url进行平坦化处理
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

			searchResultData[u] = service.GetDatabaseFromURL(u)
		}
	}

	// 从Header中获取API剩余可用次数
	jsonHeader := gResult.Get("header")

	return NewResult(jsonHeader.Get("short_limit").Int(),
		jsonHeader.Get("long_limit").Int(),
		jsonHeader.Get("short_remaining").Int(),
		jsonHeader.Get("long_remaining").Int(),
		searchResultData), err
}

func (service *Service) GetDatabaseFromURL(url string) string {
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
	} else if strings.Contains(url, "fantia.jp") {
		return "Fantia"
	} else {
		return "Unknown"
	}
}

type Result struct {
	Text        string
	URLSelector *tb.ReplyMarkup
	Success     bool
}

func NewResult(
	shortLimit int64, longLimit int64, shortRemain int64, longRemain int64, searchResult map[string]string,
) *Result {
	selector := &tb.ReplyMarkup{}
	var buttons []tb.Btn
	for key, value := range searchResult {
		buttons = append(buttons, tb.Btn{
			Text: value,
			URL:  key,
		})
	}
	var rows []tb.Row
	for i := 0; i < int(math.Ceil(float64(len(buttons))/3.0)); i++ {
		if len(buttons)-(i+1)*3 < 0 {
			rows = append(rows, selector.Row(buttons[i*3:]...))
		} else {
			rows = append(rows, selector.Row(buttons[i*3:i*3+3]...))
		}
	}
	selector.Inline(rows...)

	success := len(searchResult) != 0 && shortRemain > 0 && longRemain > 0

	var text string

	if shortRemain <= 0 {
		text = "搜索过于频繁，已达到30秒内搜索次数上限\nSauceNAO搜索失败"
	} else if longRemain <= 0 {
		text = "搜索过于频繁，已达到24小时内搜索次数上限\nSauceNAO搜索失败"
	} else if success {
		text = "SauceNAO搜索完毕"
	} else {
		text = fmt.Sprintf("SauceNAO搜索失败（搜索结果相似度均低于%g）", Instance.Similarity)
	}

	return &Result{
		Text:        text,
		URLSelector: selector,
		Success:     success,
	}
}
