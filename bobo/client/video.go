// copy from https://github.com/CuteReimu/bilibili/blob/master/video.go
package client

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	// "github.com/disintegration/imaging"
	// ffmpeg "github.com/u2takey/ffmpeg-go"
)

var regBv = regexp.MustCompile(`(?i)bv([\dA-Za-z]{10})`)

// GetBvidByShortUrl 通过视频短链接获取bvid
func (c *Client) GetBvidByShortUrl(shortUrl string) (string, error) {
	resp, err := c.resty().SetRedirectPolicy(resty.NoRedirectPolicy()).R().Get(shortUrl)
	if resp == nil {
		return "", errors.WithStack(err)
	}
	if resp.StatusCode() != 302 {
		return "", errors.Errorf("通过短链接获取视频详细信息失败，status code: %d", resp.StatusCode())
	}
	url := resp.Header().Get("Location")
	ret := regBv.FindString(url)
	if len(ret) == 0 {
		return "", errors.New("无法解析链接：" + url)
	}
	return ret, nil
}

// GetVideoInfoByAvid 通过Avid获取视频信息
func (c *Client) GetVideoInfoByAvid(avid int) (*VideoInfo, error) {
	resp, err := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetQueryParam("aid", strconv.Itoa(avid)).Get("https://api.bilibili.com/x/web-interface/view")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	data, err := getRespData(resp, "获取视频详细信息")
	if err != nil {
		return nil, err
	}
	var ret *VideoInfo
	err = json.Unmarshal(data, &ret)
	return ret, errors.WithStack(err)
}

// GetVideoInfoByBvid 通过Bvid获取视频信息
func (c *Client) GetVideoInfoByBvid(bvid string) (*VideoInfo, error) {
	// resp, err := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").
	// 	SetQueryParam("bvid", bvid).Get("https://api.bilibili.com/x/web-interface/view")
	// if err != nil {
	// 	return nil, errors.WithStack(err)
	// }
	// data, err := getRespData(resp, "获取视频详细信息")
	// if err != nil {
	// 	return nil, err
	// }
	// var ret *VideoInfo
	// err = json.Unmarshal(data, &ret)
	// return ret, errors.WithStack(err)
	ret, _, err := c.GetVideoInfoByBvidCode(bvid)
	return ret, err
}

func (c *Client) GetVideoInfoByBvidCode(bvid string) (*VideoInfo, int64, error) {
	resp, err := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetQueryParam("bvid", bvid).Get("https://api.bilibili.com/x/web-interface/view")
	if err != nil {
		return nil, 0, errors.WithStack(err)
	}
	data, code, err := getRespDataWithCode(resp, "获取视频详细信息")
	if err != nil {
		return nil, code, err
	}
	var ret *VideoInfo
	err = json.Unmarshal(data, &ret)
	return ret, 0, errors.WithStack(err)
}

// GetVideoInfoByShortUrl 通过短链接获取视频信息
func (c *Client) GetVideoInfoByShortUrl(shortUrl string) (*VideoInfo, error) {
	bvid, err := c.GetBvidByShortUrl(shortUrl)
	if err != nil {
		return nil, err
	}
	return c.GetVideoInfoByBvid(bvid)
}

// GetRecommendVideoByAvid 通过Avid获取推荐视频
func (c *Client) GetRecommendVideoByAvid(avid int) ([]*VideoInfo, error) {
	resp, err := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetQueryParam("aid", strconv.Itoa(avid)).Get("https://api.bilibili.com/x/web-interface/archive/related")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	data, err := getRespData(resp, "获取推荐视频")
	if err != nil {
		return nil, err
	}
	var ret []*VideoInfo
	err = json.Unmarshal(data, &ret)
	return ret, errors.WithStack(err)
}

