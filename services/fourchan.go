package shovel

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Fourchan struct{}
type FourchanIterator struct {
	url  string
	page int
	end  bool
}

func (s Fourchan) IsValidTarget(target string) bool {
	return strings.Contains(target, "4chan.org/") || strings.Contains(target, "4channel.org/")
}

func (s Fourchan) FetchItems(target string) ServiceIterator {
	return &FourchanIterator{
		url:  target,
		page: 1,
		end:  false,
	}
}

func (s Fourchan) Download(meta, options map[string]string) (io.Reader, error) {
	url := meta["imgURL"]
	if options["thumbnail"] == "yes" {
		url = meta["thumbnailURL"]
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func (i *FourchanIterator) Next() ([]Item, error) {
	i.end = true

	resp, err := http.Get(i.url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GET %v returned a wrong status code - %v", i.url, resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	items := []Item{}
	doc.Find("div.file").Each(func(_ int, sel *goquery.Selection) {
		imgURL, iuExists := sel.Find("a.fileThumb").Attr("href")
		if !iuExists {
			return
		}
		if imgURL[0] == '/' {
			imgURL = "https:" + imgURL
		}

		thumbnailURL, _ := sel.Find("a.fileThumb img").Attr("src")
		if thumbnailURL[0] == '/' {
			thumbnailURL = "https:" + thumbnailURL
		}

		titleSel := sel.Find("div.fileText a")
		title, titleExists := titleSel.Attr("title")
		if !titleExists || strings.TrimSpace(title) == "" {
			title = titleSel.Text()
		}

		dotParts := strings.Split(imgURL, ".")
		ext := dotParts[len(dotParts)-1]

		slashParts := strings.Split(imgURL, "/")
		lastSlashDotParts := strings.Split(slashParts[len(slashParts)-1], ".")
		id := lastSlashDotParts[0]

		items = append(items, Item{
			Meta: map[string]string{
				"title":        title,
				"imgURL":       imgURL,
				"id":           id,
				"ext":          ext,
				"thumbnailURL": thumbnailURL,
			},
			DefaultName: "%[title]",
			AvailableOptions: map[string][]string{
				"thumbnail": []string{"yes", "no"},
			},
			DefaultOptions: map[string]string{
				"thumbnail": "no",
			},
		})
	})

	return items, nil
}

func (i FourchanIterator) HasEnded() bool {
	return i.end
}
