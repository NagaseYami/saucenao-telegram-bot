package saucenao

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/ahmetb/go-linq/v3"
	"github.com/imroc/req"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

const (
	apiURL           string = "https://saucenao.com/search.php?api_key=%s&db=999&output_type=2&url=%s"
	nhentaiURL       string = "https://nhentai.net"
	nhentaiSearchURL string = "https://nhentai.net/search/?q=%s"
	ehentaiSearchURL string = "https://e-hentai.org/?f_search=%s&f_sname=on"
)

// https://saucenao.com/tools/examples/api/index_details.txt
var saucenaoDatabaseIndexList = map[string]int64{
	"h-mags":          0,
	"h-anime":         1,
	"hcg":             2,
	"ddb-objects":     3,
	"ddb-samples":     4,
	"Pixiv":           5,
	"PixivHistorical": 6,
	"anime":           7,
	"NicoNico Seiga":  8,
	"Danbooru":        9,
	"drawr":           10,
	"Nijie":           11,
	"yande.re":        12,
	"animeop":         13,
	"IMDb":            14,
	"Shutterstock":    15,
	"FAKKU":           16,
	"nhentai":         18,
	"2d_market":       19,
	"medibang":        20,
	"Anime":           21,
	"H-Anime":         22,
	"Movies":          23,
	"Shows":           24,
	"gelbooru":        25,
	"konochan":        26,
	"sankaku":         27,
	"anime-pictures":  28,
	"e621":            29,
	"idol complex":    30,
	"bcy illust":      31,
	"bcy cosplay":     32,
	"portalgraphics":  33,
	"dA":              34,
	"pawoo":           35,
	"madokami":        36,
	"mangadex":        37,
	"ehentai":         38,
	"ArtStation":      39,
	"FurAffinity":     40,
	"Twitter":         41,
	"Furry Network":   42,
}

type Config struct {
	Enable     bool    `yaml:"Enable"`
	ApiKey     string  `yaml:"ApiKey"`
	Similarity float64 `yaml:"Similarity"`
}

type Result struct {
	ShortRemain  int64
	LongRemain   int64
	SearchResult map[string]string
}

type Service struct {
	*Config
}

func (service *Service) Search(fileURL string) (*Result, error) {
	// 访问API
	resp, err := http.Get(fmt.Sprintf(apiURL, service.ApiKey, url.PathEscape(fileURL)))
	if err != nil {
		return nil, err
	}
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			log.Error(err)
		}
	}()
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return nil, err
	}

	type ArtworkData struct {
		Similarity float64
		URL        string
	}

	result := make(map[int64][]ArtworkData)
	gResult := gjson.ParseBytes(buf.Bytes())

	// 相似度底线要求
	minimumSimilarity := gResult.Get("header.minimum_similarity").Float()

	// 遍历全部结果
	jsonResults := gResult.Get("results").Array()
	for _, r := range jsonResults {

		similarity := r.Get("header.similarity").Float()

		// 相似度低于底线则跳过
		if similarity < minimumSimilarity {
			continue
		}

		dbIndex := r.Get("header.index_id").Int()
		var extURLs []string
		linq.From(r.Get("data.ext_urls").Array()).SelectT(func(r gjson.Result) string { return r.String() }).ToSlice(&extURLs)
		engName := r.Get("data.eng_name").String()
		jpName := r.Get("data.jp_name").String()

		// 获取URL
		var artworkURL string
		switch dbIndex {
		case saucenaoDatabaseIndexList["ehentai"]:
			artworkURL = service.getEHentaiGallery(engName, jpName)
		case saucenaoDatabaseIndexList["nhentai"]:
			artworkURL = service.getNHentaiGallery(engName, jpName)
		case saucenaoDatabaseIndexList["Pixiv"]:
			artworkURL = service.getPixivArtwork(extURLs[0])
		default:
			if len(extURLs) > 0 {
				// 多个ext_urls时只取第一个
				artworkURL = extURLs[0]
			} else {
				log.Warn("遇到了未知的没有ext_urls的Database：%d\n被搜索的图片URL：%s", dbIndex, fileURL)
				continue
			}
		}

		if artworkURL == "" {
			continue
		}

		result[dbIndex] = append(result[dbIndex], ArtworkData{Similarity: similarity, URL: artworkURL})
	}

	// 去重
	searchResult := make(map[string]string)
	for _, artworks := range result {
		var fixed string
		if len(artworks) > 1 {
			var ordered []ArtworkData
			linq.From(artworks).OrderByDescendingT(func(data ArtworkData) float64 {
				return data.Similarity
			}).ToSlice(&ordered)
			fixed = ordered[0].URL
		} else {
			fixed = artworks[0].URL
		}
		searchResult[service.getDatabaseFromURL(fixed)] = fixed
	}

	// 从Header中获取API剩余可用次数
	jsonHeader := gResult.Get("header")

	return &Result{
		ShortRemain:  jsonHeader.Get("short_remaining").Int(),
		LongRemain:   jsonHeader.Get("long_remaining").Int(),
		SearchResult: searchResult,
	}, err
}

