package tool

import (
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	log "github.com/sirupsen/logrus"
)

var Browser MyBrowser

type MyBrowser struct {
	RodBrowser *rod.Browser
}

func (b *MyBrowser) Init() {
	path, has := launcher.LookPath()
	if !has {
		log.Fatal("未能找到支持的浏览器，请先安装go-rod支持的浏览器")
	}
	url := launcher.New().Bin(path).
		Headless(true).
		NoSandbox(true).
		Set("–disable-sync").
		Set("–no-first-run").
		Set("--no-startup-window").
		Set("--disable-extensions").
		MustLaunch()
	log.Info("浏览器启动成功")
	b.RodBrowser = rod.New().ControlURL(url).MustConnect()
	pages := b.RodBrowser.MustPages()
	if !pages.Empty() {
		for _, page := range pages {
			page.MustClose()
		}
	}
}

func (b *MyBrowser) UnInit() {
	b.RodBrowser.MustClose()
}
