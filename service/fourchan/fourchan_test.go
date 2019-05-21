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

package fourchan

import (
	"flag"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mlvzk/piko/service"
	"github.com/mlvzk/piko/service/testutil"
)

const base = "https://boards.4channel.org"

var update = flag.Bool("update", false, "update .golden files")

func TestFourchanIsValidTarget(t *testing.T) {
	tests := map[string]bool{
		"https://boards.4channel.org/g/": true,
		"boards.4channel.org/g/":         true,
		"https://boards.4channel.org/g/thread/70377765/hpg-esg-headphone-general": true,
		"https://boards.4chan.org/pol/":                                           true,
		"https://imgur.com/":                                                      false,
	}

	for target, expected := range tests {
		if (Fourchan{}).IsValidTarget(target) != expected {
			t.Errorf("Invalid result, target: %v, expected: %v", target, expected)
		}
	}
}

func TestIteratorNext(t *testing.T) {
	ts := testutil.CacheHttpRequest(t, base, *update)
	defer ts.Close()

	iterator := FourchanIterator{
		url: ts.URL + "/adv/thread/20765545/i-want-to-be-the-very-best-like-no-one-ever-was",
	}

	items, err := iterator.Next()
	if err != nil {
		t.Fatalf("iterator.Next() error: %v", err)
	}

	expected := []service.Item{
		{
			Meta: map[string]string{
				"title":        "1306532808724.png",
				"imgURL":       "https://i.4cdn.org/adv/1554743536847.png",
				"id":           "1554743536847",
				"ext":          "png",
				"thumbnailURL": "https://i.4cdn.org/adv/1554743536847s.jpg",
			},
			DefaultName: "%[title]",
			AvailableOptions: map[string][]string{
				"thumbnail": {"yes", "no"},
			},
			DefaultOptions: map[string]string{
				"thumbnail": "no",
			},
		},
		{
			Meta: map[string]string{
				"title":        "1306532948201.png",
				"imgURL":       "https://i.4cdn.org/adv/1554745883162.png",
				"id":           "1554745883162",
				"ext":          "png",
				"thumbnailURL": "https://i.4cdn.org/adv/1554745883162s.jpg",
			},
			DefaultName: "%[title]",
			AvailableOptions: map[string][]string{
				"thumbnail": {"yes", "no"},
			},
			DefaultOptions: map[string]string{
				"thumbnail": "no",
			},
		},
	}

	if diff := pretty.Compare(items, expected); diff != "" {
		t.Errorf("%s diff:\n%s", t.Name(), diff)
	}
}
