// Copyright 2019 TiKV Project Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package slice

// AnyOf returns true if any element in the slice matches the predict func.
func AnyOf[T any](s []T, p func(int) bool) bool {
	for i := range s {
		if p(i) {
			return true
		}
	}
	return false
}

// NoneOf returns true if no element in the slice matches the predict func.
func NoneOf[T any](s []T, p func(int) bool) bool {
	return !AnyOf(s, p)
}

// AllOf returns true if all elements in the slice match the predict func.
func AllOf[T any](s []T, p func(int) bool) bool {
	np := func(i int) bool {
		return !p(i)
	}
	return NoneOf(s, np)
}

// Contains returns true if the given slice contains the value.
func Contains[T comparable](slice []T, value T) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// Remove removes the value from the slice.
func Remove[T comparable](slice []T, value T) []T {
	i, j := 0, 0
	for ; i < len(slice); i++ {
		if slice[i] != value {
			slice[j] = slice[i]
			j++
		}
	}
	return slice[:j]
}

// EqualWithoutOrder checks if two slices are equal without considering the order.
func EqualWithoutOrder[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for _, item := range a {
		if !Contains(b, item) {
			return false
		}
	}
	return true
}
