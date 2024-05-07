package client

import (
	"encoding/json"
	"strconv"

	"github.com/pkg/errors"
)

func (c *Client) GetCollected(mid int) (*CollectedInfo, error) {
	r := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").SetQueryParams(map[string]string{
		"platform": "web",
		"ps":       "20",
		"up_mid":   strconv.Itoa(mid),
	})

	ret := CollectedInfo{
		Count: 0,
		List:  make([]Collected, 0),
	}
	pn := 1
	for {
		r = r.SetQueryParam("pn", strconv.Itoa(pn))
		resp, err := r.Get("https://api.bilibili.com/x/v3/fav/folder/collected/list")
		if err != nil {
			return nil, errors.WithStack(err)
		}
		data, err := getRespData(resp, "获订阅列表")
		if err != nil {
			return nil, err
		}
		var retTmp *CollectedInfo
		err = json.Unmarshal(data, &retTmp)
		if err != nil {
			return nil, err
		}
		ret.List = append(ret.List, retTmp.List...)
		ret.Count += retTmp.Count
		if !retTmp.HasMore {
			break
		}
		pn++
	}
	if ret.Count == 0 {
		return nil, nil
	}

	return &ret, nil
}

func (c *Client) GetCollectedVideoList(seasonId int) (*CollectedVideoList, error) {
	r := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").SetQueryParams(map[string]string{
		"season_id": strconv.Itoa(seasonId),
	})
	resp, err := r.Get("https://api.bilibili.com/x/space/fav/season/list")
	if err != nil {
		return nil, errors.WithStack(err)
	}

	data, err := getRespData(resp, "获订阅视频列表")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var ret *CollectedVideoList
	err = json.Unmarshal(data, &ret)
	return ret, err
}
