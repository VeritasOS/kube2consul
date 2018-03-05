/*
Copyright 2018 Veritas Technologies LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

// Returns true if haystack containts the needle
func Contains(haystack []string, needle string) bool {

	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}

// Returns true if lhs is a subset of rhs
func Subset(lhs, rhs []string) bool {
	for _, element := range lhs {

		if !Contains(rhs, element) {
			return false
		}
	}
	return true
}
