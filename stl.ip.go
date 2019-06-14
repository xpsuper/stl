package stl

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/simplifiedchinese"
)

var (
	KeyFileUrl  = "http://update.cz88.net/ip/copywrite.rar"
	DataFileUrl = "http://update.cz88.net/ip/qqwry.rar"
)

const (
	// IndexLen 索引长度
	IndexLen = 7
	// RedirectMode1 国家的类型, 指向另一个指向
	RedirectMode1 = 0x01
	// RedirectMode2 国家的类型, 指向一个指向
	RedirectMode2 = 0x02
)

type XPIPImpl struct {
	Data   *fileData
	Offset int64
}

// 归属地信息
type IPAddressInfo struct {
	IP      string `json:"ip"`
	Country string `json:"country"`
	Area    string `json:"area"`
}

type fileData struct {
	Data     []byte
	FilePath string
	Path     *os.File
	IPNum    int64
}

var IPData fileData

// InitIPData 初始化ip库数据到内存中
func (f *fileData) InitIPData() (rs interface{}) {
	var tmpData []byte

	// 判断文件是否存在
	_, err := os.Stat(f.FilePath)
	if err != nil && os.IsNotExist(err) {
		log.Println("文件不存在，尝试从网络获取最新纯真 IP 库")
		tmpData, err = getOnline()
		if err != nil {
			rs = err
			return
		} else {
			if err := ioutil.WriteFile(f.FilePath, tmpData, 0644); err == nil {
				log.Printf("已将最新的纯真 IP 库保存到本地 %s ", f.FilePath)
			}
		}
	} else {
		// 打开文件句柄
		log.Printf("从本地数据库文件 %s 打开\n", f.FilePath)
		f.Path, err = os.OpenFile(f.FilePath, os.O_RDONLY, 0400)
		if err != nil {
			rs = err
			return
		}
		defer f.Path.Close()

		tmpData, err = ioutil.ReadAll(f.Path)
		if err != nil {
			log.Println(err)
			rs = err
			return
		}
	}

	f.Data = tmpData

	buf := f.Data[0:8]
	start := binary.LittleEndian.Uint32(buf[:4])
	end := binary.LittleEndian.Uint32(buf[4:])

	f.IPNum = int64((end-start)/IndexLen + 1)

	return true
}

// NewIPAddress 纯真数据库解析器
func NewIPAddress() *XPIPImpl {
	return &XPIPImpl{
		Data: &IPData,
	}
}

func (instance *XPIPImpl) InitData(filePath string) (rs interface{}) {
	instance.Data.FilePath = filePath
	return instance.Data.InitIPData()
}

func (instance *XPIPImpl) Validate(ipAddress string) bool {
	r := &XPRegexpImpl{}
	return r.IsMatchString(`^((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$`, ipAddress)
}

// 将ip地址转换为长整型
func (instance *XPIPImpl) IP2Long(ipAddress string) int64 {
	if ip := net.ParseIP(ipAddress); ip != nil {
		var n uint32
		ipBytes := ip.To4()
		for i := uint8(0); i <= 3; i++ {
			n |= uint32(ipBytes[i]) << ((3 - i) * 8)
		}
		return int64(n)
	}
	return 0
}

// 将长整型转换为ip地址
func (instance *XPIPImpl) Long2IP(long int64) string {
	ipBytes := net.IP{}
	for i := uint(0); i <= 3; i++ {
		ipBytes = append(ipBytes, byte(long>>((3-i)*8)))
	}
	return ipBytes.String()
}

func (instance *XPIPImpl) IP2Domain(ipAddress string) (string, error) {
	domain, err := net.LookupAddr(ipAddress)
	if domain != nil {
		return strings.TrimRight(domain[0], "."), nil
	}
	return "", err
}

func (instance *XPIPImpl) Domain2IP(domain string) ([]string, error) {
	ips, err := net.LookupIP(domain)
	if ips != nil {
		var items []string
		for _, v := range ips {
			if v.To4() != nil {
				items = append(items, v.String())
			}
		}
		return items, nil
	}
	return nil, err
}

// 判断所给ip是否为局域网ip
// A类 10.0.0.0--10.255.255.255
// B类 172.16.0.0--172.31.255.255
// C类 192.168.0.0--192.168.255.255
func (instance *XPIPImpl) IsIntranet(ipStr string) bool {
	// ip协议保留的局域网ip
	if strings.HasPrefix(ipStr, "10.") || strings.HasPrefix(ipStr, "192.168.") {
		return true
	}
	if strings.HasPrefix(ipStr, "172.") {
		// 172.16.0.0 - 172.31.255.255
		arr := strings.Split(ipStr, ".")
		if len(arr) != 4 {
			return false
		}

		second, err := strconv.ParseInt(arr[1], 10, 64)
		if err != nil {
			return false
		}

		if second >= 16 && second <= 31 {
			return true
		}
	}

	return false
}

// 获取本地局域网ip列表
func (instance *XPIPImpl) IntranetIP() (ips []string, err error) {
	ips        = make([]string, 0)
	inters, e := net.Interfaces()
	if e != nil {
		return ips, e
	}
	for _, inter := range inters {
		if inter.Flags&net.FlagUp == 0 {
			continue // interface down
		}

		if inter.Flags & net.FlagLoopback != 0 {
			continue // loopback interface
		}

		// ignore warden bridge
		if strings.HasPrefix(inter.Name, "w-") {
			continue
		}

		addr, e := inter.Addrs()
		if e != nil {
			return ips, e
		}

		for _, addr := range addr {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}

			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}

			ipStr := ip.String()
			if instance.IsIntranet(ipStr) {
				ips = append(ips, ipStr)
			}
		}
	}
	return ips, nil
}

