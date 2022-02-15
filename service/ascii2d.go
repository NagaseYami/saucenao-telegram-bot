package service

import (
	"path"

	"github.com/NagaseYami/telegram-bot/tool"
	"github.com/go-rod/rod"
	"github.com/go-rod/stealth"
	"github.com/imroc/req/v3"
	log "github.com/sirupsen/logrus"
)

const (
	ascii2dURL                   = "https://ascii2d.net"
	firstResultThumbnailSelector = "body > div > div > div.col-xs-12.col-lg-8.col-xl-8 > div:nth-child(6) > div.col-xs-12.col-sm-12.col-md-4.col-xl-4.text-xs-center.image-box > img"
	firstResultURLSelector       = "body > div > div > div.col-xs-12.col-lg-8.col-xl-8 > div:nth-child(6) > div.col-xs-12.col-sm-12.col-md-8.col-xl-8.info-box > div.detail-box.gray-link > h6:nth-child(1) > a:nth-child(2)"
)

type Ascii2dConfig struct {
	Enable        bool   `yaml:"Enable"`
	TempDirectory string `yaml:"TempDirectory"`
}

type Ascii2dService struct {
	*Ascii2dConfig
	reqClient *req.Client
	page      *rod.Page
}

type Ascii2dResult struct {
	ColorThumbnail string
	ColorURL       string
	BovwThumbnail  string
	BovwURL        string
}

func (service *Ascii2dService) Init() {
	service.reqClient = req.C().SetOutputDirectory(service.TempDirectory)
	service.page = stealth.MustPage(tool.Browser.RodBrowser)
}

func (service *Ascii2dService) Search(fileURL string) *Ascii2dResult {
	fileName := path.Base(fileURL)
	_, err := service.reqClient.R().SetOutputFile(fileName).Get(fileURL)
	if err != nil {
		log.Errorln(err)
		return &Ascii2dResult{}
	}

	service.page.MustNavigate(ascii2dURL).MustWaitLoad()
	service.page.MustElement("#file-form").MustSetFiles(path.Join(service.TempDirectory, fileName))

	service.page = service.page.MustElement("#file_upload > div > div.col-sm-1.col-xs-12 > button").MustClick().Page().MustWaitLoad()
	colorThumb := ascii2dURL + *service.page.MustElement(firstResultThumbnailSelector).MustAttribute("src")
	colorURL := *service.page.MustElement(firstResultURLSelector).MustAttribute("href")

	service.page = service.page.MustElement("body > div > div > div.col-xs-12.col-lg-8.col-xl-8 > div:nth-child(3) > div.detail-link.pull-xs-right.hidden-sm-down.gray-link > span:nth-child(2) > a").MustClick().Page()

	bovwThumbnail := ascii2dURL + *service.page.MustElement(firstResultThumbnailSelector).MustAttribute("src")
	bovwURL := *service.page.MustElement(firstResultURLSelector).MustAttribute("href")

	log.Debugf("ascii2d搜索结果：\n色合：%s\n特征：%s\n", colorURL, bovwURL)

	return &Ascii2dResult{
		ColorThumbnail: colorThumb,
		ColorURL:       colorURL,
		BovwThumbnail:  bovwThumbnail,
		BovwURL:        bovwURL,
	}
}
