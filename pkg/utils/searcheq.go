package utils

import "sort"

func SearchIntsEqual(a []int, x int) int {
	searchedIndex := sort.SearchInts(a, x)
	if searchedIndex == len(a) {
		return searchedIndex
	}
	if a[searchedIndex] == x {
		return searchedIndex
	} else {
		return len(a)
	}
}

func SearchStringsEqual(a []string, x string) int {
	searchedIndex := sort.SearchStrings(a, x)
	if searchedIndex == len(a) {
		return searchedIndex
	}
	if a[searchedIndex] == x {
		return searchedIndex
	} else {
		return len(a)
	}
}
