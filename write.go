package main

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/xuri/excelize/v2"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Data struct {
	PhoneNum []uint32
	Province string
	City     string
	Offset   int
	Count    int
}

func (d Data) marshal() []byte {
	var bs = make([]byte, len(d.PhoneNum)*4)
	for i, v := range d.PhoneNum {
		binary.LittleEndian.PutUint32(bs[i*4:], v)
	}

	return bs
}

type Index struct {
	Province string
	City     string
	Offset   uint32
	Count    uint32
	//CardType byte
}

func (i Index) marshal() []byte {
	var bs []string
	bs = append(bs, i.Province, i.City, fmt.Sprint(i.Offset), fmt.Sprint(i.Count))
	return []byte(strings.Join(bs, "|") + "\000")
}

//var datas []*Data

var dataMap = make(map[string]*Data)

func initData(url string) {

	f, err := excelize.OpenFile(url)
	if err != nil {
		fmt.Println("err:", err.Error())
		panic("url can't find")
		return
	}

	// 获取工作表中指定单元格的值
	// 获取 Sheet1 上所有单元格
	rows, err := f.GetRows("sheet1")
	//fmt.Println("rows:", rows)
	for i, row := range rows {
		// 小于三个的不要 || 第一行是表头也不要
		if len(row) < 3 || i == 0 {
			continue
		}
		//fmt.Println("row[1]+row[2]]:", row[1]+row[2])
		phoneNum, _ := strconv.Atoi(row[0])
		if _, ok := dataMap[row[1]+row[2]]; ok {
			dataMap[row[1]+row[2]].PhoneNum = append(dataMap[row[1]+row[2]].PhoneNum, uint32(phoneNum))
		} else {
			var sub = &Data{
				PhoneNum: []uint32{uint32(phoneNum)},
				Province: row[1],
				City:     row[2],
			}
			dataMap[row[1]+row[2]] = sub
		}
	}
	for _, data := range dataMap {
		sort.Slice(data.PhoneNum, func(i, j int) bool {
			return data.PhoneNum[i] < data.PhoneNum[j]
		})
	}
}

func Write() {

	initData("phonelist.xlsx")
	fmt.Println("date init end")
	var (
		total uint32
		bs    = make([]byte, 10485760)
		data  bytes.Buffer
		index bytes.Buffer
	)

	var offset = 8
	for _, d := range dataMap {
		ds := d.marshal()
		data.Write(ds)
		d.Offset = offset
		d.Count = len(d.PhoneNum)
		offset += len(ds)
	}

	for _, d := range dataMap {
		i := Index{d.Province, d.City, uint32(d.Offset), uint32(d.Count)}
		index.Write(i.marshal())
	}

	// 版本
	binary.LittleEndian.PutUint32(bs, uint32(2301))
	total += 4
	// 第一个索引开始的位置
	binary.LittleEndian.PutUint32(bs[4:], uint32(offset))
	total += 4
	// 写入数据
	copy(bs[8:], data.Bytes())
	total += uint32(data.Len())
	// 写入索引
	fmt.Println("idnex:", index.Len())
	copy(bs[8+data.Len():], index.Bytes())
	total += uint32(index.Len())

	h := md5.New()
	h.Write(bs[:total])
	retMd5 := hex.EncodeToString(h.Sum(nil))
	// 获取到md5
	encrypted := AesEncryptCBC([]byte(retMd5), InterimKey())
	fmt.Println("1retMd5:", retMd5)
	fmt.Println("encrypted:", encrypted)

	var newBs = make([]byte, 4+uint32(len(encrypted)))
	newTotal := total
	binary.LittleEndian.PutUint32(newBs, uint32(len(encrypted)))
	newTotal += 4

	copy(newBs[4:], encrypted)
	newTotal += uint32(len(encrypted))

	newBs = append(newBs, bs[:total]...)
	err := os.WriteFile(PHONE_DAT, newBs[:newTotal], 0755)
	if err != nil {
		fmt.Println("2222")
		return
	}
}
