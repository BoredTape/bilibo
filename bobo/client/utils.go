// copy from https://github.com/CuteReimu/bilibili/blob/master/client.go
package client

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"strconv"

	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

func formatError(prefix string, code int64, message ...string) error {
	for _, m := range message {
		if len(m) > 0 {
			return errors.New(prefix + "失败，返回值：" + strconv.FormatInt(code, 10) + "，返回信息：" + m)
		}
	}
	return errors.New(prefix + "失败，返回值：" + strconv.FormatInt(code, 10))
}

func getRespData(resp *resty.Response, prefix string) ([]byte, error) {
	dataRaw, _, err := getRespDataWithCode(resp, prefix)
	return dataRaw, err
}

func getRespDataWithCode(resp *resty.Response, prefix string) ([]byte, int64, error) {
	var errorCode int64 = 0
	if resp.StatusCode() != 200 {
		respCode := resp.StatusCode()
		errorCode, _ = strconv.ParseInt(fmt.Sprintf("%d%d", 999, respCode), 10, 64)
		return nil, errorCode, errors.Errorf(prefix+"失败，status code: %d", resp.StatusCode())
	}
	if !gjson.ValidBytes(resp.Body()) {
		errorCode = 999
		return nil, errorCode, errors.New("json解析失败：" + resp.String())
	}
	res := gjson.ParseBytes(resp.Body())
	code := res.Get("code").Int()
	if code != 0 {
		return nil, code, formatError(prefix, code, res.Get("message").String(), res.Get("msg").String())
	}
	return []byte(res.Get("data").Raw), errorCode, nil
}

func getRespDataWithCheckWebi(resp *resty.Response, prefix string) ([]byte, bool, error) {
	if resp.StatusCode() != 200 {
		return nil, false, errors.Errorf(prefix+"失败，status code: %d", resp.StatusCode())
	}
	if !gjson.ValidBytes(resp.Body()) {
		return nil, false, errors.New("json解析失败：" + resp.String())
	}
	res := gjson.ParseBytes(resp.Body())
	code := res.Get("code").Int()
	if code != 0 {
		return nil, false, formatError(prefix, code, res.Get("message").String(), res.Get("msg").String())
	}
	wbiError := res.Get("data").Get("v_voucher").String()
	if wbiError != "" {
		return nil, true, nil
	}
	return []byte(res.Get("data").Raw), false, nil
}

func encrypt(publicKey, data string) (string, error) {
	// pem解码
	block, _ := pem.Decode([]byte(publicKey))
	if block == nil {
		return "", errors.New("failed to decode public key")
	}
	// x509解码
	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", errors.WithStack(err)
	}
	pk := publicKeyInterface.(*rsa.PublicKey)
	// 加密
	cipherText, err := rsa.EncryptPKCS1v15(rand.Reader, pk, []byte(data))
	if err != nil {
		return "", errors.WithStack(err)
	}
	// base64
	return base64.URLEncoding.EncodeToString(cipherText), nil
}
