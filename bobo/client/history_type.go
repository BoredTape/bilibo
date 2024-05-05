package client

// 对于这个项目，该接口，只拿bvid足以
type ToViewInfo struct {
	Count int `json:"count"`
	List  []struct {
		Bvid string `json:"bvid"`
	} `json:"list"`
}