func (service *Service) getPixivArtwork(extURL string) string {
	// 从P站多图Artwork的单图链接中提取Artwork链接
	if strings.Contains(extURL, "https://i.pximg.net") {
		fileName := path.Base(extURL)
		noExt := strings.Replace(fileName, path.Ext(fileName), "", 1)
		re := regexp.MustCompile(`_p[0-9]+`)
		pixivID := re.ReplaceAllString(noExt, "")
		extURL = fmt.Sprintf("https://www.pixiv.net/artworks/%s", pixivID)
	}

	// 将旧格式P站链接转换为新格式P站链接
	re := regexp.MustCompile(`www.pixiv.net/member_illust.php\?mode=medium&illust_id=([0-9]+)`)
	extURL = re.ReplaceAllString(extURL, "www.pixiv.net/artworks/${1}")

	return extURL
}

func (service *Service) getEHentaiGallery(engName string, jpName string) string {
	if engName == "" && jpName == "" {
		return ""
	}
	name := engName
	if name == "" {
		name = jpName
	}
	resp, err := req.Get(fmt.Sprintf(ehentaiSearchURL, url.PathEscape(name)))
	if err != nil {
		log.Warnf("e-hentai搜索失败\n搜索关键词%s\n错误信息%s", name, err.Error())
		return ""
	}

	doc, err := goquery.NewDocumentFromReader(resp.Response().Body)
	if err != nil {
		log.Warnf("提取e-hentai搜索结果HTML时发生错误：%s", err.Error())
		return ""
	}

	result, exists := doc.Find(".glname a").First().Attr("href")
	if !exists {
		return ""
	}

	return result
}

func (service *Service) getNHentaiGallery(engName string, jpName string) string {
	if engName == "" && jpName == "" {
		return ""
	}

	name := engName
	if name == "" {
		name = jpName
	}

	resp, err := req.Get(fmt.Sprintf(nhentaiSearchURL, url.PathEscape(name)))
	if err != nil {
		log.Warnf("nhentai搜索失败\n搜索关键词%s\n错误信息%s", name, err.Error())
		return ""
	}

	doc, err := goquery.NewDocumentFromReader(resp.Response().Body)
	if err != nil {
		log.Warnf("提取nhentai搜索结果HTML时发生错误：%s", err.Error())
		return ""
	}

	result, exists := doc.Find(".gallery a").First().Attr("href")
	if !exists {
		return ""
	}

	return nhentaiURL + result
}

func (service *Service) getDatabaseFromURL(url string) string {
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
	} else if strings.Contains(url, "e-hentai.org") {
		return "e-hentai"
	} else if strings.Contains(url, "nhentai.net") {
		return "nhentai"
	} else if strings.Contains(url, "fantia.jp") {
		return "Fantia"
	} else {
		return "Unknown"
	}
}
