//copy from https://github.com/CuteReimu/bilibili/blob/master/fav.go

package client

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// AddFavourFolder 新建收藏夹
//
// title：收藏夹标题，必填。intro：收藏夹简介，非必填。
// privacy：是否为私密收藏夹。cover：封面图url。
func (c *Client) AddFavourFolder(title, intro string, privacy bool, cover string) (*FavourFolderInfo, error) {
	biliJct := c.getCookie("bili_jct")
	if len(biliJct) == 0 {
		return nil, errors.New("B站登录过期")
	}
	r := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").SetQueryParams(map[string]string{
		"title": title,
		"csrf":  biliJct,
	})
	if len(intro) > 0 {
		r = r.SetQueryParam("intro", intro)
	}
	if privacy {
		r = r.SetQueryParam("privacy", "1")
	}
	if len(cover) > 0 {
		r = r.SetQueryParam("cover", cover)
	}
	resp, err := r.Post("https://api.bilibili.com/x/v3/fav/folder/add")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	data, err := getRespData(resp, "新建收藏夹")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var ret *FavourFolderInfo
	err = json.Unmarshal(data, &ret)
	return ret, err
}

// EditFavourFolder 修改收藏夹
//
// media_id：目标收藏夹mdid，必填。
// title：收藏夹标题，必填。intro：收藏夹简介，非必填。
// privacy：是否为私密收藏夹。cover：封面图url。
func (c *Client) EditFavourFolder(mediaId int, title, intro string, privacy bool, cover string) (*FavourFolderInfo, error) {
	biliJct := c.getCookie("bili_jct")
	if len(biliJct) == 0 {
		return nil, errors.New("B站登录过期")
	}
	r := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").SetQueryParams(map[string]string{
		"media_id": strconv.Itoa(mediaId),
		"title":    title,
		"csrf":     biliJct,
	})
	if len(intro) > 0 {
		r = r.SetQueryParam("intro", intro)
	}
	if privacy {
		r = r.SetQueryParam("privacy", "1")
	}
	if len(cover) > 0 {
		r = r.SetQueryParam("cover", cover)
	}
	resp, err := r.Post("https://api.bilibili.com/x/v3/fav/folder/edit")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	data, err := getRespData(resp, "修改收藏夹")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var ret *FavourFolderInfo
	err = json.Unmarshal(data, &ret)
	return ret, err
}

// DeleteFavourFolder 删除收藏夹
//
// media_ids：目标收藏夹mdid列表，必填。
func (c *Client) DeleteFavourFolder(mediaIds []int) error {
	biliJct := c.getCookie("bili_jct")
	if len(biliJct) == 0 {
		return errors.New("B站登录过期")
	}
	mediaIdsStr := make([]string, 0, len(mediaIds))
	for _, mediaId := range mediaIds {
		mediaIdsStr = append(mediaIdsStr, strconv.Itoa(mediaId))
	}
	resp, err := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").SetQueryParams(map[string]string{
		"media_ids": strings.Join(mediaIdsStr, ","),
		"csrf":      biliJct,
	}).Post("https://api.bilibili.com/x/v3/fav/folder/del")
	if err != nil {
		return errors.WithStack(err)
	}
	_, err = getRespData(resp, "删除收藏夹")
	return err
}

// CopyFavourResources 批量复制收藏内容
func (c *Client) CopyFavourResources(srcMediaId, tarMediaId, mid int, resources []Resource, platform string) error {
	biliJct := c.getCookie("bili_jct")
	if len(biliJct) == 0 {
		return errors.New("B站登录过期")
	}
	resourcesStr := make([]string, 0, len(resources))
	for _, resource := range resources {
		resourcesStr = append(resourcesStr, resource.String())
	}
	r := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").SetQueryParams(map[string]string{
		"src_media_id": strconv.Itoa(srcMediaId),
		"tar_media_id": strconv.Itoa(tarMediaId),
		"mid":          strconv.Itoa(mid),
		"resources":    strings.Join(resourcesStr, ","),
		"csrf":         biliJct,
	})
	if len(platform) > 0 {
		r = r.SetQueryParam("platform", platform)
	}
	resp, err := r.Post("https://api.bilibili.com/x/v3/fav/resource/copy")
	if err != nil {
		return errors.WithStack(err)
	}
	_, err = getRespData(resp, "批量复制收藏内容")
	return err
}

// MoveFavourResources 批量移动收藏内容
func (c *Client) MoveFavourResources(srcMediaId, tarMediaId, mid int, resources []Resource, platform string) error {
	biliJct := c.getCookie("bili_jct")
	if len(biliJct) == 0 {
		return errors.New("B站登录过期")
	}
	resourcesStr := make([]string, 0, len(resources))
	for _, resource := range resources {
		resourcesStr = append(resourcesStr, resource.String())
	}
	r := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").SetQueryParams(map[string]string{
		"src_media_id": strconv.Itoa(srcMediaId),
		"tar_media_id": strconv.Itoa(tarMediaId),
		"mid":          strconv.Itoa(mid),
		"resources":    strings.Join(resourcesStr, ","),
		"csrf":         biliJct,
	})
	if len(platform) > 0 {
		r = r.SetQueryParam("platform", platform)
	}
	resp, err := r.Post("https://api.bilibili.com/x/v3/fav/resource/move")
	if err != nil {
		return errors.WithStack(err)
	}
	_, err = getRespData(resp, "批量移动收藏内容")
	return err
}

