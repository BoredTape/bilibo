// copy from https://github.com/CuteReimu/bilibili/blob/master/client.go
package bili_client

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
)

const DEFAULT_UA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.90 Safari/537.36"

type Client struct {
	cookies []*http.Cookie
	timeout time.Duration
	logger  resty.Logger
	ua      string
	imgKey  string
	subKey  string
	mid     int
}

type ClientOption func(c *Client)

func WithTimeout(timeout time.Duration) ClientOption {
	if timeout == 0 {
		timeout = 20 * time.Second
	}
	return func(c *Client) {
		c.timeout = timeout
	}
}

func WithUA(ua string) ClientOption {
	if ua == "" {
		ua = DEFAULT_UA
	}
	return func(c *Client) {
		c.ua = ua
	}
}

func WithMid(mid int) ClientOption {
	return func(c *Client) {
		c.mid = mid
	}
}

func WithImgKey(key string) ClientOption {
	return func(c *Client) {
		c.imgKey = key
	}
}

func WithSubKey(key string) ClientOption {
	return func(c *Client) {
		c.subKey = key
	}
}

func WithCookiesStrings(cookies string) ClientOption {
	return func(c *Client) {
		c.SetCookiesString(cookies)
	}
}

func NewClient(opts ...ClientOption) *Client {
	c := &Client{}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Client) RefreshWbiKey(nav *Navigation) error {
	if len(c.cookies) < 1 {
		return errors.New("未登录")
	}
	imgUrl := strings.Split(nav.WbiImg.ImgUrl, "/")
	subUrl := strings.Split(nav.WbiImg.SubUrl, "/")
	c.imgKey = strings.Split(imgUrl[len(imgUrl)-1], ".")[0]
	c.subKey = strings.Split(subUrl[len(subUrl)-1], ".")[0]
	c.mid = nav.Mid
	return nil
}

func (c *Client) GetMid() int {
	return c.mid
}

func (c *Client) GetWbiRunningTime() (string, string) {
	return c.imgKey, c.subKey
}

func (c *Client) GetWbi() (string, string, int64, error) {
	if (c.imgKey != "" || c.subKey != "" || c.mid != 0) && len(c.cookies) > 0 {
		nav, errorCode, err := c.GetNavigation()
		if err != nil {
			return "", "", errorCode, err
		}
		if err := c.RefreshWbiKey(nav); err != nil {
			return "", "", errorCode, errors.WithStack(err)
		} else {
			return c.imgKey, c.subKey, 0, nil
		}
	} else {
		return "", "", -101, errors.New("未登录")
	}
}

func (c *Client) SetLogger(logger resty.Logger) {
	c.logger = logger
}

func (c *Client) GetLogger() resty.Logger {
	return c.logger
}

// 根据key获取指定的cookie值
func (c *Client) getCookie(name string) string {
	now := time.Now()
	for _, cookie := range c.cookies {
		if cookie.Name == name && now.Before(cookie.Expires) {
			return cookie.Value
		}
	}
	return ""
}

func (c *Client) GetCookiesString() string {
	var cookieStrings []string
	for _, cookie := range c.cookies {
		cookieStrings = append(cookieStrings, cookie.String())
	}
	return strings.Join(cookieStrings, "\n")
}

func (c *Client) SetCookiesString(cookiesString string) {
	c.setCookies((&resty.Response{RawResponse: &http.Response{Header: http.Header{
		"Set-Cookie": strings.Split(cookiesString, "\n"),
	}}}).Cookies())
}

func (c *Client) setCookies(cookies []*http.Cookie) {
	c.cookies = cookies
}

func (c *Client) resty() *resty.Client {
	client := resty.New().SetTimeout(c.timeout).SetHeader("user-agent", c.ua)
	if c.logger != nil {
		client.SetLogger(c.logger)
	}
	if c.cookies != nil {
		client.SetCookies(c.cookies)
	}
	return client
}

func (c *Client) GetResty() *resty.Client {
	return c.resty()
}
