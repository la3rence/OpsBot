package utils

import (
	"fmt"
	"strings"
)

func StringIndexOf(originalArray []string, wordToFind interface{}) int {
	interfaceArray := make([]interface{}, len(originalArray))
	for i, v := range originalArray {
		interfaceArray[i] = v
	}
	n := len(originalArray)
	var i = 0
	for ; i < n; i++ {
		if strings.Compare(wordToFind.(string), originalArray[i]) == 0 {
			return i
		}
	}
	return -1
}

func main() {
	originalArray := []string{"apple", "banana", "lime", "橘子", "orange", "橙子", "pineapple", "vine"}
	// convert []string to []interface
	//interfaceArray := make([]interface{}, len(originalArray))
	//for i, v := range originalArray {
	//	interfaceArray[i] = v
	//}

	// Find index of  "orange" in array
	fmt.Printf("orange index=%d\n", StringIndexOf(originalArray, "橙子"))
}
