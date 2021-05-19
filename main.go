package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"time"

	"github.com/valyala/fasthttp"
	tb "gopkg.in/tucnak/telebot.v2"
)

var domain = "https://imgs.zhxiao1124.design"
var pics []string

func GetPics() {
	statusCode, body, err := fasthttp.Get(nil, domain+"/get")
	if err != nil {
		log.Fatalln(err.Error())
	}
	if statusCode == http.StatusOK {
		if err := json.Unmarshal(body, &pics); err != nil {
			log.Fatalln(err.Error())
		}
	}
}

func getpicuris(count int) []*tb.Photo {
	var res []*tb.Photo
	for i := 0; i < count; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(pics))))
		uri := fmt.Sprintf("http://imgs.zhxiao1124.design/static/%s", pics[n.Int64()])
		res = append(res, &tb.Photo{File: tb.FromURL(uri)})
	}
	return res
}

func main() {
	GetPics()
	b, err := tb.NewBot(tb.Settings{
		// You can also set custom API URL.
		// If field is empty it equals to "https://api.telegram.org".
		// URL: "http://195.129.111.17:8012",
		Token:  "1857698955:AAFuo00nKY0zYd9bVC0jeL5LiydJl8puoK0",
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		log.Fatal(err)
		return
	}

	b.Handle("/hello", func(m *tb.Message) {
		switch m.Payload {
		case "bitch":
			var cfg tb.Album
			for _, v := range getpicuris(4) {
				cfg = append(cfg, v)
			}
			_, err = b.SendAlbum(m.Chat, cfg)
			if err != nil {
				fmt.Println(err.Error())
			}
		default:
			_, err = b.Send(m.Chat, getpicuris(1)[0])
			if err != nil {
				fmt.Println(err.Error())
			}
		}
	})

	b.Handle("/count", func(m *tb.Message) {
		_, body, _ := fasthttp.Get(nil, domain+"/getcount")
		_, err = b.Send(m.Chat, string(body))
		if err != nil {
			fmt.Println(err.Error())
		}
	})
	b.Handle(tb.OnText, func(m *tb.Message) {
		switch m.Text {
		case "":

		default:
			_, err = b.Send(m.Chat, getpicuris(1)[0])
			if err != nil {
				fmt.Println(err.Error())
			}
		}
	})

	b.Handle(tb.OnPhoto, func(m *tb.Message) {
		// photos only
	})

	b.Handle(tb.OnChannelPost, func(m *tb.Message) {
		// channel posts only
	})

	b.Handle(tb.OnQuery, func(q *tb.Query) {
		// incoming inline queries
	})

	b.Start()
}
