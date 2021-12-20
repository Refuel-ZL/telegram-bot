package weather

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"net/http"
	"strconv"
	"time"

	browser "github.com/eddycjy/fake-useragent"
)

type WeatherResponse struct {
	Nameen         string `json:"nameen"`
	Cityname       string `json:"cityname"`
	City           string `json:"city"`
	Temp           string `json:"temp"`
	Tempf          string `json:"tempf"`
	Wd             string `json:"WD"`
	Wde            string `json:"wde"`
	Ws             string `json:"WS"`
	Wse            string `json:"wse"`
	SD             string `json:"SD"`
	WeatherModelSD string `json:"sd"`
	Qy             string `json:"qy"`
	Njd            string `json:"njd"`
	Time           string `json:"time"`
	Rain           string `json:"rain"`
	Rain24H        string `json:"rain24h"`
	Aqi            string `json:"aqi"`
	AqiPm25        string `json:"aqi_pm25"`
	Weather        string `json:"weather"`
	Weathere       string `json:"weathere"`
	Weathercode    string `json:"weathercode"`
	Limitnumber    string `json:"limitnumber"`
	Date           string `json:"date"`
}

func Get(city string) (*WeatherResponse, error) {
	url := fmt.Sprintf("http://d1.weather.com.cn/sk_2d/%s.html", city)
	client := resty.New()

	client.SetAllowGetMethodPayload(true)
	resp, err := client.R().
		SetQueryParams(map[string]string{
			"_": strconv.FormatInt(time.Now().Unix(), 10),
		}).
		SetHeaders(map[string]string{
			"Referer":    "http://www.weather.com.cn/",
			"user-agent": browser.Random(),
		}).
		SetBody(`{"request":"test"}`).
		Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("请求失败，%d", resp.StatusCode())
	}
	body := resp.Body()
	body = bytes.TrimPrefix(body, []byte("var dataSK="))
	var data WeatherResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}
	return &data, nil
}
