// Package sliceutil provides utility functions for slice manipulation.
package sliceutil

import (
	"reflect"
)

// Contains checks if a slice contains a specific element.
func Contains[T comparable](slice []T, element T) bool {
	for _, item := range slice {
		if item == element {
			return true
		}
	}
	return false
}

// Remove removes the first occurrence of an element from a slice.
func Remove[T comparable](slice []T, element T) []T {
	for i, item := range slice {
		if item == element {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

// RemoveAt removes an element at a specific index.
func RemoveAt[T any](slice []T, index int) []T {
	if index < 0 || index >= len(slice) {
		return slice
	}
	return append(slice[:index], slice[index+1:]...)
}

// Unique returns a new slice with duplicate elements removed.
func Unique[T comparable](slice []T) []T {
	keys := make(map[T]bool)
	var result []T

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}
	return result
}

// Reverse returns a new slice with elements in reverse order.
func Reverse[T any](slice []T) []T {
	result := make([]T, len(slice))
	for i, item := range slice {
		result[len(slice)-1-i] = item
	}
	return result
}

// Filter returns a new slice containing only elements that satisfy the predicate.
func Filter[T any](slice []T, predicate func(T) bool) []T {
	var result []T
	for _, item := range slice {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

// Map applies a function to each element and returns a new slice with the results.
func Map[T, R any](slice []T, mapper func(T) R) []R {
	result := make([]R, len(slice))
	for i, item := range slice {
		result[i] = mapper(item)
	}
	return result
}

// Find returns the first element that satisfies the predicate and true, or zero value and false.
func Find[T any](slice []T, predicate func(T) bool) (T, bool) {
	var zero T
	for _, item := range slice {
		if predicate(item) {
			return item, true
		}
	}
	return zero, false
}

// IndexOf returns the index of the first occurrence of an element, or -1 if not found.
func IndexOf[T comparable](slice []T, element T) int {
	for i, item := range slice {
		if item == element {
			return i
		}
	}
	return -1
}

// Chunk splits a slice into chunks of specified size.
func Chunk[T any](slice []T, size int) [][]T {
	if size <= 0 {
		return nil
	}

	var chunks [][]T
	for i := 0; i < len(slice); i += size {
		end := i + size
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}

// IsEmpty checks if a slice is empty.
func IsEmpty[T any](slice []T) bool {
	return len(slice) == 0
}

// Equal checks if two slices are equal (same elements in same order).
func Equal[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// DeepEqual checks if two slices are deeply equal using reflection.
func DeepEqual(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}
