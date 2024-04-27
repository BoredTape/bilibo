package bili_client

const (
	/*
		视频清晰度标识

		| 值   | 含义           | 备注                                                         |
		| ---- | -------------- | ------------------------------------------------------------ |
		| 6    | 240P 极速      | 仅 MP4 格式支持<br />仅`platform=html5`时有效                |
		| 16   | 360P 流畅      |                                                              |
		| 32   | 480P 清晰      |                                                              |
		| 64   | 720P 高清      | WEB 端默认值<br />B站前端需要登录才能选择，但是直接发送请求可以不登录就拿到 720P 的取流地址<br />**无 720P 时则为 720P60** |
		| 74   | 720P60 高帧率  | 登录认证                                                     |
		| 80   | 1080P 高清     | TV 端与 APP 端默认值<br />登录认证                           |
		| 112  | 1080P+ 高码率  | 大会员认证                                                   |
		| 116  | 1080P60 高帧率 | 大会员认证                                                   |
		| 120  | 4K 超清        | 需要`fnval&128=128`且`fourk=1`<br />大会员认证               |
		| 125  | HDR 真彩色     | 仅支持 DASH 格式<br />需要`fnval&64=64`<br />大会员认证      |
		| 126  | 杜比视界       | 仅支持 DASH 格式<br />需要`fnval&512=512`<br />大会员认证    |
		| 127  | 8K 超高清      | 仅支持 DASH 格式<br />需要`fnval&1024=1024`<br />大会员认证  |
	*/
	VIDEO_QUALITY_240P       = 6
	VIDEO_QUALITY_360P       = 16
	VIDEO_QUALITY_480P       = 32
	VIDEO_QUALITY_720P       = 64
	VIDEO_QUALITY_720P_60    = 74
	VIDEO_QUALITY_1080P      = 80
	VIDEO_QUALITY_1080P_PLUS = 112
	VIDEO_QUALITY_1080P_60   = 116
	VIDEO_QUALITY_4K         = 120
	VIDEO_QUALITY_HDR        = 125
	VIDEO_QUALITY_DOLBY      = 126
	VIDEO_QUALITY_8K         = 127

	/*
		视频编码代码

		| 值 | 含义     | 备注           |
		| ---- | ---------- | ---------------- |
		| 7  | AVC 编码 | 8K 视频不支持该格式 |
		| 12 | HEVC 编码 |                |
		| 13 | AV1 编码 |                |
	*/
	VIDEO_CODECID_H264 = 0 //default
	VIDEO_CODECID_AVC  = 7
	VIDEO_CODECID_HEVC = 12
	VIDEO_CODECID_AV1  = 13

	AUDIO_CODECID = 0

	/*
		视频伴音音质代码

		| 值    | 含义 |
		| ----- | ---- |
		| 30216 | 64K  |
		| 30232 | 132K |
		| 30280 | 192K |
		| 30250 | 杜比全景声 |
		| 30251 | Hi-Res无损 |
	*/
	AUDIO_QUALITY_64K    = 30216 // 64k
	AUDIO_QUALITY_132K   = 30232
	AUDIO_QUALITY_DOLBY  = 30250
	AUDIO_QUALITY_HI_RES = 30251
	AUDIO_QUALITY_192K   = 30280
)

type Detecter struct {
	data *DownloadInfo
}

func NewDetecter(data *DownloadInfo) *Detecter {
	return &Detecter{data: data}
}

type Video struct {
	Codecid  int
	Quality  int
	Url      string
	MimeType string
}

type Audio struct {
	Quality int
	Url     string
}

type MediaStreams struct {
	Videos *[]Video
	Audios *[]Audio
}

func (d *Detecter) handleDash() *MediaStreams {
	videos := make([]Video, 0)
	audios := make([]Audio, 0)
	for _, video := range *d.data.Dash.Video {
		videos = append(videos, Video{
			Quality:  video.ID,
			Url:      video.BaseURL,
			MimeType: video.MimeType,
			Codecid:  video.Codecid,
		})
	}
	if d.data.Dash.Audio != nil {
		for _, audio := range *d.data.Dash.Audio {
			audios = append(audios, Audio{
				Quality: audio.ID,
				Url:     audio.BaseURL,
			})
		}
	}

	if d.data.Dash.Dolby != nil && d.data.Dash.Dolby.Audio != nil {
		for _, dolby := range *d.data.Dash.Dolby.Audio {
			audios = append(audios, Audio{
				Quality: dolby.ID,
				Url:     dolby.BaseURL,
			})
		}
	}
	if d.data.Dash.Flac != nil && d.data.Dash.Flac.Audio != nil {
		for _, flac := range *d.data.Dash.Flac.Audio {
			audios = append(audios, Audio{
				Quality: flac.ID,
				Url:     flac.BaseURL,
			})
		}
	}

	return &MediaStreams{
		Videos: &videos,
		Audios: &audios,
	}
}

func (d *Detecter) handleDurl() *MediaStreams {
	mimeType := "mp4"
	durl := *d.data.Durl
	videoUrl := durl[0].URL
	quality := d.data.Quality
	videos := make([]Video, 0)
	videos = append(videos, Video{
		Quality:  quality,
		Url:      videoUrl,
		MimeType: mimeType,
		Codecid:  VIDEO_CODECID_H264,
	})
	return &MediaStreams{Videos: &videos}
}

func (d *Detecter) Detect() *MediaStreams {
	if d.data.Durl != nil {
		return d.handleDurl()
	} else {
		return d.handleDash()
	}
}

type MediaStream struct {
	Video *Video
	Audio *Audio
}

// AVC 的url可能会403 forbidden
func (d *Detecter) DetectBest(codeCid int) *MediaStream {
	streams := d.Detect()
	var bestVideo *Video
	videoScores := 0
	if len(*streams.Videos) == 1 {
		bestVideo = &(*streams.Videos)[0]
	} else {
		for index, video := range *streams.Videos {
			if codeCid != 0 && codeCid != video.Codecid {
				continue
			}
			if (video.Quality + video.Codecid) > videoScores {
				videoScores = video.Quality + video.Codecid
				bestVideo = &(*streams.Videos)[index]
			}
		}
	}

	var bestAudio *Audio = nil
	if streams.Audios != nil {
		audioScores := 0
		for index, audio := range *streams.Audios {
			if audio.Quality > audioScores {
				audioScores = audio.Quality
				bestAudio = &(*streams.Audios)[index]
			}
		}
	}

	return &MediaStream{Video: bestVideo, Audio: bestAudio}
}
