package ascii2d

import tb "gopkg.in/tucnak/telebot.v2"

type Result struct {
	Photo          *tb.Photo
	URLSelector    *tb.ReplyMarkup
}

func NewResult(thumbnailURL string, url string) *Result {

	photo := &tb.Photo{File: tb.FromURL(thumbnailURL)}
	selector := &tb.ReplyMarkup{}
	selector.Inline(tb.Row{
		tb.Btn{
			Text: "ascii2d搜索结果",
			URL:  url,
		},
	})

	return &Result{
		Photo:          photo,
		URLSelector:    selector,
	}
}
