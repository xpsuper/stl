package stl

import (
	"strings"
	"time"
)

// XPDateTimeImpl 日期时间工具对象
type XPDateTimeImpl struct {
}

const (
	yyyy = "2006"
	yy   = "06"
	mmmm = "January"
	mmm  = "Jan"
	mm   = "01"
	dddd = "Monday"
	ddd  = "Mon"
	dd   = "02"

	HHT = "03"
	HH  = "15"
	MM  = "04"
	SS  = "05"
	ss  = "05"
	III = "000"
	iii = "000"
	tt  = "PM"
	Z   = "MST"
	ZZZ = "MST"
)

func convertFormat(format string) string {
	var goFormat = format
	if strings.Contains(goFormat, "YYYY") {
		goFormat = strings.Replace(goFormat, "YYYY", yyyy, -1)
	} else if strings.Contains(goFormat, "yyyy") {
		goFormat = strings.Replace(goFormat, "yyyy", yyyy, -1)
	} else if strings.Contains(goFormat, "YY") {
		goFormat = strings.Replace(goFormat, "YY", yy, -1)
	} else if strings.Contains(goFormat, "yy") {
		goFormat = strings.Replace(goFormat, "yy", yy, -1)
	}

	if strings.Contains(goFormat, "MMMM") {
		goFormat = strings.Replace(goFormat, "MMMM", mmmm, -1)
	} else if strings.Contains(goFormat, "mmmm") {
		goFormat = strings.Replace(goFormat, "mmmm", mmmm, -1)
	} else if strings.Contains(goFormat, "MMM") {
		goFormat = strings.Replace(goFormat, "MMM", mmm, -1)
	} else if strings.Contains(goFormat, "mmm") {
		goFormat = strings.Replace(goFormat, "mmm", mmm, -1)
	} else if strings.Contains(goFormat, "mm") {
		goFormat = strings.Replace(goFormat, "mm", mm, -1)
	}

	if strings.Contains(goFormat, "dddd") {
		goFormat = strings.Replace(goFormat, "dddd", dddd, -1)
	} else if strings.Contains(goFormat, "ddd") {
		goFormat = strings.Replace(goFormat, "ddd", ddd, -1)
	} else if strings.Contains(goFormat, "dd") {
		goFormat = strings.Replace(goFormat, "dd", dd, -1)
	}

	if strings.Contains(goFormat, "tt") {
		if strings.Contains(goFormat, "HH") {
			goFormat = strings.Replace(goFormat, "HH", HHT, -1)
		} else if strings.Contains(goFormat, "hh") {
			goFormat = strings.Replace(goFormat, "hh", HHT, -1)
		}
		goFormat = strings.Replace(goFormat, "tt", tt, -1)
	} else {
		if strings.Contains(goFormat, "HH") {
			goFormat = strings.Replace(goFormat, "HH", HH, -1)
		} else if strings.Contains(goFormat, "hh") {
			goFormat = strings.Replace(goFormat, "hh", HH, -1)
		}
		goFormat = strings.Replace(goFormat, "tt", "", -1)
	}

	if strings.Contains(goFormat, "MM") {
		goFormat = strings.Replace(goFormat, "MM", MM, -1)
	}

	if strings.Contains(goFormat, "SS") {
		goFormat = strings.Replace(goFormat, "SS", SS, -1)
	} else if strings.Contains(goFormat, "ss") {
		goFormat = strings.Replace(goFormat, "ss", ss, -1)
	}

	if strings.Contains(goFormat, "III") {
		goFormat = strings.Replace(goFormat, "III", III, -1)
	} else if strings.Contains(goFormat, "iii") {
		goFormat = strings.Replace(goFormat, "iii", iii, -1)
	}

	if strings.Contains(goFormat, "ZZZ") {
		goFormat = strings.Replace(goFormat, "ZZZ", ZZZ, -1)
	} else if strings.Contains(goFormat, "zzz") {
		goFormat = strings.Replace(goFormat, "zzz", ZZZ, -1)
	} else if strings.Contains(goFormat, "Z") {
		goFormat = strings.Replace(goFormat, "Z", Z, -1)
	} else if strings.Contains(goFormat, "z") {
		goFormat = strings.Replace(goFormat, "z", Z, -1)
	}

	if strings.Contains(goFormat, "tt") {
		goFormat = strings.Replace(goFormat, "tt", tt, -1)
	}

	return goFormat
}

// StringToTime 按照默认的 yyyy-mm-dd HH:MM:SS 格式将字符串转换为时间
func (instance *XPDateTimeImpl) StringToTime(str string) time.Time {
	t, _ := time.ParseInLocation("2006-01-02 15:04:05", str, time.Local)
	return t
}

// TimeToString 按照默认的 yyyy-mm-dd HH:MM:SS 格式将时间转换为字符串
func (instance *XPDateTimeImpl) TimeToString(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// StringToTimeFmt 将字符串转换为时间, 需自定义格式
func (instance *XPDateTimeImpl) StringToTimeFmt(str, fmt string) time.Time {
	t, _ := time.ParseInLocation(convertFormat(fmt), str, time.Local)
	return t
}

// TimeToStringFmt 将时间转换为字符串, 需自定义格式
func (instance *XPDateTimeImpl) TimeToStringFmt(t time.Time, fmt string) string {
	return t.Format(convertFormat(fmt))
}

// Timestamp 获取时间戳(到秒共10位)
func (instance *XPDateTimeImpl) Timestamp() int64 {
	return time.Now().Unix()
}

// TimestampMillisecond 获取时间戳(到毫秒共13位)
func (instance *XPDateTimeImpl) TimestampMillisecond() int64 {
	return time.Now().UnixNano() / 1000000
}
