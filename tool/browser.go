package tool

import (
	"time"

	"github.com/go-rod/rod"
)

var Browser MyBrowser

type MyBrowser struct {
	RodBrowser *rod.Browser
}

func (b *MyBrowser) Init() {
	b.RodBrowser = rod.New().Timeout(time.Second * 30).MustConnect()
}

func (b *MyBrowser) UnInit() {
	b.RodBrowser.MustClose()
}
