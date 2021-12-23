package ascii2d

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path"

	"github.com/PuerkitoBio/goquery"
	"github.com/imroc/req"
	log "github.com/sirupsen/logrus"
)

const (
	ascii2dURL string = "https://ascii2d.net/"
	uploadURL  string = "https://ascii2d.net/search/multi"
)

type Config struct {
	Enable         bool   `yaml:"Enable"`
	TempFolderPath string `yaml:"TempFolderPath"`
}

type Result struct {
	ThumbnailURL string
	ImageURL     string
}

type Service struct {
	*Config
}

func (service *Service) Search(fileURL string) *Result {
	// 获取图片
	res, err := req.Get(fileURL)
	if err != nil {
		log.Error(err)
		return nil
	}

	// 用该图片的SHA256作为文件名
	bytes := sha256.Sum256([]byte(fileURL))
	fileName := fmt.Sprintf("%x%s", string(bytes[:]), path.Ext(fileURL))

	// 创建存放图片的临时文件夹
	err = os.MkdirAll(service.TempFolderPath, os.ModeDir)
	if err != nil {
		log.Error(err)
		return nil
	}

	// 储存图片至临时文件夹
	filePath := path.Join(service.TempFolderPath, fileName)
	err = res.ToFile(filePath)
	if err != nil {
		log.Error(err)
		return nil
	}

	// 重新读取图片
	var file *os.File
	file, err = os.Open(filePath)
	if err != nil {
		log.Error(err)
		return nil
	}

	// 获取ascii2d网站的一次性token
	var token string
	token = service.getToken()

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
		log.Error(err)
		return nil
	}
	if res.Response().StatusCode != 200 {
		log.Errorf("通过%s上传图片时收到了200以外的StatusCode：%s", uploadURL, res.Response().Status)
		return nil
	}

	// 提取搜索结果html
	doc, err := goquery.NewDocumentFromReader(res.Response().Body)
	if err != nil {
		log.Error(err)
		return nil
	}
	first := doc.Find(".item-box:has(h6)").First()
	url, exist1 := first.Find(".info-box .detail-box a").First().Attr("href")
	thumbPath, exist2 := first.Find("img[loading=\"lazy\"]").Attr("src")

	if !exist1 || !exist2 {
		log.Error("发生未知错误：无法从asii2d搜索结果页面的html中获取搜索结果")
		return nil
	}

	return &Result{
		ThumbnailURL: ascii2dURL + thumbPath,
		ImageURL:     url,
	}
}

// 获取ascii2d网站的一次性token
func (service *Service) getToken() string {
	res, err := req.Get(ascii2dURL)
	if err != nil {
		log.Error(err)
		return ""
	}
	if res.Response().StatusCode != 200 {
		log.Errorf("访问%s时收到了200以外的StatusCode：%s", ascii2dURL, res.Response().Status)
		return ""
	}

	doc, err := goquery.NewDocumentFromReader(res.Response().Body)
	if err != nil {
		log.Error(err)
		return ""
	}

	// Find the review items
	token, exist := doc.Find("head meta[name=\"csrf-token\"]").Attr("content")

	if !exist {
		log.Errorf("发生未知错误：无法从%s的html中获取token", ascii2dURL)
		return ""
	}

	return token
}
