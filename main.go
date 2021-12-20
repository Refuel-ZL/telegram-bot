package main

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"log"
	rand_ "math/rand"
	"os"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"syscall"
	"telegram-bot/utils/icp"
	"telegram-bot/utils/weather"
	"time"

	"golang.org/x/time/rate"

	"github.com/disintegration/imaging"
	"github.com/robfig/cron/v3"
	tb "gopkg.in/tucnak/telebot.v2"
)

var pics []string

const ChatID = "-1001117396121"

var c = cron.New()
var b *tb.Bot

func init() {
	rand_.Seed(time.Now().UnixNano())
}

func GetPics() error {
	pics = pics[0:0]
	err := GetAllFile("./dist", &pics)
	if err != nil {
		return err
	}
	rand_.Shuffle(len(pics), func(i, j int) {
		pics[i], pics[j] = pics[j], pics[i]
	})
	return nil
}

func getpicuris(count int) []*tb.Photo {
	var res []*tb.Photo
	if len(pics) == 0 || len(pics) < count {
		GetPics()
	}
	for i := 0; i < count; i++ {
		var src string
		for {
			if len(pics) == 0 {
				if err := GetPics(); err != nil {
					return nil
				}
			}
			src = pics[0]
			pics = pics[1:]
			if info, err := os.Stat(src); err == nil {
				if info.Size() < 5*1000*1000 {
					break
				}
			}
		}
		var fs tb.File = tb.FromDisk(src)
		res = append(res, &tb.Photo{File: fs})
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
		beian   = menu.Text("/icp")
	)
	menu.Reply(
		menu.Row(hello),
		menu.Row(reset),
		menu.Row(count),
		menu.Row(cache),
		menu.Row(Weather),
		menu.Row(beian),
	)

	b.Handle("/start", func(m *tb.Message) {
		if !m.Private() {
			return
		}
		_, err := b.Send(m.Sender, "Hello!", menu)
		if err != nil {
			fmt.Println(err.Error())
		}
	})

	r := rate.Every(70 * time.Second)
	n := 4
	limit := rate.NewLimiter(r, 20/n-1)

	b.Handle(&hello, func(m *tb.Message) {
		switch m.Payload {
		case "bitch":
			// if strconv.FormatInt(m.Chat.ID, 10) != ChatID || limit.Allow() {
			if limit.Allow() {
				fmt.Println(time.Now().Format(time.RFC3339))
				var cfg tb.Album
				var _pics = getpicuris(n)
				for _, v := range _pics {
					cfg = append(cfg, v)
				}
				_, err = b.SendAlbum(m.Chat, cfg)
				if err != nil {
					_log := fmt.Sprintf("异常:%s\n", err.Error())
					fmt.Print(_log)
					_, _ = b.Send(m.Chat, fmt.Sprintf("发送失败:%s", err.Error()))
				}
			} else {
				_, err = b.Send(m.Chat, "哇噻!!! 你发得好快哟")
				if err != nil {
					fmt.Printf("异常:%s\n", err.Error())
				}
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
		var list []string
		GetAllFile("./dist", &list)
		_, err = b.Send(m.Chat, fmt.Sprintf("一共有%d张图片", len(list)))
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

	b.Handle(&beian, func(m *tb.Message) {
		name := strings.TrimSpace(m.Payload)
		if name == "" {
			_, err = b.Send(m.Chat, "查询失败")
			if err != nil {
				log.Println(err.Error())
			}
			return
		}
		resp, err := icp.BeiAn(name)
		if err != nil {
			_, err = b.Send(m.Chat, "查询失败:%s", err.Error())
			if err != nil {
				log.Println(err.Error())
			}
			return
		}
		md := fmt.Sprintf("查询：%s 共%d条记录：\n\n", resp.UnitName, resp.Total)
		for _, item := range resp.List {
			md += fmt.Sprintf("域名主办方：%s\n", item.UnitName)
			md += fmt.Sprintf("域名：%s\n", item.Domain)
			md += fmt.Sprintf("域名类型：%s\n", item.NatureName)
			md += fmt.Sprintf("网站名称：%s\n", item.ServiceName)
			md += fmt.Sprintf("备案许可证号：%s\n", item.MainLicence)
			md += fmt.Sprintf("网站备案号：%s\n", item.ServiceLicence)
			md += fmt.Sprintf("网站前置审批项：%s\n", item.ContentTypeName)
			md += fmt.Sprintf("是否限制接入：%s\n", item.LimitAccess)
			md += fmt.Sprintf("审核通过日期：%s\n", item.UpdateRecordTime)
			md += fmt.Sprintf("\n")
		}
		md += "查询完毕"
		_, err = b.Send(m.Chat, md, &tb.SendOptions{ParseMode: tb.ModeHTML})
		if err != nil {
			log.Println(err.Error())
		}
	})

	b.Handle("/weather", func(m *tb.Message) {
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

		data, err := weather.Get(strconv.Itoa(code))
		if err != nil {
			_, err = b.Send(m.Chat, "查询失败")
			if err != nil {
				log.Println(err.Error())
			}
			return
		}
		md := fmt.Sprintf("地区:%s\n时间:%s %s\n气温:%s℃\n相对湿度:%s\n天气:%s\n风向:%s\n风速:%s\npm2\\.5:%s",
			data.Cityname, strings.ReplaceAll(strings.ReplaceAll(data.Date, "(", "\\("), ")", "\\)"), data.Time, data.Temp, data.SD, data.Weather, data.Wd, data.Ws, data.AqiPm25)
		_, err = b.Send(m.Chat, md, &tb.SendOptions{ParseMode: tb.ModeMarkdownV2})
		if err != nil {
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
		case "我爱学习", "来点涩图":
			var cfg tb.Album
			var _pics = getpicuris(4)
			for _, v := range _pics {
				cfg = append(cfg, v)
			}
			_, err = b.SendAlbum(m.Chat, cfg)
			if err != nil {
				fmt.Println(err.Error())
			}
		default:
		}
	})

	b.Handle(tb.OnPhoto, func(m *tb.Message) {
		// photos only
	})

	b.Handle(tb.OnChannelPost, func(m *tb.Message) {
		// channel posts only
	})

	b.Handle(tb.OnQuery, func(m *tb.Message) {
		// incoming inline queries
	})

	c.AddJob("0 9-17 * * *", sendPic{Num: 9, b: b, chatId: ChatID})
	c.AddJob("30 18 * * *", sendPic{Num: 9, b: b, chatId: ChatID})
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
	err = Retry(3, 500, func() error {
		_, _err := s.b.SendAlbum(c_, cfg)
		return _err
	})
	if err != nil {
		_log := fmt.Sprintf("异常:%s\n", err.Error())
		fmt.Print(_log)
	}
}

func Retry(attempts int, sleep time.Duration, f func() error) (err error) {
	for i := 0; ; i++ {
		err = f()
		if err == nil {
			return
		}
		if i >= (attempts - 1) {
			break
		}
		time.Sleep(sleep)
	}
	return fmt.Errorf("after %d attempts, last error: %v", attempts, err)
}

func Size(r io.Reader, size int) (out *bytes.Buffer, err error) {
	src, _, err := image.Decode(r)
	if err != nil {
		return nil, err
	}
	src = imaging.Resize(src, size, 0, imaging.Lanczos)
	out = &bytes.Buffer{}
	err = imaging.Encode(out, src, imaging.JPEG)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func GetAllFile(pathname string, list *[]string) error {
	rd, err := ioutil.ReadDir(pathname)
	for _, fi := range rd {
		if fi.IsDir() {
			GetAllFile(path.Join(pathname, fi.Name()), list)
		} else {
			*list = append(*list, path.Join(pathname, fi.Name()))
		}
	}
	return err
}
