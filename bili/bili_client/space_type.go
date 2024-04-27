package bili_client

type SpaceMyInfo struct {
	Mid            int    `json:"mid"`
	Name           string `json:"name"`
	Sex            string `json:"sex"`
	Face           string `json:"face"`
	Sign           string `json:"sign"`
	Rank           int    `json:"rank"`
	Level          int    `json:"level"`
	JoinTime       int    `json:"jointime"`
	Moral          int    `json:"moral"`
	Silence        int    `json:"silence"`
	EmailStatus    int    `json:"email_status"`
	TelStatus      int    `json:"tel_status"`
	Identification int    `json:"identification"`
	Birthday       int    `json:"birthday"`
	IsTourist      int    `json:"is_tourist"`
	IsFakeAccount  int    `json:"is_fake_account"`
	PinPrompting   int    `json:"pin_prompting"`
	IsDeleted      int    `json:"is_deleted"`
	Coins          int    `json:"coins"`
	Following      int    `json:"following"`
	Follower       int    `json:"follower"`
	Vip            struct {
		Type            int    `json:"type"`
		Status          int    `json:"status"`
		DueDate         int    `json:"due_date"`
		ThemeType       int    `json:"theme_type"`
		AvatarSubscript int    `json:"avatar_subscript"`
		NicknameColor   string `json:"nickname_color"`
		Lable           struct {
			Path       string `json:"path"`
			Text       string `json:"text"`
			LabelTheme string `json:"label_theme"`
		} `json:"label"`
	} `json:"vip"`
	Pendant   Pendant `json:"pendant"`
	Nameplate struct {
		Nid        int    `json:"nid"`
		Name       string `json:"name"`
		Image      string `json:"image"`
		ImageSmall string `json:"image_small"`
		Level      string `json:"level"`
		Condition  string `json:"condition"`
	} `json:"nameplate"`
	Official  Official  `json:"official"`
	LevelInfo LevelInfo `json:"level_info"`
}
