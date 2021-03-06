// Copyright 2019 mlvzk
// This file is part of the piko library.
//
// The piko library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The piko library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the piko library. If not, see <http://www.gnu.org/licenses/>.

package twitter

import (
	"flag"
	"strings"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mlvzk/piko/service"
	"github.com/mlvzk/piko/service/testutil"
)

const base = "https://twitter.com"

var update = flag.Bool("update", false, "update .golden files")

func TestIsValidTarget(t *testing.T) {
	tests := map[string]bool{
		"https://twitter.com/golang":                            true,
		"https://twitter.com/golang/status/1116531752602951681": true,
		"twitter.com/golang/status/1116531752602951681":         true,
		"https://soundcloud.com/":                               false,
	}

	for target, expected := range tests {
		if (Twitter{}).IsValidTarget(target) != expected {
			t.Errorf("Invalid result, target: %v, expected: %v", target, expected)
		}
	}
}

func TestIteratorNext(t *testing.T) {
	ts := testutil.CacheHttpRequest(t, base, *update)
	defer ts.Close()

	iterator := TwitterIterator{
		url: ts.URL + "/golang/status/1106303553474301955",
	}

	items, err := iterator.Next()
	if err != nil {
		t.Fatalf("iterator.Next() error: %v", err)
	}

	if len(items) < 1 {
		t.Fatalf("Items array is empty")
	}

	if !strings.Contains(items[0].Meta["downloadURL"], "pbs.twimg.com") {
		t.Fatalf("Incorrect downloadURL")
	}
	items[0].Meta["downloadURL"] = "ignore"

	expected := []service.Item{
		{
			Meta: map[string]string{
				"downloadURL": "ignore",
				"index":       "0",
				"id":          "1106303553474301955",
				"author":      "golang",
				"description": "🎉 Go 1.12.1 and 1.11.6 are released!\n\n🗣 Announcement: https://t.co/PAttJybffj\n\nHappy Pi day! 🥧\n\n#golang",
				"type":        "image",
				"ext":         "jpg",
			},
			DefaultName: "%[author]-%[id]-%[index].%[ext]",
		},
	}

	if diff := pretty.Compare(items, expected); diff != "" {
		t.Errorf("%s diff:\n%s", t.Name(), diff)
	}
}
