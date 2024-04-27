// copy from https://github.com/CuteReimu/bilibili/blob/master/fav.go
package bili_client

type FavourFolderInfo struct {
	Id    int      `json:"id"`    // 收藏夹mlid（完整id），收藏夹原始id+创建者mid尾号2位
	Fid   int      `json:"fid"`   // 收藏夹原始id
	Mid   int      `json:"mid"`   // 创建者mid
	Attr  int      `json:"attr"`  // 属性位（？）
	Title string   `json:"title"` // 收藏夹标题
	Cover string   `json:"cover"` // 	收藏夹封面图片url
	Upper struct { // 创建者信息
		Mid       int    `json:"mid"`        // 创建者mid
		Name      string `json:"name"`       // 创建者昵称
		Face      string `json:"face"`       // 创建者头像url
		Followed  bool   `json:"followed"`   // 是否已关注创建者
		VipType   int    `json:"vip_type"`   // 会员类别，0：无，1：月大会员，2：年度及以上大会员
		VipStatue int    `json:"vip_statue"` // 0：无，1：有
	} `json:"upper"`
	CoverType int      `json:"cover_type"` // 封面图类别（？）
	CntInfo   struct { // 收藏夹状态数
		Collect int `json:"collect"`  // 收藏数
		Play    int `json:"play"`     // 播放数
		ThumbUp int `json:"thumb_up"` // 点赞数
		Share   int `json:"share"`    // 分享数
	} `json:"cnt_info"`
	Type       int    `json:"type"`        // 类型（？）
	Intro      string `json:"intro"`       // 备注
	Ctime      int    `json:"ctime"`       // 创建时间戳
	Mtime      int    `json:"mtime"`       // 收藏时间戳
	State      int    `json:"state"`       // 状态（？）
	FavState   int    `json:"fav_state"`   // 收藏夹收藏状态，已收藏：1，未收藏：0
	LikeState  int    `json:"like_state"`  // 点赞状态，已点赞：1，未点赞：0
	MediaCount int    `json:"media_count"` // 收藏夹内容数量
}

type AllFavourFolderInfo struct {
	Count int        `json:"count"` // 创建的收藏夹总数
	List  []struct { // 创建的收藏夹列表
		Id         int    `json:"id"`          // 收藏夹mlid（完整id），收藏夹原始id+创建者mid尾号2位
		Fid        int    `json:"fid"`         // 收藏夹原始id
		Mid        int    `json:"mid"`         // 创建者mid
		Attr       int    `json:"attr"`        // 属性位（？）
		Title      string `json:"title"`       // 收藏夹标题
		FavState   int    `json:"fav_state"`   // 目标id是否存在于该收藏夹，存在于该收藏夹：1，不存在于该收藏夹：0
		MediaCount int    `json:"media_count"` // 收藏夹内容数量
	} `json:"list"`
}

type FavourInfo struct {
	Id       int    `json:"id"`
	Type     int    `json:"type"`
	Title    string `json:"title"`
	Cover    string `json:"cover"`
	Intro    string `json:"intro"`
	Page     int    `json:"page"`
	Duration int    `json:"duration"`
	Upper    struct {
		Mid  int    `json:"mid"`
		Name string `json:"name"`
		Face string `json:"face"`
	} `json:"upper"`
	Attr    int `json:"attr"`
	CntInfo struct {
		Collect int `json:"collect"`
		Play    int `json:"play"`
		Danmaku int `json:"danmaku"`
	} `json:"cnt_info"`
	Link    string      `json:"link"`
	Ctime   int         `json:"ctime"`
	Pubtime int         `json:"pubtime"`
	FavTime int         `json:"fav_time"`
	BvId    string      `json:"bv_id"`
	Bvid    string      `json:"bvid"`
	Season  interface{} `json:"season"`
}