// GetRecommendVideoByBvid 通过Bvid获取推荐视频
func (c *Client) GetRecommendVideoByBvid(bvid string) ([]*VideoInfo, error) {
	resp, err := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetQueryParam("bvid", bvid).Get("https://api.bilibili.com/x/web-interface/archive/related")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	data, err := getRespData(resp, "获取推荐视频")
	if err != nil {
		return nil, err
	}
	var ret []*VideoInfo
	err = json.Unmarshal(data, &ret)
	return ret, errors.WithStack(err)
}

// GetVideoDetailInfoByAvid 通过Avid获取视频超详细信息
func (c *Client) GetVideoDetailInfoByAvid(avid int) (*VideoDetailInfo, error) {
	resp, err := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetQueryParam("aid", strconv.Itoa(avid)).Get("https://api.bilibili.com/x/web-interface/view/detail")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	data, err := getRespData(resp, "获取视频超详细信息")
	if err != nil {
		return nil, err
	}
	var ret *VideoDetailInfo
	err = json.Unmarshal(data, &ret)
	return ret, errors.WithStack(err)
}

// GetVideoDetailInfoByBvid 通过Bvid获取视频超详细信息
func (c *Client) GetVideoDetailInfoByBvid(bvid string) (*VideoDetailInfo, error) {
	resp, err := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetQueryParam("bvid", bvid).Get("https://api.bilibili.com/x/web-interface/view/detail")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	data, err := getRespData(resp, "获取视频超详细信息")
	if err != nil {
		return nil, err
	}
	var ret *VideoDetailInfo
	err = json.Unmarshal(data, &ret)
	return ret, errors.WithStack(err)
}

// GetVideoDetailInfoByShortUrl 通过短链接获取视频超详细信息
func (c *Client) GetVideoDetailInfoByShortUrl(shortUrl string) (*VideoDetailInfo, error) {
	bvid, err := c.GetBvidByShortUrl(shortUrl)
	if err != nil {
		return nil, err
	}
	return c.GetVideoDetailInfoByBvid(bvid)
}

// GetVideoDescByAvid 通过Avid获取视频简介
func (c *Client) GetVideoDescByAvid(avid int) (string, error) {
	resp, err := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetQueryParam("aid", strconv.Itoa(avid)).Get("https://api.bilibili.com/x/archive/desc")
	if err != nil {
		return "", errors.WithStack(err)
	}
	if resp.StatusCode() != 200 {
		return "", errors.Errorf("获取视频简介失败，status code: %d", resp.StatusCode())
	}
	if !gjson.ValidBytes(resp.Body()) {
		return "", errors.New("json解析失败：" + resp.String())
	}
	res := gjson.ParseBytes(resp.Body())
	code := res.Get("code").Int()
	if code != 0 {
		return "", formatError("获取视频简介", code, res.Get("message").String())
	}
	return res.Get("data").String(), errors.WithStack(err)
}

// GetVideoDescByBvid 通过Bvid获取视频简介
func (c *Client) GetVideoDescByBvid(bvid string) (string, error) {
	resp, err := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetQueryParam("bvid", bvid).Get("https://api.bilibili.com/x/archive/desc")
	if err != nil {
		return "", errors.WithStack(err)
	}
	if resp.StatusCode() != 200 {
		return "", errors.Errorf("获取视频简介失败，status code: %d", resp.StatusCode())
	}
	if !gjson.ValidBytes(resp.Body()) {
		return "", errors.New("json解析失败：" + resp.String())
	}
	res := gjson.ParseBytes(resp.Body())
	code := res.Get("code").Int()
	if code != 0 {
		return "", formatError("获取视频简介", code, res.Get("message").String())
	}
	return res.Get("data").String(), errors.WithStack(err)
}

// GetVideoDescByShortUrl 通过短链接获取视频简介
func (c *Client) GetVideoDescByShortUrl(shortUrl string) (string, error) {
	bvid, err := c.GetBvidByShortUrl(shortUrl)
	if err != nil {
		return "", err
	}
	return c.GetVideoDescByBvid(bvid)
}

