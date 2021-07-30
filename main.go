package main

import (
	"encoding/json"
	"fmt"
	"log"
	rand_ "math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"telegram-bot/weather"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/valyala/fasthttp"
	tb "gopkg.in/tucnak/telebot.v2"
)

var domain = "https://imgs.zhxiao1124.cn"

type piclist struct {
	Path     string `json:"path"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Format   string `json:"format"`
	MIMEType string `json:"mimeType"`
	Width    int64  `json:"width"`
	Height   int64  `json:"height"`
}

var pics []piclist

const ChatID = "-1001117396121"

var c = cron.New()
var b *tb.Bot

func init() {
	rand_.Seed(time.Now().UnixNano())
}

func GetPics() error {
	statusCode, body, err := fasthttp.Get(nil, domain+"/get?type=jpg")
	if err != nil {
		log.Println(err.Error())
		return err
	}
	if statusCode == http.StatusOK {
		if err := json.Unmarshal(body, &pics); err != nil {
			log.Println(err.Error())
			return err
		}
	}
	rand_.Shuffle(len(pics), func(i, j int) {
		pics[i], pics[j] = pics[j], pics[i]
	})
	return nil
}

func getpicuris(count int) []*tb.Photo {
	var res []*tb.Photo
	if len(pics) == 0 {
		GetPics()
	}
	for i := 0; i < count; i++ {
		uri := fmt.Sprintf(domain+"/%s?size=500", strings.TrimPrefix(pics[0].Path, "./"))
		res = append(res, &tb.Photo{File: tb.FromURL(uri)})
		pics = pics[1:]
		log.Println(uri)
	}
	return res
}

func main() {
	var err error
	b, err = tb.NewBot(tb.Settings{
		// You can also set custom API URL.
		// If field is empty it equals to "https://api.telegram.org".
		// URL: "http://195.129.111.17:8012",
		Token:  "1857698955:AAFuo00nKY0zYd9bVC0jeL5LiydJl8puoK0",
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	//curl -F "url=" https://api.telegram.org/bot1857698955:AAFuo00nKY0zYd9bVC0jeL5LiydJl8puoK0/setWebhook
	if err != nil {
		log.Fatal(err)
		return
	}
	var (
		menu    = &tb.ReplyMarkup{ResizeReplyKeyboard: true}
		hello   = menu.Text("/hello")
		reset   = menu.Text("/reset")
		count   = menu.Text("/count")
		cache   = menu.Text("/cache")
		Weather = menu.Text("/weather")
	)
	menu.Reply(
		menu.Row(hello),
		menu.Row(reset),
		menu.Row(count),
		menu.Row(cache),
		menu.Row(Weather),
	)

	b.Handle("/start", func(m *tb.Message) {
		if !m.Private() {
			return
		}
		b.Send(m.Sender, "Hello!", menu)
	})

	b.Handle(&hello, func(m *tb.Message) {
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

	b.Handle(&reset, func(m *tb.Message) {
		err := GetPics()
		if err != nil {
			_, err = b.Send(m.Chat, fmt.Sprintf("重新拉取失败:%s", err.Error()))
		} else {
			_, err = b.Send(m.Chat, fmt.Sprintf("重新拉取了%d张图片", len(pics)))
		}
		if err != nil {
			fmt.Println(err.Error())
		}
	})

	b.Handle(&count, func(m *tb.Message) {
		_, body, _ := fasthttp.Get(nil, domain+"/count")
		_, err = b.Send(m.Chat, string(body))
		if err != nil {
			fmt.Println(err.Error())
		}
	})

	b.Handle(&cache, func(m *tb.Message) {
		_, err = b.Send(m.Chat, fmt.Sprintf("当前缓存区，还有%d张图片", len(pics)))
		if err != nil {
			fmt.Println(err.Error())
		}
	})

	b.Handle(&Weather, func(m *tb.Message) {
		code := 101250105
		var err error
		if strings.TrimSpace(m.Payload) != "" {
			_code, err := strconv.Atoi(m.Payload)
			if err != nil {
				if _, err = b.Send(m.Chat, "错误的城市编码"); err != nil {
					log.Println(err.Error())
				}
				_code = 101250105
			}
			code = _code
		}

		data := weather.Get(strconv.Itoa(code))
		if data == nil {
			if _, err = b.Send(m.Chat, "查询失败"); err != nil {
				log.Println(err.Error())
			}
			return
		}
		md := fmt.Sprintf("地区:%s\n时间:%s %s\n气温:%s℃\n相对湿度:%s\n天气:%s\n风向:%s\n风速:%s\npm2\\.5:%s",
			data.Cityname, strings.ReplaceAll(strings.ReplaceAll(data.Date, "(", "\\("), ")", "\\)"), data.Time, data.Temp, data.SD, data.Weather, data.Wd, data.Ws, data.AqiPm25)

		if _, err = b.Send(m.Chat, md, &tb.SendOptions{ParseMode: tb.ModeMarkdownV2}); err != nil {
			log.Println(err.Error())
		}
	})

	b.Handle("/help", func(m *tb.Message) {
		_, err = b.Send(m.Chat, fmt.Sprintf("%s bitch\n%s\n%s\n%s\n%s", hello.Text, reset.Text, count.Text, cache.Text, Weather.Text))
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

	c.AddJob("@hourly", sendPic{Num: 4, b: b, chatId: ChatID})
	c.AddJob("30 18 * * 1-5", sendPic{Num: 6, b: b, chatId: ChatID})
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

type sendPic struct {
	Num    int
	b      *tb.Bot
	chatId string
}

func (s sendPic) Run() {
	c_, err := s.b.ChatByID(s.chatId)
	if err != nil {
		log.Fatalln(err.Error())
	}
	var cfg tb.Album
	for _, v := range getpicuris(s.Num) {
		cfg = append(cfg, v)
	}
	_, err = s.b.SendAlbum(c_, cfg)
	if err != nil {
		fmt.Println(err.Error())
	}
}
