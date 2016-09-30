package qqwry

import (
	"encoding/binary"
	"errors"
	"io/ioutil"
	"net"
	"os"
	"strings"

	"github.com/axgle/mahonia"
)

/*------------------------------------------------------------------------------
public
------------------------------------------------------------------------------*/

const (
	INDEX_LEN       = 7    // 索引长度
	REDIRECT_MODE_1 = 0x01 // 国家的类型, 指向另一个指向
	REDIRECT_MODE_2 = 0x02 // 国家的类型, 指向一个指向
)

type ResultQQwry struct {
	IP      string `json:"ip"`
	Country string `json:"country"`
	Area    string `json:"area"`
}

type QQWry interface {
	Query(ip string) (ResultQQwry, error)
}

var ERR_QQWRY_ALREADY_INIT error
var ERR_QQWRY_INVALID_IP error
var ERR_QQWRY_NOT_FOUND error
var ERR_QQWRY_NOT_INIT error

/*------------------------------------------------------------------------------
private
------------------------------------------------------------------------------*/

type fileData struct {
	data  []byte
	ipNum int64
}

type qqwry struct {
	ipdata *fileData
	offset int64
}

var ipData *fileData

func init() {
	ERR_QQWRY_ALREADY_INIT = errors.New("qqwry aleady init")
	ERR_QQWRY_NOT_FOUND = errors.New("qqwry not found")
	ERR_QQWRY_INVALID_IP = errors.New("qqwry invalid ip")
	ERR_QQWRY_NOT_INIT = errors.New("qqwry not init")
}

// 初始化ip库数据到内存中
func Init(qqwryDatPath string) error {
	if ipData != nil {
		return ERR_QQWRY_ALREADY_INIT
	}

	// 判断文件是否存在
	_, err := os.Stat(qqwryDatPath)
	if err != nil && os.IsNotExist(err) {
		return err
	}

	var fd fileData

	tmpData, err := ioutil.ReadFile(qqwryDatPath)
	if err != nil {
		return err
	}
	fd.data = tmpData

	buf := tmpData[0:8]
	start := binary.LittleEndian.Uint32(buf[:4])
	end := binary.LittleEndian.Uint32(buf[4:])

	fd.ipNum = int64((end-start)/INDEX_LEN + 1)

	ipData = &fd
	return nil
}

// 新建 qqwry  类型
func NewQQwry() (QQWry, error) {
	if ipData == nil {
		return nil, ERR_QQWRY_NOT_INIT
	}
	return &qqwry{
		ipdata: ipData,
	}, nil
}

func (q *qqwry) Query(ip string) (ResultQQwry, error) {
	res := ResultQQwry{}

	if q.ipdata == nil {
		return res, ERR_QQWRY_NOT_INIT
	}

	res.IP = ip
	if strings.Count(ip, ".") != 3 {
		return res, ERR_QQWRY_INVALID_IP
	}
	offset := q.searchIndex(binary.BigEndian.Uint32(net.ParseIP(ip).To4()))
	if offset <= 0 {
		return res, ERR_QQWRY_NOT_FOUND
	}

	var country []byte
	var area []byte

	mode := q.readMode(offset + 4)
	if mode == REDIRECT_MODE_1 {
		countryOffset := q.readUInt24()
		mode = q.readMode(countryOffset)
		if mode == REDIRECT_MODE_2 {
			c := q.readUInt24()
			country = q.readString(c)
			countryOffset += 4
		} else {
			country = q.readString(countryOffset)
			countryOffset += uint32(len(country) + 1)
		}
		area = q.readArea(countryOffset)
	} else if mode == REDIRECT_MODE_2 {
		countryOffset := q.readUInt24()
		country = q.readString(countryOffset)
		area = q.readArea(offset + 8)
	} else {
		country = q.readString(offset + 4)
		area = q.readArea(offset + uint32(5+len(country)))
	}

	enc := mahonia.NewDecoder("gbk")
	res.Country = enc.ConvertString(string(country))
	res.Area = enc.ConvertString(string(area))
	return res, nil
}

func (q *qqwry) readMode(offset uint32) byte {
	mode := q.readData(1, int64(offset))
	return mode[0]
}

func (q *qqwry) readArea(offset uint32) []byte {
	mode := q.readMode(offset)
	if mode == REDIRECT_MODE_1 || mode == REDIRECT_MODE_2 {
		areaOffset := q.readUInt24()
		if areaOffset == 0 {
			return []byte("")
		} else {
			return q.readString(areaOffset)
		}
	} else {
		return q.readString(offset)
	}
	return []byte("")
}

func (q *qqwry) readString(offset uint32) []byte {
	q.setOffset(int64(offset))
	data := make([]byte, 0, 30)
	buf := make([]byte, 1)
	for {
		buf = q.readData(1)
		if buf[0] == 0 {
			break
		}
		data = append(data, buf[0])
	}
	return data
}

func (q *qqwry) searchIndex(ip uint32) uint32 {
	header := q.readData(8, 0)

	start := binary.LittleEndian.Uint32(header[:4])
	end := binary.LittleEndian.Uint32(header[4:])

	buf := make([]byte, INDEX_LEN)
	mid := uint32(0)
	_ip := uint32(0)

	for {
		mid = q.getMiddleOffset(start, end)
		buf = q.readData(INDEX_LEN, int64(mid))
		_ip = binary.LittleEndian.Uint32(buf[:4])

		if end-start == INDEX_LEN {
			offset := byteToUInt32(buf[4:])
			buf = q.readData(INDEX_LEN)
			if ip < binary.LittleEndian.Uint32(buf[:4]) {
				return offset
			} else {
				return 0
			}
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
	return 0
}

func (q *qqwry) readUInt24() uint32 {
	buf := q.readData(3)
	return byteToUInt32(buf)
}

func (q *qqwry) getMiddleOffset(start uint32, end uint32) uint32 {
	records := ((end - start) / INDEX_LEN) >> 1
	return start + records*INDEX_LEN
}

// 将 byte 转换为uint32
func byteToUInt32(data []byte) uint32 {
	i := uint32(data[0]) & 0xff
	i |= (uint32(data[1]) << 8) & 0xff00
	i |= (uint32(data[2]) << 16) & 0xff0000
	return i
}

func (q *qqwry) readData(num int, offset ...int64) (rs []byte) {
	if len(offset) > 0 {
		q.setOffset(offset[0])
	}
	nums := int64(num)
	end := q.offset + nums
	dataNum := int64(len(q.ipdata.data))
	if q.offset > dataNum {
		return nil
	}

	if end > dataNum {
		end = dataNum
	}
	rs = q.ipdata.data[q.offset:end]
	q.offset = end
	return
}

// 设置偏移量
func (q *qqwry) setOffset(offset int64) {
	q.offset = offset
}
