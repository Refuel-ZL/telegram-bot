package main

import (
	"encoding/json"
	"fmt"
	"log"
	rand_ "math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/valyala/fasthttp"
	tb "gopkg.in/tucnak/telebot.v2"
)

var domain = "https://imgs.zhxiao1124.design"
var pics []string

const ChatID = "-380625611"

var c = cron.New(cron.WithSeconds())

func init() {
	rand_.Seed(time.Now().UnixNano())
}

func GetPics() {
	statusCode, body, err := fasthttp.Get(nil, domain+"/get")
	if err != nil {
		log.Println(err.Error())
		return
	}
	if statusCode == http.StatusOK {
		if err := json.Unmarshal(body, &pics); err != nil {
			log.Println(err.Error())
			return
		}
	}
	rand_.Shuffle(len(pics), func(i, j int) {
		pics[i], pics[j] = pics[j], pics[i]
	})
}

func getpicuris(count int) []*tb.Photo {
	var res []*tb.Photo
	if len(pics) == 0 {
		GetPics()
	}
	for i := 0; i < count; i++ {
		uri := fmt.Sprintf("http://imgs.zhxiao1124.design/static/%s", pics[0])
		res = append(res, &tb.Photo{File: tb.FromURL(uri)})
		pics = pics[1:]
		log.Println(uri)
	}
	return res
}

func main() {
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

	c.AddFunc("0 0 * * * ?", func() {
		c_, err := b.ChatByID(ChatID)
		if err != nil {
			log.Fatalln(err.Error())
		}
		var cfg tb.Album
		for _, v := range getpicuris(4) {
			cfg = append(cfg, v)
		}
		_, err = b.SendAlbum(c_, cfg)
		if err != nil {
			fmt.Println(err.Error())
		}
	})
	c.Start()
	go b.Start()
	fmt.Println("程序启动")

	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt, syscall.SIGTERM)
	<-s
	// 收到信号退出无限循环
	b.Stop()
	fmt.Println("\ngraceful shuwdown")
}

type TimingTask struct {
	Bot    *tb.Bot
	ChatID string
	Func   func(TimingTask)
}

func (t TimingTask) Run() {
	t.Func(t)
}
