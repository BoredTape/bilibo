package bili_client

import (
	"encoding/json"

	"github.com/pkg/errors"
)

func (c *Client) GetSpaceMyInfo() (*SpaceMyInfo, error) {
	resp, err := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").
		Get("https://api.bilibili.com/x/space/myinfo")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	data, err := getRespData(resp, "登录用户空间详细信息")
	if err != nil {
		return nil, err
	}
	var ret *SpaceMyInfo
	err = json.Unmarshal(data, &ret)
	return ret, err
}
