package ascii2d

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/PuerkitoBio/goquery"
	"github.com/imroc/req"
	tb "gopkg.in/tucnak/telebot.v2"
)

const ascii2dURL string = "https://ascii2d.net/"
const uploadURL string = "https://ascii2d.net/search/multi"

type Config struct {
	Enable         bool   `yaml:"Enable"`
	TempFolderPath string `yaml:"TempFolderPath"`
}

type Service struct {
	*Config
}

var Instance *Service

func (service *Service) Search(fileURL string) (*Result, error) {

	// 获取图片
	res, err := req.Get(fileURL)
	if err != nil {
		return &Result{}, err
	}

	// 用该图片的SHA256作为文件名
	bytes := sha256.Sum256([]byte(fileURL))
	fileName := fmt.Sprintf("%x%s", string(bytes[:]), path.Ext(fileURL))

	// 创建存放图片的临时文件夹
	err = os.MkdirAll(service.TempFolderPath, os.ModeDir)
	if err != nil {
		return nil, err
	}

	// 储存图片至临时文件夹
	filePath := path.Join(service.TempFolderPath, fileName)
	err = res.ToFile(filePath)
	if err != nil {
		return nil, err
	}

	// 重新读取图片
	var file *os.File
	file, err = os.Open(filePath)
	if err != nil {
		return nil, err
	}

	// 获取ascii2d网站的一次性token
	var token string
	token, err = service.getToken()
	if err != nil {
		return nil, err
	}

	// 上传图片并搜索
	res, err = req.Post(uploadURL, req.FileUpload{
		FileName:  fileName,
		FieldName: "file",
		File:      file,
	}, req.Param{
		"authenticity_token": token,
		"utf8":               "✓",
	})
	if err != nil {
		return nil, err
	}
	if res.Response().StatusCode != 200 {
		return nil, errors.New(res.Response().Status)
	}

	// 提取搜索结果html
	doc, err := goquery.NewDocumentFromReader(res.Response().Body)
	if err != nil {
		return nil, err
	}
	first := doc.Find(".item-box:has(h6)").First()
	url, exist1 := first.Find(".info-box .detail-box a").First().Attr("href")
	thumbPath, exist2 := first.Find("img[loading=\"lazy\"]").Attr("src")

	if !exist1 || !exist2 {
		return nil, err
	}

	return NewResult(ascii2dURL+thumbPath, url), err

}

// 获取ascii2d网站的一次性token
func (service *Service) getToken() (string, error) {

	res, err := req.Get(ascii2dURL)
	if err != nil {
		return "", err
	}
	if res.Response().StatusCode != 200 {
		return "", errors.New(res.Response().Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Response().Body)
	if err != nil {
		return "", err
	}

	// Find the review items
	token, exist := doc.Find("head meta[name=\"csrf-token\"]").Attr("content")

	if !exist {
		return "", errors.New("无法获取ascii2d的token")
	}

	return token, err
}

type Result struct {
	Photo       *tb.Photo
	URLSelector *tb.ReplyMarkup
}

func NewResult(thumbnailURL string, url string) *Result {

	photo := &tb.Photo{File: tb.FromURL(thumbnailURL)}
	selector := &tb.ReplyMarkup{}
	selector.Inline(tb.Row{
		tb.Btn{
			Text: "ascii2d搜索结果",
			URL:  url,
		},
	})

	return &Result{
		Photo:       photo,
		URLSelector: selector,
	}
}
