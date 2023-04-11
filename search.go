package main

import (
	"fmt"
	"strconv"
)

func TestFindNum() {
	c := "福建厦门"
	phoneNum := "1999684"

	offsetInt, _ := strconv.Atoi(phoneNum[0:7])
	data := region[c].PhoneNum
	fmt.Println("data:", data)
	if FindNum(data, uint32(offsetInt)) {
		fmt.Printf("号码%s地址是%s", phoneNum, c)
	}
}

func FindNum(arrInt []uint32, number uint32) bool {
	return find(arrInt, 0, len(arrInt)-1, number)
}

func find(arrInt []uint32, leftPtr int, rightPtr int, findNum uint32) bool {
	if leftPtr > rightPtr {
		return false
	}
	//先求出中间的指针位置
	mPtr := (leftPtr + rightPtr) / 2

	if (arrInt)[mPtr] > findNum {
		//递归调用,从mPtr ---> leftPtr之间查找
		return find(arrInt, leftPtr, mPtr-1, findNum)
	} else if ((arrInt)[mPtr]) < findNum {
		//递归调用，从leftPtr + 1 ---->  rightPtr之间查找
		return find(arrInt, mPtr+1, rightPtr, findNum)
	} else {
		//当leftPtr == rightPtr，则找到该数值
		return true
	}
}
