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

func HasIntOnSorted(a []int, x int) bool {
	return SearchIntsEqual(a, x) != len(a)
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

func HasStringOnSorted(a []string, x string) bool {
	return SearchStringsEqual(a, x) != len(a)
}
