package main

import (
	"unicode"
)

func AllLower(word string) bool {
	for _, c := range word {
		if unicode.IsUpper(c) {
			return false
		}
	}
	return true
}
