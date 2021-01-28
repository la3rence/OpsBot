package utils

import (
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
