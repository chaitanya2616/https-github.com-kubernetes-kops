/*
Copyright 2019 The Kubernetes Authors.

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

package openstackmodel

import "k8s.io/kops/upup/pkg/fi"

// s is a helper that builds a *string from a string value
func s(v string) *string {
	return fi.PtrTo(v)
}

// i32 is a helper that builds a *int32 from an int32 value
func i32(v int32) *int32 {
	return fi.PtrTo(v)
}

// i is a helper that builds a *int from an int value
func i(v int) *int {
	return fi.PtrTo(v)
}