// DeleteFavourResources 批量删除收藏内容
func (c *Client) DeleteFavourResources(mediaId int, resources []Resource, platform string) error {
	biliJct := c.getCookie("bili_jct")
	if len(biliJct) == 0 {
		return errors.New("B站登录过期")
	}
	resourcesStr := make([]string, 0, len(resources))
	for _, resource := range resources {
		resourcesStr = append(resourcesStr, resource.String())
	}
	r := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").SetQueryParams(map[string]string{
		"media_id":  strconv.Itoa(mediaId),
		"resources": strings.Join(resourcesStr, ","),
		"csrf":      biliJct,
	})
	if len(platform) > 0 {
		r.SetQueryParam("platform", platform)
	}
	resp, err := r.Post("https://api.bilibili.com/x/v3/fav/resource/batch-del")
	if err != nil {
		return errors.WithStack(err)
	}
	_, err = getRespData(resp, "批量删除收藏内容")
	return err
}

// CleanFavourResources 清空所有失效收藏内容
func (c *Client) CleanFavourResources(mediaId int) error {
	biliJct := c.getCookie("bili_jct")
	if len(biliJct) == 0 {
		return errors.New("B站登录过期")
	}
	resp, err := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").SetQueryParams(map[string]string{
		"media_id": strconv.Itoa(mediaId),
		"csrf":     biliJct,
	}).Post("https://api.bilibili.com/x/v3/fav/resource/clean")
	if err != nil {
		return errors.WithStack(err)
	}
	_, err = getRespData(resp, "清空所有失效收藏内容")
	return err
}

// GetFavourFolderInfo 获取收藏夹元数据
func (c *Client) GetFavourFolderInfo(mediaId int) (*FavourFolderInfo, error) {
	resp, err := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetQueryParam("media_id", strconv.Itoa(mediaId)).Get("https://api.bilibili.com/x/v3/fav/folder/info")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	data, err := getRespData(resp, "获取收藏夹元数据")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var ret *FavourFolderInfo
	err = json.Unmarshal(data, &ret)
	return ret, err
}

// GetAllFavourFolderInfo 获取指定用户创建的所有收藏夹信息
// attrType 目标内容属性, 0：全部 2：视频稿件
// rid 目标内容id,视频稿件：视频稿件avid
func (c *Client) GetAllFavourFolderInfo(upMid, attrType, rid int) (*AllFavourFolderInfo, error) {
	r := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").SetQueryParams(map[string]string{
		"up_mid": strconv.Itoa(upMid),
		"type":   strconv.Itoa(attrType),
	})
	if rid != 0 {
		r = r.SetQueryParam("rid", strconv.Itoa(rid))
	}
	resp, err := r.Get("https://api.bilibili.com/x/v3/fav/folder/created/list-all")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	data, err := getRespData(resp, "获取指定用户创建的所有收藏夹信息")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var ret *AllFavourFolderInfo
	err = json.Unmarshal(data, &ret)
	return ret, err
}

// GetFavourInfo 获取收藏内容
func (c *Client) GetFavourInfo(resources []Resource, platform string) ([]*FavourInfo, error) {
	resourcesStr := make([]string, 0, len(resources))
	for _, resource := range resources {
		resourcesStr = append(resourcesStr, resource.String())
	}
	r := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").SetQueryParam("resources", strings.Join(resourcesStr, ","))
	if len(platform) > 0 {
		r = r.SetQueryParam("platform", platform)
	}
	resp, err := r.Get("https://api.bilibili.com/x/v3/fav/resource/infos")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	data, err := getRespData(resp, "获取收藏内容")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var ret []*FavourInfo
	err = json.Unmarshal(data, &ret)
	return ret, err
}

// GetFavourList 获取收藏夹内容明细列表
func (c *Client) GetFavourList(mediaId, tid int, keyword, order string, searchType, ps, pn int, platform string) (*FavourList, error) {
	if pn == 0 {
		pn = 1
	}
	r := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").SetQueryParams(map[string]string{
		"media_id": strconv.Itoa(mediaId),
		"tid":      strconv.Itoa(tid),
		"type":     strconv.Itoa(searchType),
		"ps":       strconv.Itoa(ps),
		"pn":       strconv.Itoa(pn),
	})
	if len(keyword) > 0 {
		r = r.SetQueryParam("keyword", keyword)
	}
	if len(order) > 0 {
		r = r.SetQueryParam("order", order)
	}
	if len(platform) > 0 {
		r = r.SetQueryParam("platform", platform)
	}
	resp, err := r.Get("https://api.bilibili.com/x/v3/fav/resource/list")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	data, err := getRespData(resp, "获取收藏夹内容明细列表")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var ret *FavourList
	err = json.Unmarshal(data, &ret)
	return ret, err
}

// GetFavourIds 获取收藏夹全部内容id
func (c *Client) GetFavourIds(mediaId int, platform string) ([]*FavourId, error) {
	r := c.resty().R().SetHeader("Content-Type", "application/x-www-form-urlencoded").SetQueryParam("media_id", strconv.Itoa(mediaId))
	if len(platform) > 0 {
		r = r.SetQueryParam("platform", platform)
	}
	resp, err := r.Get("https://api.bilibili.com/x/v3/fav/resource/ids")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	data, err := getRespData(resp, "获取收藏夹全部内容id")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var ret []*FavourId
	err = json.Unmarshal(data, &ret)
	return ret, err
}