// GetVideoPageListByAvid 通过Avid获取视频分P列表(Avid转cid)
func (c *Client) GetVideoPageListByAvid(avid int) ([]*VideoPage, error) {
	resp, err := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetQueryParam("aid", strconv.Itoa(avid)).Get("https://api.bilibili.com/x/player/pagelist")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	data, err := getRespData(resp, "获取视频分P列表")
	if err != nil {
		return nil, err
	}
	var ret []*VideoPage
	err = json.Unmarshal(data, &ret)
	return ret, errors.WithStack(err)
}

// GetVideoPageListByBvid 通过Bvid获取视频分P列表(Bvid转cid)
func (c *Client) GetVideoPageListByBvid(bvid string) ([]*VideoPage, error) {
	resp, err := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetQueryParam("bvid", bvid).Get("https://api.bilibili.com/x/player/pagelist")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	data, err := getRespData(resp, "获取视频分P列表")
	if err != nil {
		return nil, err
	}
	var ret []*VideoPage
	err = json.Unmarshal(data, &ret)
	return ret, errors.WithStack(err)
}

// GetVideoPageListByShortUrl 通过短链接获取视频分P列表
func (c *Client) GetVideoPageListByShortUrl(shortUrl string) ([]*VideoPage, error) {
	bvid, err := c.GetBvidByShortUrl(shortUrl)
	if err != nil {
		return nil, err
	}
	return c.GetVideoPageListByBvid(bvid)
}

// GetVideoTagsByAvid 通过Avid获取视频TAG
func (c *Client) GetVideoTagsByAvid(avid int) ([]*VideoTag, error) {
	resp, err := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetQueryParam("aid", strconv.Itoa(avid)).Get("https://api.bilibili.com/x/tag/archive/tags")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	data, err := getRespData(resp, "获取视频TAG")
	if err != nil {
		return nil, err
	}
	var ret []*VideoTag
	err = json.Unmarshal(data, &ret)
	return ret, errors.WithStack(err)
}

// GetVideoTagsByBvid 通过Bvid获取视频TAG
func (c *Client) GetVideoTagsByBvid(bvid string) ([]*VideoTag, error) {
	resp, err := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetQueryParam("bvid", bvid).Get("https://api.bilibili.com/x/tag/archive/tags")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	data, err := getRespData(resp, "获取视频TAG")
	if err != nil {
		return nil, err
	}
	var ret []*VideoTag
	err = json.Unmarshal(data, &ret)
	return ret, errors.WithStack(err)
}

// GetVideoTagsByShortUrl 通过短链接获取视频TAG
func (c *Client) GetVideoTagsByShortUrl(shortUrl string) ([]*VideoTag, error) {
	bvid, err := c.GetBvidByShortUrl(shortUrl)
	if err != nil {
		return nil, err
	}
	return c.GetVideoTagsByBvid(bvid)
}

// LikeVideoTag 点赞视频TAG，重复访问为取消
func (c *Client) LikeVideoTag(avid, tagId int) error {
	biliJct := c.getCookie("bili_jct")
	if len(biliJct) == 0 {
		return errors.New("B站登录过期")
	}
	resp, err := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").SetQueryParams(map[string]string{
		"aid":    strconv.Itoa(avid),
		"tag_id": strconv.Itoa(tagId),
		"csrf":   biliJct,
	}).Post("https://api.bilibili.com/x/tag/archive/like2")
	if err != nil {
		return errors.WithStack(err)
	}
	_, err = getRespData(resp, "点赞视频TAG")
	return err
}

// HateVideoTag 点踩视频TAG，重复访问为取消
func (c *Client) HateVideoTag(avid, tagId int) error {
	biliJct := c.getCookie("bili_jct")
	if len(biliJct) == 0 {
		return errors.New("B站登录过期")
	}
	resp, err := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").SetQueryParams(map[string]string{
		"aid":    strconv.Itoa(avid),
		"tag_id": strconv.Itoa(tagId),
		"csrf":   biliJct,
	}).Post("https://api.bilibili.com/x/tag/archive/hate2")
	if err != nil {
		return errors.WithStack(err)
	}
	_, err = getRespData(resp, "点踩视频TAG")
	return err
}

