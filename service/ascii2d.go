package service

import (
	"path"

	"github.com/NagaseYami/saucenao-telegram-bot/tool"
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
	client *req.Client
}

type Ascii2dResult struct {
	ColorThumbnail string
	ColorURL       string
	BovwThumbnail  string
	BovwURL        string
}

func (service *Ascii2dService) Init() {
	service.client = req.C().SetOutputDirectory(service.TempDirectory)
}

func (service *Ascii2dService) Search(fileURL string) *Ascii2dResult {
	fileName := path.Base(fileURL)
	_, err := service.client.R().SetOutputFile(fileName).Get(fileURL)
	if err != nil {
		log.Errorln(err)
		return &Ascii2dResult{}
	}

	page := stealth.MustPage(tool.Browser.RodBrowser)
	page.MustNavigate(ascii2dURL)
	page.MustWaitLoad()

	page.MustElement("#file-form").MustSetFiles(path.Join(service.TempDirectory, fileName))
	page = page.MustElement("#file_upload > div > div.col-sm-1.col-xs-12 > button").MustClick().Page()
	page.MustWaitLoad()

	colorThumb := path.Join(ascii2dURL, *page.MustElement(firstResultThumbnailSelector).MustAttribute("src"))
	colorURL := *page.MustElement(firstResultURLSelector).MustAttribute("href")

	page = page.MustElement("body > div > div > div.col-xs-12.col-lg-8.col-xl-8 > div:nth-child(3) > div.detail-link.pull-xs-right.hidden-sm-down.gray-link > span:nth-child(2) > a").MustClick().Page()

	bovwThumbnail := path.Join(ascii2dURL, *page.MustElement(firstResultThumbnailSelector).MustAttribute("src"))
	bovwURL := *page.MustElement(firstResultURLSelector).MustAttribute("href")
	return &Ascii2dResult{
		ColorThumbnail: colorThumb,
		ColorURL:       colorURL,
		BovwThumbnail:  bovwThumbnail,
		BovwURL:        bovwURL,
	}
}
