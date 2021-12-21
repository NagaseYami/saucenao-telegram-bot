package ascii2d

import (
	"testing"
)

func TestGetToken(t *testing.T) {
	t.Log("Getting ascii2d token")
	token, err := getToken()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(token)
	t.Log("Finish")
}

func TestSearch(t *testing.T) {
	service := &Service{Config:&Config{
		Enable:         true,
		TempFolderPath: "../../temp",
	}}
	t.Log("Testing ascii2d search")
	result, err := service.Search("https://pbs.twimg.com/media/E9n_MXLUYAY6c3B.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if result==nil {
		t.Log("No result")
	} else {
		t.Log(result.URLSelector.InlineKeyboard[0][0].URL)
	}
	t.Log("Finish")
}
