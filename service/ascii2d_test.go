package service

import (
	"testing"

	"github.com/NagaseYami/telegram-bot/tool"
)

func TestAscii2dSearch(T *testing.T) {
	tool.Browser.Init()
	defer tool.Browser.UnInit()
	ascii2d := Ascii2dService{
		Ascii2dConfig: &Ascii2dConfig{Enable: true, TempDirectory: "../temp"},
		reqClient:     nil,
	}
	ascii2d.Init()
	result := ascii2d.Search("https://p.sda1.dev/5/d5b30c1295c66793b00a0e9e71fc0b15/FLjM4FiaMAAYbwu.png")
	T.Log(result)
}
