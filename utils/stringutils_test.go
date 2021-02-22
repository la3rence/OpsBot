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

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	if (a == nil) != (b == nil) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func TestGetTagNextAllParams(t *testing.T) {
	// use table-driven tests
	cases := []struct {
		Name          string
		Message       string
		ExpectedLabel []string
	}{
		{"single tag with param", "/tag a", []string{"a"}},
		{"two same tags with different params", "/tag a /tag b", []string{"a", "b"}},
		{"two same tags but one lack of param", "/tag a /tag", []string{"a"}},
		{"two different tags", "/tag a /untag b", []string{"a"}},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			params := GetTagNextAllParams(c.Message, tagName)
			if !stringSliceEqual(params, c.ExpectedLabel) {
				t.Errorf("The param after the tag expected to be %s, but %s got", c.ExpectedLabel, params)
			}
		})
	}
}
