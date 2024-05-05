package services

import (
	"bilibo/config"
	"bilibo/consts"
	"bilibo/log"
	"bilibo/universal"
	"bilibo/utils"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/skip2/go-qrcode"
	"github.com/tidwall/gjson"
)

type client struct {
	cookies []*http.Cookie
	timeout time.Duration
	logger  resty.Logger
	ua      string
	imgKey  string
	subKey  string
	mid     int
	face    string
	uname   string
}

func (c *client) resty() *resty.Client {
	client := resty.New().SetTimeout(c.timeout).SetHeader("user-agent", c.ua)
	if c.logger != nil {
		client.SetLogger(c.logger)
	}
	if c.cookies != nil {
		client.SetCookies(c.cookies)
	}
	return client
}

func getRespData(resp *resty.Response, prefix string) ([]byte, int64, error) {
	var errorCode int64 = 0
	if resp.StatusCode() != 200 {
		respCode := resp.StatusCode()
		errorCode, _ = strconv.ParseInt(fmt.Sprintf("%d%d", 999, respCode), 10, 64)
		return nil, errorCode, errors.New(prefix + "失败，status code: " + strconv.Itoa(resp.StatusCode()))
	}
	if !gjson.ValidBytes(resp.Body()) {
		errorCode = 999
		return nil, errorCode, errors.New("json解析失败：" + resp.String())
	}
	res := gjson.ParseBytes(resp.Body())
	code := res.Get("code").Int()
	if code != 0 {
		return nil, code, errors.New(prefix + "失败，返回值：" + strconv.FormatInt(code, 10))
	}
	return []byte(res.Get("data").Raw), errorCode, nil
}

func (c *client) setCookies(cookies []*http.Cookie) {
	c.cookies = cookies
}

func (c *client) GetCookiesString() string {
	var cookieStrings []string
	for _, cookie := range c.cookies {
		cookieStrings = append(cookieStrings, cookie.String())
	}
	return strings.Join(cookieStrings, "\n")
}

func (c *client) GetMid() int {
	return c.mid
}

func (c *client) GetWbi() (string, string) {
	return c.imgKey, c.subKey
}

type WbiImg struct {
	ImgUrl string `json:"img_url"`
	SubUrl string `json:"sub_url"`
}

type navigation struct {
	Face   string `json:"face"`
	Mid    int    `json:"mid"`
	Uname  string `json:"uname"`
	WbiImg WbiImg `json:"wbi_img"`
}

func (c *client) getNavigation() (*navigation, int64, error) {
	resp, err := c.resty().R().Get("https://api.bilibili.com/x/web-interface/nav")
	if err != nil {
		return nil, 0, err
	}
	data, errorCode, err := getRespData(resp, "导航栏用户信息")
	if err != nil {
		return nil, errorCode, err
	}
	var ret *navigation
	err = json.Unmarshal(data, &ret)
	if err == nil {
		c.imgKey = ret.WbiImg.ImgUrl
		c.subKey = ret.WbiImg.SubUrl
		c.mid = ret.Mid
		c.face = ret.Face
		c.uname = ret.Uname
	}
	return ret, errorCode, err
}

func new() (*client, *qrCode, error) {
	c := &client{
		ua:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.90 Safari/537.36",
		cookies: []*http.Cookie{},
		timeout: 20 * time.Second,
		logger:  nil,
	}

	qr, err := c.getQRCode()
	if err != nil {
		return nil, nil, err
	}
	return c, qr, nil
}

type qrCode struct {
	Url       string `json:"url"`        // 二维码内容url
	QrcodeKey string `json:"qrcode_key"` // 扫码登录秘钥
}

func (result *qrCode) Encode() ([]byte, error) {
	return qrcode.Encode(result.Url, qrcode.Medium, 256)
}

func (c *client) getQRCode() (*qrCode, error) {
	resp, err := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").
		Get("https://passport.bilibili.com/x/passport-login/web/qrcode/generate")
	if err != nil {
		return nil, err
	}
	data, _, err := getRespData(resp, "申请二维码")
	if err != nil {
		return nil, err
	}
	var result *qrCode
	err = json.Unmarshal(data, &result)
	return result, err
}

type Info struct {
	Cookies string
	Mid     int
	UName   string
	Face    string
	ImgKey  string
	SubKey  string
}

