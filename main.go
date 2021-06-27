package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gocolly/colly"
)

// Info is a struct for my website data
type Data struct {
	ID          int      `json:"id"`
	ProductName string   `json:"product_name"`
	Description string   `json:"description"`
	ImageLink   []string `json:"image_link"`
	Price       string   `json:"price"`
	Rating      string   `json:"rating"`
	NameOfStore string   `json:"name_of_store"`
	Link        string   `json:"link"`
}

func main() {
	datas := crawlData()

	writeDataToCSV(datas)
}

func writeDataToCSV(datas []Data) {
	csvFile, err := os.Create("result.csv")
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	csvwriter := csv.NewWriter(csvFile)

	_ = csvwriter.Write([]string{"No", "Product Name", "Description", "Image Link", "Price", "Rating", "Name Of Store"})

	for i, data := range datas {
		imageLink := strings.Join(data.ImageLink, ",")
		_ = csvwriter.Write([]string{fmt.Sprintf("%d", i+1), data.ProductName, data.Description, imageLink, data.Price, data.Rating, data.NameOfStore})
	}

	csvwriter.Flush()
	csvFile.Close()
}

func crawlData() []Data {
	// make new variable's array of data
	allDatas := make([]Data, 0)

	// new collector Colly
	c := colly.NewCollector(
		colly.AllowedDomains("www.tokopedia.com", "tokopedia.com"),
		colly.UserAgent("xy"),
	)

	// make clone for get product page
	infoCollector := c.Clone()

	// get link product page
	c.OnHTML(".e1nlzfl3", func(e *colly.HTMLElement) {
		profileUrl := e.ChildAttr("a", "href")
		profileUrl = e.Request.AbsoluteURL(profileUrl)
		infoCollector.Visit(profileUrl)
	})

	// get data
	infoCollector.OnHTML("#main-pdp-container", func(e *colly.HTMLElement) {
		productName := e.ChildText("h1.css-1wtrxts")
		description := e.ChildText("span.css-168ydy0")
		Price := e.ChildText("div.price")
		rating := e.ChildText("h5.css-zeq6c8 > span")
		nameShop := e.ChildAttr("a.css-1n8curp", "href")

		var imageLink []string
		e.ForEach("div.css-1aplawl", func(_ int, kf *colly.HTMLElement) {
			linkImage := kf.ChildAttr("div.css-19i5z4j > img", "src")
			imageLink = append(imageLink, linkImage)
		})

		allDatas = append(allDatas, Data{
			ProductName: productName,
			Description: description,
			Price:       Price,
			Rating:      rating,
			ImageLink:   imageLink,
			NameOfStore: nameShop,
		})
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting: ", r.URL.String())
	})

	infoCollector.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting Profile URL: ", r.URL.String())
	})

	c.Visit("https://tokopedia.com/p/handphone-tablet/handphone")

	js, err := json.MarshalIndent(allDatas, "", "    ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(js))

	return allDatas
}
