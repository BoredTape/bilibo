package client

import (
	"encoding/json"

	"github.com/pkg/errors"
)

func (c *Client) GetToView() (*ToViewInfo, error) {
	resp, err := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").
		Get("https://api.bilibili.com/x/v2/history/toview")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	data, err := getRespData(resp, "获取稍后再看视频列表")
	if err != nil {
		return nil, err
	}
	var ret *ToViewInfo
	err = json.Unmarshal(data, &ret)
	return ret, err
}
