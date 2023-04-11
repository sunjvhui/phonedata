package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	INT_LEN            = 4
	HEAD_LENGTH        = 8
	PHONE_INDEX_LENGTH = 8
	PHONE_DAT          = "phone.dat"
)

var (
	content []byte
	region  = make(map[string]Data)
)

func get4(b []byte) int32 {
	if len(b) < 4 {
		return 0
	}
	return int32(b[0]) | int32(b[1])<<8 | int32(b[2])<<16 | int32(b[3])<<24
}

func Read() {
	var err error
	content, err = os.ReadFile(PHONE_DAT)
	if err != nil {
		return
	}
	totalLen := int32(len(content))

	// content需要截取加密信息
	ciphertextLen := get4(content[:INT_LEN])
	fmt.Println("ciphertextLen:", ciphertextLen)
	fmt.Println("ciphertextLen+INT_LEN:", ciphertextLen+INT_LEN)
	ciphertext := content[INT_LEN : ciphertextLen+INT_LEN]

	originalMd5 := AesDecryptCBC(ciphertext, InterimKey())
	fmt.Println("originalMd5:", string(originalMd5))
	fmt.Println("ciphertext:", ciphertext)

	newContent := content[INT_LEN+ciphertextLen:]
	h := md5.New()
	h.Write(newContent)
	retMd5 := hex.EncodeToString(h.Sum(nil))

	fmt.Println("originalMd5:", string(originalMd5))
	fmt.Println("retMd5:", retMd5)
	if string(originalMd5) != retMd5 {
		return
	}
	//panic(21)
	totalLen = int32(len(newContent))
	fmt.Println("total_len:", totalLen)
	firstOffset := get4(newContent[INT_LEN : INT_LEN*2])
	if firstOffset > totalLen {
		return
	}
	addressByte := newContent[firstOffset:]
	var address = strings.Split(string(addressByte), "\000")
	for _, v := range address {
		if v == "" {
			continue
		}
		data := strings.Split(v, "|")
		if len(data) < 4 {
			continue
		}
		offsetInt, _ := strconv.Atoi(data[2])
		count, _ := strconv.Atoi(data[3])
		regionMapKey := data[0] + data[1]
		region[regionMapKey] = Data{
			Province: data[0],
			City:     data[1],
			Offset:   offsetInt,
			Count:    count,
			PhoneNum: getNumber(newContent, offsetInt, count),
		}
		fmt.Println("regionMapKey:", regionMapKey)
	}
	fmt.Println("22", len(region))
	for s, index := range region {
		if s == "贵州贵阳" {
			fmt.Printf("PhoneNum:%+v\n", index.PhoneNum)
		}
	}

}

func getNumber(content []byte, offsetInt, count int) []uint32 {
	var phoneNumList []uint32
	if offsetInt > len(content) {
		return phoneNumList
	}
	numberOffsetData := content[offsetInt : offsetInt+count*4]
	//endOffset := int32(bytes.Index(numberOffsetData, []byte("\000")))
	//data := numberOffsetData[:endOffset]

	//count := len(data) / 4
	for i := 0; i < count; i++ {
		offset := i * 4
		number := get4(numberOffsetData[offset : offset+4])
		phoneNumList = append(phoneNumList, uint32(number))
	}

	//number := bytes.Split(numberOffsetData[:endOffset], []byte("|"))
	//for _, n := range number {
	//
	//	offsetInt, _ := strconv.Atoi(string(n))
	//	phoneNumList = append(phoneNumList, uint32(offsetInt))
	//}
	return phoneNumList
}