// LikeVideoByAvid 通过Avid点赞视频，like为false表示取消点赞
func (c *Client) LikeVideoByAvid(avid int, like bool) error {
	biliJct := c.getCookie("bili_jct")
	if len(biliJct) == 0 {
		return errors.New("B站登录过期")
	}
	var likeNum string
	if like {
		likeNum = "1"
	} else {
		likeNum = "2"
	}
	resp, err := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").SetQueryParams(map[string]string{
		"aid":  strconv.Itoa(avid),
		"like": likeNum,
		"csrf": biliJct,
	}).Post("https://api.bilibili.com/x/web-interface/archive/like")
	if err != nil {
		return errors.WithStack(err)
	}
	_, err = getRespData(resp, "点赞视频")
	return err
}

// LikeVideoByBvid 通过Bvid点赞视频，like为false表示取消点赞
func (c *Client) LikeVideoByBvid(bvid string, like bool) error {
	biliJct := c.getCookie("bili_jct")
	if len(biliJct) == 0 {
		return errors.New("B站登录过期")
	}
	var likeNum string
	if like {
		likeNum = "1"
	} else {
		likeNum = "2"
	}
	resp, err := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").SetQueryParams(map[string]string{
		"bvid": bvid,
		"like": likeNum,
		"csrf": biliJct,
	}).Post("https://api.bilibili.com/x/web-interface/archive/like")
	if err != nil {
		return errors.WithStack(err)
	}
	_, err = getRespData(resp, "点赞视频")
	return err
}

// LikeVideoByShortUrl 通过短链接点赞视频，like为false表示取消点赞
func (c *Client) LikeVideoByShortUrl(shortUrl string, like bool) error {
	bvid, err := c.GetBvidByShortUrl(shortUrl)
	if err != nil {
		return err
	}
	return c.LikeVideoByBvid(bvid, like)
}

// CoinVideoByAvid 通过Avid投币视频，multiply为投币数量，上限为2，like为是否附加点赞。返回是否附加点赞成功
func (c *Client) CoinVideoByAvid(avid int, multiply int, like bool) (bool, error) {
	biliJct := c.getCookie("bili_jct")
	if len(biliJct) == 0 {
		return false, errors.New("B站登录过期")
	}
	var likeNum string
	if like {
		likeNum = "1"
	} else {
		likeNum = "0"
	}
	resp, err := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").SetQueryParams(map[string]string{
		"aid":         strconv.Itoa(avid),
		"select_like": likeNum,
		"multiply":    strconv.Itoa(multiply),
		"csrf":        biliJct,
	}).Post("https://api.bilibili.com/x/web-interface/coin/add")
	if err != nil {
		return false, errors.WithStack(err)
	}
	data, err := getRespData(resp, "投币视频")
	if err != nil {
		return false, err
	}
	return gjson.GetBytes(data, "like").Bool(), nil
}

// CoinVideoByBvid 通过Bvid投币视频，multiply为投币数量，上限为2，like为是否附加点赞。返回是否附加点赞成功
func (c *Client) CoinVideoByBvid(bvid string, multiply int, like bool) (bool, error) {
	biliJct := c.getCookie("bili_jct")
	if len(biliJct) == 0 {
		return false, errors.New("B站登录过期")
	}
	var likeNum string
	if like {
		likeNum = "1"
	} else {
		likeNum = "0"
	}
	resp, err := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").SetQueryParams(map[string]string{
		"bvid":        bvid,
		"select_like": likeNum,
		"multiply":    strconv.Itoa(multiply),
		"csrf":        biliJct,
	}).Post("https://api.bilibili.com/x/web-interface/coin/add")
	if err != nil {
		return false, errors.WithStack(err)
	}
	data, err := getRespData(resp, "投币视频")
	if err != nil {
		return false, err
	}
	return gjson.GetBytes(data, "like").Bool(), nil
}

