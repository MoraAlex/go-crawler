package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/gocolly/colly/v2"
)

type Product struct {
	Id        string `json:"id"`
	Title     string `json:"title"`
	Url       string `json:"url"`
	Condition string `json:"condition"`
}

const (
	baseUrl = "https://www.ebay.com"
)

func main() {
	// change param to you want to filter (new, pre-owned, open box, brand new)
	// if you want to get all the products just pass an empty string ""
	param := "de segunda mano"
	crawl(param)
}

// function to start crawling the ebay site.
// pass the condition that you want to get (new, pre-owned, open box, brand new)
func crawl(condition string) error {
	c := colly.NewCollector()
	if err := createFolder("data"); err != nil {
		return nil
	}
	condition, err := validateParam(condition)
	if err != nil {
		return err
	}
	err = scrap(c, condition)
	if err != nil {
		return err
	}
	return nil
}

func createFolder(path string) error {
	err := os.Mkdir(path, 0755)
	// if error is that the foldear already exists we return nil.
	// the program can still working correctly
	if os.IsExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	return nil
}

func scrap(c *colly.Collector, condition string) error {
	var wg sync.WaitGroup
	// the list that have all the products have this class srp-river-results
	c.OnHTML(".srp-river-results", func(e *colly.HTMLElement) {
		// once in the list, for each element that have s-item__info get the requested info.
		e.ForEach(".s-item__info", func(i int, h *colly.HTMLElement) {
			temp := Product{}
			temp.Title = h.ChildText("span[role='heading']")
			temp.Url = h.ChildAttr("a[class='s-item__link']", "href")
			if after, ok := strings.CutPrefix(temp.Url, (baseUrl + "/itm/")); ok {
				temp.Id, _, _ = strings.Cut(after, "?")
			}
			conditionHtml := h.ChildText("span[class='SECONDARY_INFO']")
			temp.Condition = conditionHtml
			// only save the products that comply with the condition
			if strings.ToLower(conditionHtml) == strings.ToLower(condition) || condition == "" {
				wg.Add(1)
				go saveFile(temp, &wg)
			}
		})
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.Visit(baseUrl + "/sch/garlandcomputer/m.html")
	defer wg.Wait()
	return nil
}

func saveFile(product Product, wg *sync.WaitGroup) error {
	productJSON, err := json.MarshalIndent(product, "", "    ")
	defer wg.Done()
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return err
	}
	filePath := fmt.Sprintf("data/%v.json", product.Id)
	err = os.WriteFile(filePath, productJSON, 0644)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return err
	}
	return nil
}

// function to validate params.
func validateParam(param string) (string, error) {
	switch strings.ToLower(param) {
	case "nuevo", "totalmente nuevo":
		return "totalmente nuevo", nil
	case "new", "brand new":
		return "brand new", nil
	case "open box", "caja abierta":
		return param, nil
	case "de segunda mano", "usado":
		return "de segunda mano", nil
	case "pre-owned", "used":
		return "pre-owned", nil
	case "":
		return param, nil
	default:
		return "", errors.New("Parameter is not valid")
	}
}
