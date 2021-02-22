package utils

import (
	"math/rand"
	"testing"
	"time"
)

var (
	originalMessage = ""
	tagName         = "/tag"
	tagParam        = "object"
)

func TestStringToInt32(t *testing.T) {

	t.Run("NormalNumber", func(t *testing.T) {
		normalNumber := "1"
		if resultInt, err := StringToInt32(normalNumber); resultInt != 1 && err == nil {
			t.Errorf("The String %s expected be int32 %s, but %d got", normalNumber, normalNumber, resultInt)
		}
	})

	t.Run("UnexpectedString", func(t *testing.T) {
		if _, err := StringToInt32("not a number"); err == nil {
			t.Errorf("The String must be the number string, otherwise the error wouldn't be nil")
		}
	})
}

func TestStringIndexOf(t *testing.T) {
	// the string word must exists in the array.
	indexes := StringIndexOf([]string{"x", tagName, "y"}, tagName)
	if indexes[0] != 1 {
		t.Errorf("The index of the word [%s] expected to be 1, but %d got", tagName, indexes[0])
	}
	indexes = StringIndexOf([]string{"x", tagName, "y", tagName}, tagName)
	if indexes[0] != 1 || indexes[1] != 3 {
		t.Errorf("The indexes of the word [%s] expected to be 1, 3, but %d, %d got", tagName, indexes[0], indexes[1])
	}
}

func setupForOneParam(t *testing.T) {
	// use `go clean -testcache` to clear the previous cache
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 3; i++ {
		randomNumber := rand.Intn(5) + 1
		blank := ""
		for j := 0; j < randomNumber; j++ {
			blank += " "
		}
		if i == 1 {
			originalMessage += tagName
		}
		if i == 2 {
			originalMessage += tagParam
		}
		originalMessage += blank
	}
	t.Log(originalMessage)
}

func TestGetTagNextOneParam(t *testing.T) {
	setupForOneParam(t)
	param, err := GetTagNextOneParam(originalMessage, tagName)
	if param != tagParam {
		t.Errorf("The param after the tag expected to be %s, but %s got", tagParam, param)
	}
	if err != nil {
		t.Errorf(err.Error())
	}
}
