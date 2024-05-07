package client

// 对于这个项目，该接口，只拿这些信息足以
type Collected struct {
	Id         int    `json:"id"`
	Mid        int    `json:"mid"`
	Attr       int    `json:"attr"`
	Title      string `json:"title"`
	MediaCount int    `json:"media_count"`
}
type CollectedInfo struct {
	Count   int         `json:"count"`
	List    []Collected `json:"list"`
	HasMore bool        `json:"has_more"`
}

type CollectedVideoList struct {
	Medias []struct {
		BvId string `json:"bvid"`
	} `json:"medias"`
}
