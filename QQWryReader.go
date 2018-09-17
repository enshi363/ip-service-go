package main

import (
	"bytes"
	"encoding/binary"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"regexp"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

const (
	// INDEXLEN 索引长度
	INDEXLEN = 7
	// REDIRECTMODE1 重定向模式1
	REDIRECTMODE1 = 0x01
	// REDIRECTMODE2 重定向模式2
	REDIRECTMODE2 = 0x02
)

// QQwry 纯真库结构
type QQwry struct {
	filepath string
	file     *bytes.Reader
}

// IPResult 返回结果
type IPResult struct {
	IP      string
	Country string
	Area    string
}

// GbkToUtf8 gbk to utf8
func GbkToUtf8(s []byte) []byte {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		// return nil, e
	}
	return d
}

var content = []byte{}

//NewQQwry 初始化
func NewQQwry(file string) (qqwry *QQwry) {
	qqwry = &QQwry{filepath: file}

	return qqwry
}

// LoadIPData 加载数据到内存
func (q *QQwry) LoadIPData() {
	matched, _ := regexp.MatchString("http[s]{0,1}://.*", q.filepath)
	if matched == true {
		log.Println("http 地址")
		resp, err := http.Get(q.filepath)
		if err != nil {
			log.Println(err)
			return
		}
		defer resp.Body.Close()
		content, _ = ioutil.ReadAll(resp.Body)
		q.file = bytes.NewReader(content)
	} else {
		c, err := ioutil.ReadFile(q.filepath)
		if err != nil {
			log.Println(err)
			return
		}
		content = c
		c = nil
		q.file = bytes.NewReader(content)
	}

}

// Find 获取ip信息
func (q *QQwry) Find(ip string) *IPResult {
	// q.IP = ip
	var parsedIP = (net.ParseIP(ip).To4())
	var result = &IPResult{IP: string(parsedIP)}
	if parsedIP == nil {
		return result
	}

	offset := q.searchIndex(binary.BigEndian.Uint32(parsedIP))
	// log.Println("loc offset:", offset)
	if offset <= 0 {
		return result
	}
	var country []byte
	var area []byte

	mode := q.readMode(offset + 4)
	// log.Println("mode", mode)
	if mode == REDIRECTMODE1 {
		countryOffset := q.readUInt24()
		mode = q.readMode(countryOffset)
		// log.Println("1 - mode", mode)
		if mode == REDIRECTMODE2 {
			c := q.readUInt24()
			country = q.readString(c)
			countryOffset += 4
		} else {
			country = q.readString(countryOffset)
			countryOffset += uint32(len(country) + 1)
		}
		area = q.readArea(countryOffset)
	} else if mode == REDIRECTMODE2 {
		countryOffset := q.readUInt24()
		country = q.readString(countryOffset)
		area = q.readArea(offset + 8)
	} else {
		country = q.readString(offset + 4)
		area = q.readArea(offset + uint32(5+len(country)))
	}

	result.Country = string(GbkToUtf8(country))
	result.Area = string(GbkToUtf8(area))
	return result

}

func (q *QQwry) readMode(offset uint32) byte {
	q.file.Seek(int64(offset), 0)
	mode := make([]byte, 1)
	q.file.Read(mode)
	return mode[0]
}

func (q *QQwry) readArea(offset uint32) []byte {
	mode := q.readMode(offset)
	areaOffset := q.readUInt24()
	if mode == REDIRECTMODE1 || mode == REDIRECTMODE2 {
		if areaOffset == 0 {
			return []byte("")
		}
	} else {
		return q.readString(offset)
	}
	return q.readString(areaOffset)
	// return []byte("")
}

func (q *QQwry) readString(offset uint32) []byte {
	q.file.Seek(int64(offset), 0)
	data := make([]byte, 0, 30)
	buf := make([]byte, 1)
	for {
		q.file.Read(buf)
		if buf[0] == 0 {
			break
		}
		data = append(data, buf[0])
	}
	return data
}

func (q *QQwry) searchIndex(ip uint32) uint32 {
	header := make([]byte, 8)
	q.file.Seek(0, 0)
	q.file.Read(header)

	start := binary.LittleEndian.Uint32(header[:4])
	end := binary.LittleEndian.Uint32(header[4:])

	// log.Printf("len info %v, %v ---- %v, %v", start, end, hex.EncodeToString(header[:4]), hex.EncodeToString(header[4:]))

	for {
		mid := q.getMiddleOffset(start, end)
		q.file.Seek(int64(mid), 0)
		buf := make([]byte, INDEXLEN)
		q.file.Read(buf)
		_ip := binary.LittleEndian.Uint32(buf[:4])

		// log.Printf(">> %v, %v, %v -- %v", start, mid, end, hex.EncodeToString(buf[:4]))

		if end-start == INDEXLEN {
			offset := byte3ToUInt32(buf[4:])
			q.file.Read(buf)
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
			return byte3ToUInt32(buf[4:])
		}

	}
}

func (q *QQwry) readUInt24() uint32 {
	buf := make([]byte, 3)
	q.file.Read(buf)
	return byte3ToUInt32(buf)
}

func (q *QQwry) getMiddleOffset(start uint32, end uint32) uint32 {
	records := ((end - start) / INDEXLEN) >> 1
	return start + records*INDEXLEN
}

func byte3ToUInt32(data []byte) uint32 {
	i := uint32(data[0]) & 0xff
	i |= (uint32(data[1]) << 8) & 0xff00
	i |= (uint32(data[2]) << 16) & 0xff0000
	return i
}
