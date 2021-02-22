package utils

import (
	"fmt"
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

func GetTagNextOneParam(originalMessage string, tagName string) (nextString string, err error) {
	wordArray := strings.Fields(originalMessage)
	indexes := StringIndexOf(wordArray, tagName)
	if len(wordArray) > indexes[0]+1 {
		nextString = wordArray[indexes[0]+1]
		return nextString, nil
	}
	return nextString, fmt.Errorf("param of %s required", tagName)
}

func GetTagNextAllParams(originalMessage string, tagName string) (params []string) {
	wordArray := strings.Fields(originalMessage)
	tagIndexes := StringIndexOf(wordArray, tagName)
	for _, v := range tagIndexes {
		if v+1 < len(wordArray) {
			param := wordArray[v+1]
			params = append(params, param)
		}
	}
	return
}