func (c *client) loginWithQRCode(qrCode *qrCode) (*Info, error) {
	logger := log.GetLogger()
	if qrCode == nil {
		return nil, errors.New("请先获取二维码")
	}
	for {
		ok, err := c.qrCodeSuccess(qrCode)
		if err != nil {
			logger.Info("qrCodeSuccess")
			logger.Info(err)
			return nil, err
		}
		if ok {
			if _, _, err := c.getNavigation(); err != nil {
				logger.Info("getNavigation")
				logger.Info(err)
				return nil, err
			} else {
				return &Info{
					Cookies: c.GetCookiesString(),
					Mid:     c.mid,
					UName:   c.uname,
					Face:    c.face,
					ImgKey:  c.imgKey,
					SubKey:  c.subKey,
				}, nil
			}
		}
		time.Sleep(3 * time.Second) // 主站 3s 一次请求
	}
}

func (c *client) qrCodeSuccess(qrCode *qrCode) (bool, error) {
	resp, err := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetQueryParam("qrcode_key", qrCode.QrcodeKey).Get("https://passport.bilibili.com/x/passport-login/web/qrcode/poll")
	if err != nil {
		return false, err
	}
	if resp.StatusCode() != 200 {
		return false, errors.New("登录bilibili失败")
	}
	if !gjson.ValidBytes(resp.Body()) {
		return false, errors.New("json invalid: " + resp.String())
	}
	result := gjson.ParseBytes(resp.Body())
	retCode := result.Get("code").Int()
	if retCode != 0 {
		return false, errors.New("登录bilibili失败，错误码：" + strconv.FormatInt(retCode, 10) + "，错误信息：" + gjson.GetBytes(resp.Body(), "message").String())
	} else {
		codeValue := result.Get("data.code")
		if !codeValue.Exists() || codeValue.Type != gjson.Number {
			return false, errors.New("扫码登录未成功，返回异常")
		}
		code := codeValue.Int()
		switch code {
		case 86038: // 二维码已失效
			return false, errors.New("扫码登录未成功，原因：二维码已失效")
		case 86090: // 二维码已扫码未确认
			return false, nil
		case 86101: // 未扫码
			return false, nil
		case 0:
			c.setCookies(resp.Cookies())
			return true, nil
		default:
			return false, errors.New("由于未知原因，扫码登录未成功，错误码：" + strconv.FormatInt(code, 10))
		}
	}
}

func SetAccountInfo() (string, int64, error) {
	c, qr, err := new()
	if err != nil {
		return "", 0, err
	}

	qrImgByte, err := qr.Encode()
	if err != nil {
		return "", 0, err
	}

	conf := config.GetConfig()
	qrId := time.Now().UnixNano()
	fileName := fmt.Sprintf("%d.png", qrId)
	filePath := filepath.Join(conf.Download.Path, ".tmp", fileName)
	err = os.WriteFile(filePath, qrImgByte, os.ModePerm)
	if err != nil {
		return "", 0, err
	}

	url := "/api/account/qrcode/" + fileName
	AddQRCodeInfo(fmt.Sprintf("%d", qrId))
	go func() {
		if info, err := c.loginWithQRCode(qr); err == nil {
			SaveAccountInfo(
				info.Mid,
				info.UName, info.Face,
				c.GetCookiesString(),
				info.ImgKey, info.SubKey,
			)
			SetQRCodeStatus(fmt.Sprintf("%d", qrId), consts.QRCODE_STATUS_SCANNED)
			*universal.GetCH() <- universal.CH{
				Mid:     info.Mid,
				UName:   info.UName,
				Face:    info.Face,
				ImgKey:  info.ImgKey,
				SubKey:  info.SubKey,
				Cookies: info.Cookies,
				Action:  consts.CHANNEL_ACTION_ADD_CLIENT,
			}
			os.MkdirAll(utils.GetFavourPath(info.Mid, conf.Download.Path), os.ModePerm)
			os.MkdirAll(utils.GetRecyclePath(info.Mid, conf.Download.Path), os.ModePerm)
		} else {
			SetQRCodeStatus(fmt.Sprintf("%d", qrId), consts.QRCODE_STATUS_INVALID)
		}
		os.Remove(filePath)
	}()
	return url, qrId, nil
}
