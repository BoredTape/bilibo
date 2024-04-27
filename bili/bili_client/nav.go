package bili_client

import (
	"encoding/json"

	"github.com/pkg/errors"
)

// GetNavigation 获取导航
func (c *Client) GetNavigation() (*Navigation, int64, error) {
	resp, err := c.resty().R().Get("https://api.bilibili.com/x/web-interface/nav")
	if err != nil {
		return nil, 0, errors.WithStack(err)
	}
	data, errorCode, err := getRespDataWithCode(resp, "导航栏用户信息")
	if err != nil {
		return nil, errorCode, err
	}
	var ret *Navigation
	err = json.Unmarshal(data, &ret)
	return ret, errorCode, err
}
