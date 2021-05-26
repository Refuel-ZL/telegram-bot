package weather

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/valyala/fasthttp"
)

type WeatherModel struct {
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

func Get(city string) *WeatherModel {
	statusCode, body, err := fasthttp.Get(nil, fmt.Sprintf("http://d1.weather.com.cn/sk_2d/%s.html?_=%d", city, time.Now().Unix()))
	if err != nil {
		log.Println(err.Error())
		return nil
	}
	if statusCode != http.StatusOK {
		return nil
	}
	body = bytes.TrimPrefix(body, []byte("var dataSK="))
	var data WeatherModel
	if err := json.Unmarshal(body, &data); err != nil {
		return nil
	}
	return &data
}
