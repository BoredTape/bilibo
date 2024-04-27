package bili_client

type LevelInfo struct {
	CurrentLevel int         `json:"current_level"`
	CurrentMin   int         `json:"current_min"`
	CurrentExp   int         `json:"current_exp"`
	NextExp      interface{} `json:"next_exp"` //fuck,小于6级时:int 6级时:string
}

type Official struct {
	Role  int    `json:"role"`
	Title string `json:"title"`
	Desc  string `json:"desc"`
	Type  int    `json:"type"`
}

type VipLabel struct {
	Path       string `json:"path"`
	Text       string `json:"text"`
	LabelTheme string `json:"label_theme"`
}

type Wallet struct {
	Mid           int     `json:"mid"`
	BcoinBalance  float32 `json:"bcoin_balance"`
	CouponBalance int     `json:"coupon_balance"`
	CouponDueTime int     `json:"coupon_due_time"`
}

type WbiImg struct {
	ImgUrl string `json:"img_url"`
	SubUrl string `json:"sub_url"`
}

type Navigation struct {
	IsLogin            bool           `json:"isLogin"`
	EmailVerified      int            `json:"email_verified"`
	Face               string         `json:"face"`
	LevelInfo          LevelInfo      `json:"level_info"`
	Mid                int            `json:"mid"`
	MobileVerified     int            `json:"mobile_verified"`
	Money              int            `json:"money"`
	Moral              int            `json:"moral"`
	Official           Official       `json:"official"`
	OfficialVerify     OfficialVerify `json:"officialVerify"`
	Pendant            Pendant        `json:"pendant"`
	Scores             int            `json:"scores"`
	Uname              string         `json:"uname"`
	VipDueDate         int            `json:"vipDueDate"`
	VipStatus          int            `json:"vipStatus"`
	VipType            int            `json:"vipType"`
	VipPayType         int            `json:"vip_pay_type"`
	VipThemeType       int            `json:"vip_theme_type"`
	VipLabel           VipLabel       `json:"vip_label"`
	VipAvatarSubscript int            `json:"vip_avatar_subscript"`
	VipNicknameColor   string         `json:"vip_nickname_color"`
	Wallet             Wallet         `json:"wallet"`
	HasShop            bool           `json:"has_shop"`
	ShopUrl            string         `json:"shop_url"`
	AllowanceCount     int            `json:"allowance_count"`
	AnswerStatus       int            `json:"answer_status"`
	IsSeniorMember     int            `json:"is_senior_member"`
	WbiImg             WbiImg         `json:"wbi_img"`
	IsJury             bool           `json:"is_jury"`
}
