package main

import (
	"bufio"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"os"
	"strings"
)

type koreanEnglishPair struct {
	korean string
	english string
}

func generateUrls() []*string {
	fmt.Printf("Indexing list of URLs...\n")
	urls := make([]*string, 0, 0)
	resp, err := http.DefaultClient.Get("https://korean.arts.ubc.ca/online-textbook-korn-102/")
	if err != nil {
		panic(err)
	}

	document, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		panic(err)
	}

	document.Find(".accordion .accordion-group").Each(func(i int, selection *goquery.Selection) {
		selection.Find(".accordion-inner a").Each(func(i int, subselection *goquery.Selection) {
			preText := subselection.Text()
			cutText := subselection.Text()
			cutText = strings.ToLower(cutText)
			cutText = strings.ReplaceAll(cutText, "vocabulary", "")
			cutText = strings.ReplaceAll(cutText, "korean script", "")
			if len(preText) == len(cutText) {
				return
			}

			hrefUrl, _ := subselection.Attr("href")
			urls = append(urls, &hrefUrl);
		})
	})
	return urls

}

func crawlAndSave(url string) {
	fmt.Printf("Crawling '%v'...\n", url)
	resp, err := http.DefaultClient.Get(url);
	if err != nil {
		panic(err)
	}

	document, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		panic(err)
	}

	var title string
	title = document.Find("div.breadcrumb.expand").Find("a").Eq(2).Text()
	title = strings.ReplaceAll(title, ":", " -")
	title = strings.ReplaceAll(title, "/", "")

	tableOffset := 0
	if len(title) != len(strings.ReplaceAll(title, "Lesson 1 ", "")) || len(title) != len(strings.ReplaceAll(title, "Lesson 2 ", "")) {
		tableOffset = 1
	}

	f, err := os.Create(title+".csv")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if _, err := f.WriteString("\"Korean\";\"English\"\n"); err != nil {
		panic(err)
	}

	kep := koreanEnglishPair{}

	document.Find(".entry-content .row-table tr").Each(func(i int, selection *goquery.Selection) {
		selection.Find("td").Each(func(i int, subselection *goquery.Selection) {
			switch i {
			case tableOffset+0:
				kep.korean = subselection.Text()
			case tableOffset+1:
				kep.english = subselection.Text()
			}
		})
		if len(kep.korean)+len(kep.english) == 0 {
			return
		}
		if _, err := f.WriteString(fmt.Sprintf("\"%v\";\"%v\"\n", kep.korean, kep.english)); err != nil {
			panic(err)
		}
	})

	if err := f.Sync(); err != nil {
		panic(err)
	}
	fmt.Printf("'%v.csv' generated.\n", title)
}

func getUrl() string {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("Please insert URL of the page: ")
	scanner.Scan()
	if len(scanner.Text()) <= 1 {
		fmt.Printf("No valid URL entered, using the URL for first lesson...\n")
		return "https://korean.arts.ubc.ca/online-textbook-korn-102/lesson-1-basic-expressions/basic-expressions-i-korean-script/"
	}
	return scanner.Text()
}

func main() {
	urls := generateUrls()
	for _, url := range urls {
		crawlAndSave(*url)
	}
	fmt.Printf("Done!\n")
}