// CoinVideoByShortUrl 通过短链接投币视频，multiply为投币数量，上限为2，like为是否附加点赞。返回是否附加点赞成功
func (c *Client) CoinVideoByShortUrl(shortUrl string, multiply int, like bool) (bool, error) {
	bvid, err := c.GetBvidByShortUrl(shortUrl)
	if err != nil {
		return false, err
	}
	return c.CoinVideoByBvid(bvid, multiply, like)
}

// FavourVideoByAvid 通过Avid收藏视频，addMediaIds和delMediaIds为要增加/删除的收藏列表，非必填。返回是否为未关注用户收藏
func (c *Client) FavourVideoByAvid(avid int, addMediaIds, delMediaIds []int) (bool, error) {
	biliJct := c.getCookie("bili_jct")
	if len(biliJct) == 0 {
		return false, errors.New("B站登录过期")
	}
	var addMediaIdStr, delMediaIdStr []string
	for _, id := range addMediaIds {
		addMediaIdStr = append(addMediaIdStr, strconv.Itoa(id))
	}
	for _, id := range delMediaIds {
		delMediaIdStr = append(delMediaIdStr, strconv.Itoa(id))
	}
	resp, err := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").SetQueryParams(map[string]string{
		"rid":           strconv.Itoa(avid),
		"type":          "2",
		"add_media_ids": strings.Join(addMediaIdStr, ","),
		"del_media_ids": strings.Join(delMediaIdStr, ","),
		"csrf":          biliJct,
	}).Post("https://api.bilibili.com/medialist/gateway/coll/resource/deal")
	if err != nil {
		return false, errors.WithStack(err)
	}
	data, err := getRespData(resp, "收藏视频")
	if err != nil {
		return false, err
	}
	return gjson.GetBytes(data, "prompt").Bool(), nil
}

// FavourVideoByBvid 通过Bvid收藏视频，addMediaIds和delMediaIds为要增加/删除的收藏列表，非必填。返回是否为未关注用户收藏
func (c *Client) FavourVideoByBvid(bvid string, addMediaIds, delMediaIds []int) (bool, error) {
	biliJct := c.getCookie("bili_jct")
	if len(biliJct) == 0 {
		return false, errors.New("B站登录过期")
	}
	var addMediaIdStr, delMediaIdStr []string
	for _, id := range addMediaIds {
		addMediaIdStr = append(addMediaIdStr, strconv.Itoa(id))
	}
	for _, id := range delMediaIds {
		delMediaIdStr = append(delMediaIdStr, strconv.Itoa(id))
	}
	resp, err := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").SetQueryParams(map[string]string{
		"rid":           bvid,
		"type":          "2",
		"add_media_ids": strings.Join(addMediaIdStr, ","),
		"del_media_ids": strings.Join(delMediaIdStr, ","),
		"csrf":          biliJct,
	}).Post("https://api.bilibili.com/medialist/gateway/coll/resource/deal")
	if err != nil {
		return false, errors.WithStack(err)
	}
	data, err := getRespData(resp, "收藏视频")
	if err != nil {
		return false, err
	}
	return gjson.GetBytes(data, "prompt").Bool(), nil
}

// FavourVideoByShortUrl 通过短链接收藏视频，addMediaIds和delMediaIds为要增加/删除的收藏列表，非必填。返回是否为未关注用户收藏
func (c *Client) FavourVideoByShortUrl(shortUrl string, addMediaIds, delMediaIds []int) (bool, error) {
	bvid, err := c.GetBvidByShortUrl(shortUrl)
	if err != nil {
		return false, err
	}
	return c.FavourVideoByBvid(bvid, addMediaIds, delMediaIds)
}