// Find ip地址查询对应归属地信息
func (instance *XPIPImpl) Find(ip string) (res IPAddressInfo) {

	res = IPAddressInfo{}

	res.IP = ip
	if strings.Count(ip, ".") != 3 {
		return res
	}
	offset := instance.searchIndex(binary.BigEndian.Uint32(net.ParseIP(ip).To4()))
	if offset <= 0 {
		return
	}

	var country []byte
	var area []byte

	mode := instance.readMode(offset + 4)
	if mode == RedirectMode1 {
		countryOffset := instance.readUInt24()
		mode = instance.readMode(countryOffset)
		if mode == RedirectMode2 {
			c := instance.readUInt24()
			country = instance.readString(c)
			countryOffset += 4
		} else {
			country = instance.readString(countryOffset)
			countryOffset += uint32(len(country) + 1)
		}
		area = instance.readArea(countryOffset)
	} else if mode == RedirectMode2 {
		countryOffset := instance.readUInt24()
		country = instance.readString(countryOffset)
		area = instance.readArea(offset + 8)
	} else {
		country = instance.readString(offset + 4)
		area = instance.readArea(offset + uint32(5+len(country)))
	}

	enc := simplifiedchinese.GBK.NewDecoder()
	res.Country, _ = enc.String(string(country))
	res.Area, _ = enc.String(string(area))

	return
}

// ReadData 从文件中读取数据
func (instance *XPIPImpl) readData(num int, offset ...int64) (rs []byte) {
	if len(offset) > 0 {
		instance.setOffset(offset[0])
	}
	nums := int64(num)
	end := instance.Offset + nums
	dataNum := int64(len(instance.Data.Data))
	if instance.Offset > dataNum {
		return nil
	}

	if end > dataNum {
		end = dataNum
	}
	rs = instance.Data.Data[instance.Offset:end]
	instance.Offset = end
	return
}

// SetOffset 设置偏移量
func (instance *XPIPImpl) setOffset(offset int64) {
	instance.Offset = offset
}

// readMode 获取偏移值类型
func (instance *XPIPImpl) readMode(offset uint32) byte {
	mode := instance.readData(1, int64(offset))
	return mode[0]
}

// readArea 读取区域
func (instance *XPIPImpl) readArea(offset uint32) []byte {
	mode := instance.readMode(offset)
	if mode == RedirectMode1 || mode == RedirectMode2 {
		areaOffset := instance.readUInt24()
		if areaOffset == 0 {
			return []byte("")
		}
		return instance.readString(areaOffset)
	}
	return instance.readString(offset)
}

// readString 获取字符串
func (instance *XPIPImpl) readString(offset uint32) []byte {
	instance.setOffset(int64(offset))
	data := make([]byte, 0, 30)
	buf := make([]byte, 1)
	for {
		buf = instance.readData(1)
		if buf[0] == 0 {
			break
		}
		data = append(data, buf[0])
	}
	return data
}

// searchIndex 查找索引位置
func (instance *XPIPImpl) searchIndex(ip uint32) uint32 {
	header := instance.readData(8, 0)

	start := binary.LittleEndian.Uint32(header[:4])
	end := binary.LittleEndian.Uint32(header[4:])

	buf := make([]byte, IndexLen)
	mid := uint32(0)
	_ip := uint32(0)

	for {
		mid = instance.getMiddleOffset(start, end)
		buf = instance.readData(IndexLen, int64(mid))
		_ip = binary.LittleEndian.Uint32(buf[:4])

		if end-start == IndexLen {
			offset := byteToUInt32(buf[4:])
			buf = instance.readData(IndexLen)
			if ip < binary.LittleEndian.Uint32(buf[:4]) {
				return offset
			}
			return 0
		}

		// 找到的比较大，向前移
		if _ip > ip {
			end = mid
		} else if _ip < ip { // 找到的比较小，向后移
			start = mid
		} else if _ip == ip {
			return byteToUInt32(buf[4:])
		}
	}
}

// readUInt24
func (instance *XPIPImpl) readUInt24() uint32 {
	buf := instance.readData(3)
	return byteToUInt32(buf)
}

// getMiddleOffset
func (instance *XPIPImpl) getMiddleOffset(start uint32, end uint32) uint32 {
	records := ((end - start) / IndexLen) >> 1
	return start + records*IndexLen
}

// byteToUInt32 将 byte 转换为uint32
func byteToUInt32(data []byte) uint32 {
	i := uint32(data[0]) & 0xff
	i |= (uint32(data[1]) << 8) & 0xff00
	i |= (uint32(data[2]) << 16) & 0xff0000
	return i
}

func getKey() (uint32, error) {
	resp, err := http.Get(KeyFileUrl)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if body, err := ioutil.ReadAll(resp.Body); err != nil {
		return 0, err
	} else {
		return binary.LittleEndian.Uint32(body[5*4:]), nil
	}
}

func getOnline() ([]byte, error) {
	resp, err := http.Get(DataFileUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if body, err := ioutil.ReadAll(resp.Body); err != nil {
		return nil, err
	} else {
		if key, err := getKey(); err != nil {
			return nil, err
		} else {
			for i := 0; i < 0x200; i++ {
				key = key * 0x805
				key++
				key = key & 0xff

				body[i] = byte(uint32(body[i]) ^ key)
			}

			reader, err := zlib.NewReader(bytes.NewReader(body))
			if err != nil {
				return nil, err
			}

			return ioutil.ReadAll(reader)
		}
	}
}
