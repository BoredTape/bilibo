package bili_client

import (
	"bilibo/consts"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/dustin/go-humanize"
)

type WriteCounter struct {
	Total uint64
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	wc.PrintProgress()
	return n, nil
}

func (wc WriteCounter) PrintProgress() {
	fmt.Printf("\r%s", strings.Repeat(" ", 35))
	fmt.Printf("\rDownloading... %s complete", humanize.Bytes(wc.Total))
}

func download(ua, url, filepath string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	c := http.Client{}
	req, _ := http.NewRequest("GET", url, nil)

	if ua == "" {
		ua = DEFAULT_UA
	}
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Referer", "https://www.bilibili.com")
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode == 403 {
		return consts.ERROR_DOWNLOAD_403
	} else if resp.StatusCode != 200 {
		return errors.New("download failed,status code: " + resp.Status)
	}
	defer resp.Body.Close()
	counter := &WriteCounter{}
	_, err = io.Copy(out, io.TeeReader(resp.Body, counter))
	if err != nil {
		return err
	}

	out.Close()
	fmt.Println("")
	return nil
}
