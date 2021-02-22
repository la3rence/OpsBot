package utils

import (
	"fmt"
	"strconv"
	"strings"
)

func StringIndexOf(originalArray []string, wordToFind interface{}) []int {
	length := len(originalArray)
	interfaceArray := make([]interface{}, length)
	for i, v := range originalArray {
		interfaceArray[i] = v
	}
	var i = 0
	var indexArray []int
	for ; i < length; i++ {
		if strings.Compare(wordToFind.(string), originalArray[i]) == 0 {
			indexArray = append(indexArray, i)
		}
	}
	return indexArray
}

func StringToInt32(numberString string) (int32, error) {
	parseNumber, err := strconv.ParseInt(numberString, 10, 32)
	if err == nil {
		return int32(parseNumber), nil
	} else {
		return 0, err
	}
}

// 封装：获取 /tag 后一位字符串
func GetTagNextOneParam(originalMessage string, tagName string) (nextString string, err error) {
	wordArray := strings.Fields(originalMessage)
	indexes := StringIndexOf(wordArray, tagName)
	if len(wordArray) > indexes[0]+1 {
		nextString = wordArray[indexes[0]+1]
		return nextString, nil
	}
	return nextString, fmt.Errorf("param of %s required", tagName)
}