type FavourList struct {
	Info struct { // 收藏夹元数据
		Id    int      `json:"id"`    // 收藏夹mlid（完整id），收藏夹原始id+创建者mid尾号2位
		Fid   int      `json:"fid"`   // 收藏夹原始id
		Mid   int      `json:"mid"`   // 创建者mid
		Attr  int      `json:"attr"`  // 属性，0：正常，1：失效
		Title string   `json:"title"` // 收藏夹标题
		Cover string   `json:"cover"` // 收藏夹封面图片url
		Upper struct { // 创建者信息
			Mid       int    `json:"mid"`        // 创建者mid
			Name      string `json:"name"`       // 创建者昵称
			Face      string `json:"face"`       // 创建者头像url
			Followed  bool   `json:"followed"`   // 是否已关注创建者
			VipType   int    `json:"vip_type"`   // 会员类别，0：无，1：月大会员，2：年度及以上大会员
			VipStatue int    `json:"vip_statue"` // 会员开通状态，0：无，1：有
		} `json:"upper"`
		CoverType int      `json:"cover_type"` // 封面图类别（？）
		CntInfo   struct { // 收藏夹状态数
			Collect int `json:"collect"`  // 收藏数
			Play    int `json:"play"`     // 播放数
			ThumbUp int `json:"thumb_up"` // 点赞数
			Share   int `json:"share"`    // 分享数
		} `json:"cnt_info"`
		Type       int    `json:"type"`        // 类型（？），一般是11
		Intro      string `json:"intro"`       // 备注
		Ctime      int    `json:"ctime"`       // 创建时间戳
		Mtime      int    `json:"mtime"`       // 收藏时间戳
		State      int    `json:"state"`       // 状态（？），一般为0
		FavState   int    `json:"fav_state"`   // 收藏夹收藏状态，已收藏收藏夹：1，未收藏收藏夹：0
		LikeState  int    `json:"like_state"`  // 点赞状态，已点赞：1，未点赞：0
		MediaCount int    `json:"media_count"` // 收藏夹内容数量
	} `json:"info"`
	Medias []struct { // 收藏夹内容
		Id       int      `json:"id"`       // 内容id，视频稿件：视频稿件avid，音频：音频auid，视频合集：视频合集id
		Type     int      `json:"type"`     // 内容类型，2：视频稿件，12：音频，21：视频合集
		Title    string   `json:"title"`    // 标题
		Cover    string   `json:"cover"`    // 封面url
		Intro    string   `json:"intro"`    // 简介
		Page     int      `json:"page"`     // 视频分P数
		Duration int      `json:"duration"` // 音频/视频时长
		Upper    struct { // UP主信息
			Mid  int    `json:"mid"`  // UP主mid
			Name string `json:"name"` // UP主昵称
			Face string `json:"face"` // UP主头像url
		} `json:"upper"`
		Attr    int      `json:"attr"` // 属性位（？）
		CntInfo struct { // 状态数
			Collect int `json:"collect"` // 收藏数
			Play    int `json:"play"`    // 播放数
			Danmaku int `json:"danmaku"` // 弹幕数
		} `json:"cnt_info"`
		Link    string `json:"link"`     // 跳转uri
		Ctime   int    `json:"ctime"`    // 投稿时间戳
		Pubtime int    `json:"pubtime"`  // 发布时间戳
		FavTime int    `json:"fav_time"` // 收藏时间戳
		BvId    string `json:"bv_id"`    // 视频稿件bvid
		Bvid    string `json:"bvid"`     // 视频稿件bvid
	} `json:"medias"`
	HasMore bool `json:"has_more"`
}

type FavourId struct {
	Id   int    `json:"id"`    // 内容id，视频稿件：视频稿件avid，音频：音频auid，视频合集：视频合集id
	Type int    `json:"type"`  // 内容类型，2：视频稿件，12：音频，21：视频合集
	BvId string `json:"bv_id"` // 视频稿件bvid
	Bvid string `json:"bvid"`  // 视频稿件bvid
}