// LikeCoinFavourVideoByAvid 通过Avid一键三连视频
func (c *Client) LikeCoinFavourVideoByAvid(avid int) (*LikeCoinFavourResult, error) {
	biliJct := c.getCookie("bili_jct")
	if len(biliJct) == 0 {
		return nil, errors.New("B站登录过期")
	}
	resp, err := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").SetQueryParams(map[string]string{
		"aid":  strconv.Itoa(avid),
		"csrf": biliJct,
	}).Post("https://api.bilibili.com/x/web-interface/archive/like/triple")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	data, err := getRespData(resp, "一键三连视频")
	if err != nil {
		return nil, err
	}
	var ret *LikeCoinFavourResult
	err = json.Unmarshal(data, &ret)
	return ret, errors.WithStack(err)
}

// LikeCoinFavourVideoByBvid 通过Bvid一键三连视频
func (c *Client) LikeCoinFavourVideoByBvid(bvid string) (*LikeCoinFavourResult, error) {
	biliJct := c.getCookie("bili_jct")
	if len(biliJct) == 0 {
		return nil, errors.New("B站登录过期")
	}
	resp, err := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").SetQueryParams(map[string]string{
		"bvid": bvid,
		"csrf": biliJct,
	}).Post("https://api.bilibili.com/x/web-interface/archive/like/triple")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	data, err := getRespData(resp, "一键三连视频")
	if err != nil {
		return nil, err
	}
	var ret *LikeCoinFavourResult
	err = json.Unmarshal(data, &ret)
	return ret, errors.WithStack(err)
}

// LikeCoinFavourVideoByShortUrl 通过短链接一键三连视频
func (c *Client) LikeCoinFavourVideoByShortUrl(shortUrl string) (*LikeCoinFavourResult, error) {
	bvid, err := c.GetBvidByShortUrl(shortUrl)
	if err != nil {
		return nil, err
	}
	return c.LikeCoinFavourVideoByBvid(bvid)
}

// GetVideoOnlineInfoByAvid 通过Avid获取视频在线人数
func (c *Client) GetVideoOnlineInfoByAvid(avid, cid int) (*VideoOnlineInfo, error) {
	resp, err := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").SetQueryParams(map[string]string{
		"aid": strconv.Itoa(avid),
		"cid": strconv.Itoa(cid),
	}).Get("https://api.bilibili.com/x/player/online/total")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	data, err := getRespData(resp, "获取视频在线人数")
	if err != nil {
		return nil, err
	}
	var ret *VideoOnlineInfo
	err = json.Unmarshal(data, &ret)
	return ret, errors.WithStack(err)
}

// GetVideoOnlineInfoByBvid 通过Bvid获取视频在线人数
func (c *Client) GetVideoOnlineInfoByBvid(bvid string, cid int) (*VideoOnlineInfo, error) {
	resp, err := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").SetQueryParams(map[string]string{
		"bvid": bvid,
		"cid":  strconv.Itoa(cid),
	}).Get("https://api.bilibili.com/x/player/online/total")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	data, err := getRespData(resp, "获取视频在线人数")
	if err != nil {
		return nil, err
	}
	var ret *VideoOnlineInfo
	err = json.Unmarshal(data, &ret)
	return ret, errors.WithStack(err)
}

// GetVideoOnlineInfoByShortUrl 通过短链接获取视频在线人数
func (c *Client) GetVideoOnlineInfoByShortUrl(shortUrl string, cid int) (*VideoOnlineInfo, error) {
	bvid, err := c.GetBvidByShortUrl(shortUrl)
	if err != nil {
		return nil, err
	}
	return c.GetVideoOnlineInfoByBvid(bvid, cid)
}

// GetVideoPbPInfo 获取视频弹幕趋势顶点列表（高能进度条）
func (c *Client) GetVideoPbPInfo(cid int) (*VideoPbPInfo, error) {
	resp, err := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetQueryParam("cid", strconv.Itoa(cid)).Get("https://api.bilibili.com/pbp/data")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	data, err := getRespData(resp, "获取视频弹幕趋势顶点列表")
	if err != nil {
		return nil, err
	}
	var ret *VideoPbPInfo
	err = json.Unmarshal(data, &ret)
	return ret, errors.WithStack(err)
}

