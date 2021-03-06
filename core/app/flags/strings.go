// Copyright (C) 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package flags

import (
	"net/url"
	"strings"
)

// Strings implements flag.Value for lists of strings.
type Strings []string

func (f *Strings) String() string    { return strings.Join(f.Strings(), ":") }
func (f *Strings) Strings() []string { return ([]string)(*f) }

func (f *Strings) Set(value string) error {
	*f = append(*f, value)
	return nil
}

// URL implements flag.Value for a url.URL flag.
type URL struct {
	url.URL
}

func (f *URL) Set(value string) error {
	u, err := url.Parse(value)
	if err != nil {
		return err
	}
	f.URL = *u
	return nil
}
