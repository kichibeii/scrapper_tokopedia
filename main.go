package main

import (
	"encoding/csv"
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

	datas := scrawlData()

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

func scrawlData() []Data {
	// create data
	allDatas := make([]Data, 0)

	// init collector
	collector := colly.NewCollector(
		colly.AllowedDomains("www.tokopedia.com", "tokopedia.com"),
		colly.UserAgent("xy"),

		// for  debugging
		// colly.Debugger(&debug.LogDebugger{}),
	)

	// collector for product page
	infoCollector := collector.Clone()

	// get link of product
	collector.OnHTML("a[href]", func(element *colly.HTMLElement) {
		profileURL := element.Attr("href")

		//check correct link
		s := strings.Split(profileURL, ".")
		if s[0] == "https://ta" {
			url := getRealIP(profileURL)
			infoCollector.Visit(url)
		}
	})

	// get data that needed
	infoCollector.OnHTML("#main-pdp-container", func(element *colly.HTMLElement) {
		productName := element.ChildText("h1.css-1wtrxts")
		description := element.ChildText(".css-168ydy0 e1iszlzh1")
		Price := element.ChildAttr(".css-aqsd8m", "div")
		rating := element.ChildText("#lblPDPDetailProductRatingNumber")
		nameShop := element.ChildAttr(".css-1n8curp", "h2")

		var imageLink []string
		element.ForEach("div.css-1aplawl", func(_ int, kf *colly.HTMLElement) {
			linkImage := kf.ChildAttr("div.css-19i5z4j > img.success fade", "src")
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

	collector.OnRequest(func(request *colly.Request) {
		fmt.Println("Visiting ", request.URL.String())
	})

	infoCollector.OnRequest(func(request *colly.Request) {
		fmt.Println("infoCollector visiting ", request.URL.String())
	})

	collector.Visit("https://tokopedia.com/p/handphone-tablet/handphone?page=1")

	return allDatas
}

// function for get specified IP
func getRealIP(fakeIP string) string {
	s := strings.Split(fakeIP, "%2F")
	domain := s[2]
	store := s[3]
	productLink := s[4]
	productLinks := strings.Split(productLink, "%3F")

	return fmt.Sprintf("%s/%s/%s", domain, store, productLinks[0])
}