// GetVideoStatusNumberByAvid 通过Avid获取视频状态数视频
func (c *Client) GetVideoStatusNumberByAvid(avid int) (*VideoStatusNumber, error) {
	resp, err := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetQueryParam("aid", strconv.Itoa(avid)).Get("https://api.bilibili.com/x/web-interface/archive/stat")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	data, err := getRespData(resp, "获取视频状态数视频")
	if err != nil {
		return nil, err
	}
	var ret *VideoStatusNumber
	err = json.Unmarshal(data, &ret)
	return ret, errors.WithStack(err)
}

// GetVideoStatusNumberByBvid 通过Bvid获取视频状态数
func (c *Client) GetVideoStatusNumberByBvid(bvid string) (*VideoStatusNumber, error) {
	resp, err := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetQueryParam("bvid", bvid).Get("https://api.bilibili.com/x/web-interface/archive/stat")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	data, err := getRespData(resp, "获取视频状态数")
	if err != nil {
		return nil, err
	}
	var ret *VideoStatusNumber
	err = json.Unmarshal(data, &ret)
	return ret, errors.WithStack(err)
}

// GetVideoStatusNumberByShortUrl 通过短链接获取视频状态数
func (c *Client) GetVideoStatusNumberByShortUrl(shortUrl string) (*VideoStatusNumber, error) {
	bvid, err := c.GetBvidByShortUrl(shortUrl)
	if err != nil {
		return nil, err
	}
	return c.GetVideoStatusNumberByBvid(bvid)
}

// GetTopRecommendVideo 获取首页视频推荐列表，freshType相关性（默认为3），ps单页返回的记录条数（默认为8）
func (c *Client) GetTopRecommendVideo(freshType, ps int) ([]*VideoInfo, error) {
	request := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").SetQueryParam("version", "1")
	if freshType != 0 {
		request.SetQueryParam("fresh_type", strconv.Itoa(freshType))
	}
	if ps != 0 {
		request.SetQueryParam("ps", strconv.Itoa(ps))
	}
	resp, err := request.Get("https://api.bilibili.com/x/web-interface/index/top/rcmd")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	data, err := getRespData(resp, "获取首页视频推荐列表")
	if err != nil {
		return nil, err
	}
	var ret []*VideoInfo
	err = json.Unmarshal(data, &ret)
	return ret, errors.WithStack(err)
}

func (c *Client) GetVideoPlayUrlByBvid(cid int, bvid string) (*DownloadInfo, error) {
	request := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded")
	request.SetQueryParam("bvid", bvid)
	request.SetQueryParam("cid", strconv.Itoa(cid))
	request.SetQueryParam("qn", "127")
	request.SetQueryParam("otype", "json")
	request.SetQueryParam("fnval", "4048")
	request.SetQueryParam("fourk", "1")

	resp, err := request.Get("https://api.bilibili.com/x/player/playurl")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	data, err := getRespData(resp, "获取首页视频推荐列表")
	if err != nil {
		return nil, err
	}
	var ret DownloadInfo
	err = json.Unmarshal(data, &ret)
	return &ret, errors.WithStack(err)
}

