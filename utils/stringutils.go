package utils

import (
	"errors"
	"fmt"
	"strings"
)

func StringIndexOf(originalArray []string, wordToFind interface{}) []int {
	// Validate input
	if originalArray == nil {
		return []int{}
	}

	word, ok := wordToFind.(string)
	if !ok {
		return []int{}
	}

	var indexes []int
	for i, v := range originalArray {
		if strings.Compare(word, v) == 0 {
			indexes = append(indexes, i)
		}
	}
	return indexes
}

func GetTagNextOneParam(originalMessage string, tagName string) (string, error) {
	if tagName == "" {
		return "", errors.New("tagName cannot be empty")
	}

	wordArray := strings.Fields(originalArray)
	indexes := StringIndexOf(wordArray, tagName)

	if len(indexes) == 0 {
		return "", fmt.Errorf("tag '%s' not found", tagName)
	}

	firstIndex := indexes[0]
	if firstIndex < 0 || firstIndex >= len(wordArray)-1 {
		return "", fmt.Errorf("no parameter found after tag '%s'", tagName)
	}

	return wordArray[firstIndex+1], nil
}

func GetTagNextAllParams(originalMessage string, tagName string) []string {
	if tagName == "" {
		return []string{}
	}

	wordArray := strings.Fields(originalMessage)
	tagIndexes := StringIndexOf(wordArray, tagName)

	var params []string
	for _, v := range tagIndexes {
		if v >= 0 && v < len(wordArray)-1 {
			params = append(params, wordArray[v+1])
		}
	}
	return params
}
