package consts

import "errors"

var ERROR_DOWNLOAD_403 = errors.New("download failed,status code: 403")

const (
	FAVOUR_NOT_SYNC  = 0
	FAVOUR_NEED_SYNC = 1
)

// 任务类型
const (
	TASK_TYPE_FAVOUR   = 1
	TASK_TYPE_DOWNLOAD = 2
)

const (
	VIDEO_STATUS_INIT           = -1
	VIDEO_STATUS_TO_BE_DOWNLOAD = 0
	VIDEO_STATUS_DOWNLOADING    = 1
	VIDEO_STATUS_DOWNLOAD_DONE  = 2
	VIDEO_STATUS_DOWNLOAD_FAIL  = 3
	VIDEO_STATUS_DOWNLOAD_RETRY = 4
)

const (
	QRCODE_STATUS_NOT_SCAN = 1
	QRCODE_STATUS_SCANNED  = 2
	QRCODE_STATUS_INVALID  = 3
)

const (
	ACCOUNT_STATUS_NORMAL    = 0
	ACCOUNT_STATUS_NOT_LOGIN = 1
	ACCOUNT_STATUS_INVALID   = 2
)

const (
	VIDEO_MESSAGE_ERROR   = 999
	VIDEO_MESSAGE_SUCCESS = 0
)
