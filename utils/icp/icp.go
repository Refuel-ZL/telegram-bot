package icp

import (
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"gocv.io/x/gocv"
	"net/http"
	"strconv"
	"time"
)

type BeiAnResponse struct {
	Total    int
	UnitName string
	List     []*Info
}

type Info struct {
	ContentTypeName  string `json:"contentTypeName"`
	Domain           string `json:"domain"`
	DomainId         int64  `json:"domainId"`
	HomeUrl          string `json:"homeUrl"`
	LeaderName       string `json:"leaderName"`
	LimitAccess      string `json:"limitAccess"`
	MainId           int    `json:"mainId"`
	MainLicence      string `json:"mainLicence"`
	NatureName       string `json:"natureName"`
	ServiceId        int    `json:"serviceId"`
	ServiceLicence   string `json:"serviceLicence"`
	ServiceName      string `json:"serviceName"`
	UnitName         string `json:"unitName"`
	UpdateRecordTime string `json:"updateRecordTime"`
}

func BeiAn(unitName string) (beianResponse *BeiAnResponse, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("崩溃:%v", e)
		}
	}()
	client := resty.New()
	// 使用代理
	// client.SetTimeout(15 * time.Second)
	// SetProxy("socks5://127.0.0.1:1080")

	// 获取cookie
	var cookieHeaders = map[string]string{
		"accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
		"accept-encoding": "gzip, deflate, br",
		"accept-language": "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6",
		"user-agent":      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.71 Safari/537.36",
	}
	resp, err := client.R().
		SetHeaders(cookieHeaders).
		Get("https://beian.miit.gov.cn/")
	if err != nil {
		return nil, fmt.Errorf("请求cookie失败:%w", err)
	}
	var cookie *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == "__jsluid_s" {
			cookie = c
		}
	}
	if cookie == nil {
		return nil, errors.New("获取cookie失败")
	}

	// 	构造authKey
	authAccount := "test"
	authSecret := "test"
	timeStamp := time.Now().UnixNano() / 1e6
	var authKey = toMd5string(fmt.Sprintf("%s%s%d", authAccount, authSecret, timeStamp))

	// 请求获取token
	tUrl := "https://hlwicpfwc.miit.gov.cn/icpproject_query/api/auth"
	tHeader := map[string]string{
		"Host":             "hlwicpfwc.miit.gov.cn",
		"Connection":       "keep-alive",
		"sec-ch-ua":        `" Not A;Brand";v="99", "Chromium";v="90", "Microsoft Edge";v="90"`,
		"Accept":           "*/*",
		"DNT":              "1",
		"sec-ch-ua-mobile": "?0",
		"User-Agent":       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.71 Safari/537.36",
		"Origin":           "https://beian.miit.gov.cn",
		"Sec-Fetch-Site":   "same-site",
		"Sec-Fetch-Mode":   "cors",
		"Sec-Fetch-Dest":   "empty",
		"Referer":          "https://beian.miit.gov.cn/",
		"Accept-Encoding":  "gzip, deflate, br",
		"Accept-Language":  "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6",
	}
	var tbody struct {
		Code   int    `json:"code"`
		Msg    string `json:"msg"`
		Params struct {
			Bussiness string `json:"bussiness"`
			Expire    int    `json:"expire"`
			Refresh   string `json:"refresh"`
		} `json:"params"`
		Success bool `json:"success"`
	}
	tResp, err := client.R().
		SetHeaders(tHeader).
		SetFormData(map[string]string{
			"authKey":   authKey,
			"timeStamp": strconv.Itoa(int(timeStamp)),
		}).
		SetCookie(cookie).
		SetResult(&tbody).
		ForceContentType("application/json").
		Post(tUrl)
	if err != nil {
		return nil, fmt.Errorf("获取token http失败:%w", err)
	}
	if tResp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("小黑屋欢迎你。")
	}
	if tbody.Params.Bussiness == "" {
		return nil, fmt.Errorf("获取token失败:%s", tbody.Msg)
	}

	// 获取验证图像,UUID
	pUrl := "https://hlwicpfwc.miit.gov.cn/icpproject_query/api/image/getCheckImage"
	pHeader := map[string]string{
		"Host":             "hlwicpfwc.miit.gov.cn",
		"Connection":       "keep-alive",
		"Content-Length":   "0",
		"Accept":           "application/json, text/plain, */*",
		"DNT":              "1",
		"sec-ch-ua-mobile": "?0",
		"User-Agent":       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.71 Safari/537.36",
		"token":            tbody.Params.Bussiness,
		"Origin":           "https://beian.miit.gov.cn",
		"Sec-Fetch-Site":   "same-site",
		"Sec-Fetch-Mode":   "cors",
		"Sec-Fetch-Dest":   "empty",
		"Referer":          "https://beian.miit.gov.cn/",
		"Accept-Encoding":  "gzip, deflate, br",
		"Accept-Language":  "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6",
	}
	var pBody struct {
		Code   int    `json:"code"`
		Msg    string `json:"msg"`
		Params struct {
			BigImage   string `json:"bigImage"`
			Height     string `json:"height"`
			SmallImage string `json:"smallImage"`
			Uuid       string `json:"uuid"`
		} `json:"params"`
		Success bool `json:"success"`
	}
	pResp, err := client.R().
		SetHeaders(pHeader).
		SetCookie(cookie).
		SetResult(&pBody).
		Post(pUrl)
	if err != nil {
		return nil, fmt.Errorf("请求验证图失败:%w", err)
	}
	if pResp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("小黑屋欢迎你。")
	}
	if pBody.Code != 200 {
		return nil, fmt.Errorf("请求验证图失败:%s", pBody.Msg)
	}
	imgSrc, err := toBase64Img(pBody.Params.BigImage)
	if err != nil {
		return nil, fmt.Errorf("解析背景图失败:%w", err)
	}
	defer func(imgSrc *gocv.Mat) {
		_ = imgSrc.Close()
	}(imgSrc)

	imgTempl, err := toBase64Img(pBody.Params.SmallImage)
	if err != nil {
		return nil, fmt.Errorf("解析模块图失败:%w", err)
	}
	defer func(imgTempl *gocv.Mat) {
		_ = imgTempl.Close()
	}(imgTempl)

	result := gocv.NewMat()
	defer func(result *gocv.Mat) {
		_ = result.Close()
	}(&result)

	m := gocv.NewMat()
	gocv.MatchTemplate(*imgTempl, *imgSrc, &result, gocv.TmCcoeffNormed, m)
	defer func(m *gocv.Mat) {
		_ = m.Close()
	}(&m)

	_, maxConfidence, _, maxLoc := gocv.MinMaxLoc(result)
	fmt.Println(maxConfidence)
	if maxConfidence < 0.8 {
		return nil, fmt.Errorf("图片验证可靠性过低:%f", maxConfidence)
	}

	// 通过拼图验证，获取sign
	cUrl := "https://hlwicpfwc.miit.gov.cn/icpproject_query/api/image/checkImage"
	cHeaders := map[string]string{
		"Host":             "hlwicpfwc.miit.gov.cn",
		"Accept":           "application/json, text/plain, */*",
		"Connection":       "keep-alive",
		"Content-Length":   "60",
		"sec-ch-ua":        `" Not A;Brand";v="99", "Chromium";v="90", "Microsoft Edge";v="90"`,
		"DNT":              "1",
		"sec-ch-ua-mobile": "?0",
		"User-Agent":       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.72 Safari/537.36 Edg/90.0.818.42",
		"token":            tbody.Params.Bussiness,
		"Content-Type":     "application/json",
		"Origin":           "https://beian.miit.gov.cn",
		"Sec-Fetch-Site":   "same-site",
		"Sec-Fetch-Mode":   "cors",
		"Sec-Fetch-Dest":   "empty",
		"Referer":          "https://beian.miit.gov.cn/",
		"Accept-Encoding":  "gzip, deflate, br",
		"Accept-Language":  "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6",
	}
	var cBody struct {
		Code    int    `json:"code"`
		Msg     string `json:"msg"`
		Params  string `json:"params"`
		Success bool   `json:"success"`
	}
	cResp, err := client.R().
		SetHeaders(cHeaders).
		SetCookie(cookie).
		SetBody(fmt.Sprintf(`{"key":"%s","value":"%d"}`, pBody.Params.Uuid, maxLoc.X+1)).
		SetResult(&cBody).
		ForceContentType("application/json").
		Post(cUrl)
	if err != nil {
		return nil, fmt.Errorf("验证图片失败:%w", err)
	}
	if cResp.StatusCode() != http.StatusOK || cBody.Code != 200 || cBody.Params == "" {
		return nil, fmt.Errorf("验证图片响应失败:%s", cBody.Msg)
	}

	// 获取最终的备案信息
	iUrl := "https://hlwicpfwc.miit.gov.cn/icpproject_query/api/icpAbbreviateInfo/queryByCondition"
	iHeaders := map[string]string{
		"Host":             "hlwicpfwc.miit.gov.cn",
		"Connection":       "keep-alive",
		"Content-Length":   "78",
		"sec-ch-ua":        `" Not A;Brand";v="99", "Chromium";v="90", "Microsoft Edge";v="90"`,
		"DNT":              "1",
		"sec-ch-ua-mobile": "?0",
		"User-Agent":       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.72 Safari/537.36 Edg/90.0.818.42",
		"Content-Type":     "application/json",
		"Accept":           "application/json, text/plain, */*",
		"uuid":             pBody.Params.Uuid,
		"token":            tbody.Params.Bussiness,
		"sign":             cBody.Params,
		"Origin":           "https://beian.miit.gov.cn",
		"Sec-Fetch-Site":   "same-site",
		"Sec-Fetch-Mode":   "cors",
		"Sec-Fetch-Dest":   "empty",
		"Referer":          "https://beian.miit.gov.cn/",
		"Accept-Encoding":  "gzip, deflate, br",
		"Accept-Language":  "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6",
	}

	var param struct {
		PageNum  string `json:"pageNum"`
		PageSize string `json:"pageSize"`
		UnitName string `json:"unitName"`
	}
	param.UnitName = unitName

	var iBody struct {
		Code   int    `json:"code"`
		Msg    string `json:"msg"`
		Params struct {
			EndRow           int    `json:"endRow"`
			FirstPage        int    `json:"firstPage"`
			HasNextPage      bool   `json:"hasNextPage"`
			HasPreviousPage  bool   `json:"hasPreviousPage"`
			IsFirstPage      bool   `json:"isFirstPage"`
			IsLastPage       bool   `json:"isLastPage"`
			LastPage         int    `json:"lastPage"`
			List             []Info `json:"list"`
			NavigatePages    int    `json:"navigatePages"`
			NavigatepageNums []int  `json:"navigatepageNums"`
			NextPage         int    `json:"nextPage"`
			PageNum          int    `json:"pageNum"`
			PageSize         int    `json:"pageSize"`
			Pages            int    `json:"pages"`
			PrePage          int    `json:"prePage"`
			Size             int    `json:"size"`
			StartRow         int    `json:"startRow"`
			Total            int    `json:"total"`
		} `json:"params"`
		Success bool `json:"success"`
	}
	iResp, err := client.R().
		SetHeaders(iHeaders).
		SetCookie(cookie).
		SetBody(param).
		SetResult(&iBody).
		ForceContentType("application/json").
		Post(iUrl)
	if err != nil {
		return nil, fmt.Errorf("获取备案结果失败:%w", err)
	}
	if iResp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("获取备案结果被拒绝: http状态码%d", iResp.StatusCode())
	}
	if iBody.Code != 200 {
		return nil, fmt.Errorf("获取备案结果失败:%s", iBody.Msg)
	}
	beianResponse = new(BeiAnResponse)
	beianResponse.UnitName = unitName
	for i := 0; i < iBody.Params.Pages; i++ {
		for _, item := range iBody.Params.List {
			beianResponse.Total++
			beianResponse.List = append(beianResponse.List, &item)
		}
		param.PageSize = "10"
		param.PageNum = strconv.Itoa(iBody.Params.PageNum + 1)
		if iBody.Params.PageNum >= iBody.Params.Pages {
			break
		}
		time.Sleep(time.Second)
		iResp, err = client.R().
			SetHeaders(iHeaders).
			SetCookie(cookie).
			SetBody(param).
			SetResult(&iBody).
			ForceContentType("application/json").
			Post(iUrl)
		if err != nil {
			return nil, fmt.Errorf("获取备案结果失败:%w", err)
		}
		if iResp.StatusCode() != http.StatusOK {
			return nil, fmt.Errorf("获取备案结果失败:%d", iResp.StatusCode())
		}
	}
	return beianResponse, nil
}

func toMd5string(in string) string {
	h := md5.New()
	h.Write([]byte(in))
	re := h.Sum(nil)
	return fmt.Sprintf("%x", re)
}

func toBase64Img(in string) (*gocv.Mat, error) {
	ddd, _ := base64.StdEncoding.DecodeString(in)
	imgSrc, err := gocv.IMDecode(ddd, gocv.IMReadGrayScale)
	if err != nil {
		return nil, err
	}
	if imgSrc.Empty() {
		return nil, errors.New("invalid read")
	}
	return &imgSrc, nil
}
