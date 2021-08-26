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
	t.Log("Testing ascii2d search")
	result, err := Search("https://pbs.twimg.com/media/E9n_MXLUYAY6c3B.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if !result.Exist {
		t.Log("No result")
	} else {
		t.Log(result.ThumbnailURL)
		t.Log(result.URL)
	}
	t.Log("Finish")
}