func (c *Client) GetWbiVideoPlayUrlByBvid(cid int, bvid string) (*DownloadInfo, error) {
	request := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded")
	request.SetQueryParam("bvid", bvid)
	request.SetQueryParam("cid", strconv.Itoa(cid))
	request.SetQueryParam("qn", "127")
	request.SetQueryParam("otype", "json")
	request.SetQueryParam("fnval", "4048")
	request.SetQueryParam("fourk", "1")
	requestNew, err := signAndQueryParams(request.QueryParam, c.imgKey, c.subKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	request.SetQueryParamsFromValues(requestNew)
	resp, err := request.Get("https://api.bilibili.com/x/player/playurl")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	data, webRefresh, err := getRespDataWithCheckWebi(resp, "获取视频地址")
	if err != nil {
		return nil, err
	}
	if webRefresh {
		nav, _, err := c.GetNavigation()
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if err := c.RefreshWbiKey(nav); err == nil {
			return c.GetWbiVideoPlayUrlByBvid(cid, bvid)
		} else {
			return nil, errors.WithStack(err)
		}
	}
	var ret DownloadInfo
	err = json.Unmarshal(data, &ret)
	return &ret, errors.WithStack(err)
}

func (c *Client) DownloadVideoBestByBvidCid(cid int, bvid, filePath, fileName string) (string, string, error) {
	videoDownloadInfo, err := c.GetWbiVideoPlayUrlByBvid(cid, bvid)
	if err != nil {
		return "", "", err
	}
	stream := NewDetecter(videoDownloadInfo).DetectBest(VIDEO_CODECID_AVC)
	if stream == nil || stream.Video == nil {
		return "", "", errors.New("无可用视频流")
	}
	mimeType := strings.Split(stream.Video.MimeType, "/")[1]
	targetName := fmt.Sprintf("%s.%s", fileName, mimeType)
	videoPath := filepath.Join(filePath, targetName)

	videoTmpName := fmt.Sprintf("%sv", targetName)
	videoTmpPath := filepath.Join(filePath, videoTmpName)
	defer func() {
		os.Remove(videoTmpPath)
	}()
	if stream.Audio != nil {
		err = download(c.ua, stream.Video.Url, videoTmpPath)
	} else {
		err = download(c.ua, stream.Video.Url, videoPath)
	}
	if err != nil {
		return "", "", err
	}
	fmt.Printf("%s video download complate!\n", fileName)
	if stream.Audio != nil {
		audioTmpName := fmt.Sprintf("%ss", targetName)
		audioTmpPath := filepath.Join(filePath, audioTmpName)
		defer func() {
			os.Remove(audioTmpPath)
		}()
		if err := download(c.ua, stream.Audio.Url, audioTmpPath); err != nil {
			return "", "", err
		}

		fmt.Printf("%s audio download complate!\n", fileName)
		if err := exec.Command(
			"ffmpeg", "-i", videoTmpPath, "-i", audioTmpPath,
			"-c:v", "copy", "-c:a", "copy", "-f",
			mimeType, videoPath,
		).Run(); err != nil {
			fmt.Printf("merge video and audio error: %v", err)
			fmt.Println("ffmpeg", " -i ", videoTmpPath, " -i ", audioTmpPath,
				" -c:v", " copy", " -c:a", " copy", " -f ",
				mimeType, " ", videoPath)
			return "", "", err
		}
	}

	return videoPath, mimeType, nil
}

// func GenerateCover(videoPath string) error {
// 	// 生成封面
// 	logger := log.GetLogger()
// 	buf := bytes.NewBuffer(nil)
// 	frameNum := 1
// 	err := ffmpeg.Input(videoPath).
// 		Filter("select", ffmpeg.Args{fmt.Sprintf("gte(n,%d)", frameNum)}).
// 		Output("pipe:", ffmpeg.KwArgs{"vframes": 1, "format": "image2", "vcodec": "mjpeg"}).
// 		WithOutput(buf, os.Stdout).
// 		Run()
// 	if err != nil {
// 		logger.Info("ffmpeg截取缩略图失败：", err)
// 		return err
// 	}
// 	img, err := imaging.Decode(buf)
// 	if err != nil {
// 		logger.Info("imaging Decode缩略图失败：", err)
// 		return err
// 	}
// 	err = imaging.Save(img, videoPath+".png")
// 	if err != nil {
// 		logger.Info("保存略图失败：", err)
// 		return err
// 	}
// 	return nil
// }
