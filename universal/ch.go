package universal

var ch *chan CH

type CH struct {
	Mid     int
	UName   string
	Face    string
	ImgKey  string
	SubKey  string
	Cookies string
	Action  int
}

func Init() {
	chh := make(chan CH, 1)
	ch = &chh
}

func GetCH() *chan CH {
	return ch
}
