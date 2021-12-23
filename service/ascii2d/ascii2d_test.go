package ascii2d

import (
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestGetToken(t *testing.T) {
	service := &Service{
		Config: &Config{
			Enable:         true,
			TempFolderPath: "../../temp",
		},
	}
	t.Log("Getting ascii2d token")
	token := service.getToken()
	if token == "" {
		log.Fatal("Can't get token")
	}
	t.Log(token)
	t.Log("Finish")
}

func TestSearch(t *testing.T) {
	service := &Service{
		Config: &Config{
			Enable:         true,
			TempFolderPath: "../../temp",
		},
	}
	t.Log("Testing ascii2d search")
	result := service.Search("https://pbs.twimg.com/media/E9n_MXLUYAY6c3B.jpg")
	if result == nil {
		t.Fatal("No result")
	} else {
		t.Log(result.ImageURL)
	}
	t.Log("Finish")
}
