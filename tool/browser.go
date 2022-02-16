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
	log.Debugf("已找到浏览器，路径：%s", path)
	url := launcher.New().Bin(path).MustLaunch()
	log.Debugf("浏览器启动成功，Devtools监听地址：%s", url)
	b.RodBrowser = rod.New().ControlURL(url)
	err := b.RodBrowser.Connect()
	if err != nil {
		log.Fatalf("尝试连接到浏览器时发生了错误：%s", err)
	}
	log.Debug("成功连接到浏览器")
}

func (b *MyBrowser) UnInit() {
	b.RodBrowser.MustClose()
}
