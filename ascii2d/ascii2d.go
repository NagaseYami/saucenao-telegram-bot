package ascii2d

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/PuerkitoBio/goquery"
	"github.com/imroc/req"
)

const ascii2dURL string = "https://ascii2d.net/"
const uploadURL string = "https://ascii2d.net/search/multi"
const tempFolder string = "temp"

type Result struct {
	ThumbnailURL string
	URL          string
	Exist        bool
}

func Search(fileURL string) (Result, error) {
	res, err := req.Get(fileURL)

	if err != nil {
		return Result{}, err
	}

	bytes := sha256.Sum256([]byte(fileURL))
	fileName := fmt.Sprintf("%x%s", string(bytes[:]), path.Ext(fileURL))
	err = os.Mkdir(tempFolder, os.ModeDir)
	if err != nil {
		return Result{}, err
	}
	filePath := path.Join(tempFolder, fileName)
	err = res.ToFile(filePath)
	if err != nil {
		return Result{}, err
	}

	var file *os.File
	file, err = os.Open(filePath)
	if err != nil {
		return Result{}, err
	}

	var token string
	token, err = getToken()
	if err != nil {
		return Result{}, err
	}

	param := req.Param{
		"authenticity_token": token,
		"utf8":               "✓",
	}

	res, err = req.Post(uploadURL, req.FileUpload{
		FileName:  fileName,
		FieldName: "file",
		File:      file,
	}, param)

	if err != nil {
		return Result{}, err
	}

	if res.Response().StatusCode != 200 {
		return Result{}, errors.New(res.Response().Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Response().Body)
	if err != nil {
		return Result{}, err
	}

	url, exist1 := doc.Find(".row .item-box .info-box .detail-box a").First().Attr("href")
	thumbPath, exist2 := doc.Find(".item-box .image-box img[loading=\"eager\"]").Attr("src")

	return Result{
		ThumbnailURL: ascii2dURL + thumbPath,
		URL:          url,
		Exist:        exist1 && exist2,
	}, err

}

func getToken() (string, error) {

	res, err := req.Get(ascii2dURL)
	if err != nil {
		return "", err
	}

	doc, err := goquery.NewDocumentFromReader(res.Response().Body)
	if err != nil {
		return "", err
	}

	// Find the review items
	token, exist := doc.Find("head meta[name=\"csrf-token\"]").Attr("content")

	if !exist {
		return "", errors.New("Token not found")
	}

	return token, err
}
