package billitem

import (
	"bytes"
	"encoding/base64"
	"io"
	"time"
)

func getMonthDays(year, month int) []int {
	// 获取该月的最后一天
	lastDay := time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.UTC).Day()

	// 创建日期列表
	days := make([]int, lastDay)
	for day := 1; day <= int(lastDay); day++ {
		days[day-1] = day
	}

	return days
}

func getReader(str string) io.Reader {
	return base64.NewDecoder(base64.StdEncoding, bytes.NewReader([]byte(str)))
}
