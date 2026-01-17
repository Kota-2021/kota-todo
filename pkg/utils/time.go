package utils

import (
	"time"
)

var (
	// JST はAsia/Tokyoタイムゾーンを表します
	JST *time.Location
)

func init() {
	var err error
	JST, err = time.LoadLocation("Asia/Tokyo")
	if err != nil {
		// タイムゾーンデータが利用できない場合はUTCを使用
		JST = time.UTC
	}
}

// NowJST は現在時刻をJST（Asia/Tokyo）で返します
func NowJST() time.Time {
	return time.Now().In(JST)
}
