package main

import (
	"fmt"
	"log"
	"os"
	"time"

	s "github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

type Product struct{
	Name string
	Link string
	Avg string
	Price string
	Reviews string
}

var products []Product

func main() {
	service, err := s.NewChromeDriverService("./chromedriver", 4444)
	if err != nil {
		log.Fatal("не удается запустить драйвер: ", err)
	}

	defer service.Stop()

	caps := s.Capabilities{"browserName": "chrome"}

	caps.AddChrome(chrome.Capabilities{
		Args: []string{
			"window-size=1920x1080",
			"--no-sandbox",
			"--disable-dev-shm-usage",
			"--disable-blink-features=AutomationControlled",
			"--disable-infobars",
		},
	})

	driver, err := s.NewRemote(caps, "")
	if err != nil {
		log.Fatal("не удается подключится к драйверу: ", err)
	}

	err = driver.Get("твой поисковой запрос")
	if err != nil {
		log.Fatal("не удалось открыть страницу")
	}

	time.Sleep(10 * time.Second)
	var lastHeight, nowHeight float64
	var count int
	for {
		_, err = driver.ExecuteScript("window.scrollBy(0, 1000);", nil)
		if err != nil {
			log.Fatal("Ошибка при скроллинге: ", err)
		}

		now, err := driver.ExecuteScript("return [document.documentElement.scrollHeight];", nil)
		if err != nil {
			log.Fatal("Ошибка при выполнении скрипта: ", err)
		}
		
		values := now.([]interface{})
		nowHeight = values[0].(float64)

		if nowHeight == lastHeight {
			count += 1
			if count == 5 {
				log.Println("Достигнут конец страницы")
			break
			}
		}else {
			count = 0
		}

		lastHeight = nowHeight

		time.Sleep(100 * time.Millisecond)
	}

	paginator, err := driver.FindElement(s.ByID, "paginator")
	if err != nil {
		log.Println("не удалось найти пагинатор: ", err)
		return
	}

	beforeProducts, err := paginator.FindElement(s.ByXPATH, "./div[1]")
	if err != nil {
		log.Println("не удалось найти див cо всеми продуктами: ", err)
		return
	}
	
	allProducts, _ := beforeProducts.FindElements(s.ByXPATH, "./div")
	
	for _, elem := range allProducts {
		beforeTenProducts, err := elem.FindElement(s.ByXPATH, "./div")
		if err != nil {
			log.Println("не удалось найти товары: ", err)
			continue
		}

		tenProducts, err := beforeTenProducts.FindElements(s.ByXPATH, "./div")
		if err != nil {
			log.Println("не удалось найти товары: ", err)
			continue
		}

		for _, product := range tenProducts {
			var productJSON Product
			beforeInfo, err := product.FindElement(s.ByXPATH, "./div[1]")
			if err != nil {
				log.Println("не удалось найти див перед инфой о товаре: ", err)
				continue
			}

			

			info, err := beforeInfo.FindElement(s.ByXPATH, "./div[1]")
			if err != nil {
				log.Println("нет инфы: ", err)
				continue
			}

			spanPrice := findPrice(info)

			price, _ := spanPrice.Text()
			productJSON.Price = price

			spanName, href := findName(info)
			name, _ := spanName.Text()

			productJSON.Link = href
			productJSON.Name = name
			
			spans := findAvgReviews(info)

			avg, _ := spans[0].Text()
			if avg[1] == '.' {
				productJSON.Avg = avg

				reviews, _ := spans[2].Text()
				productJSON.Reviews = reviews
			} else {
				productJSON.Avg = "нет авг"
				productJSON.Reviews = "нет отзывов"
			}			

			products = append(products, productJSON)
		}
	} 

	json, _ := os.Create("products.json")
	for _, product := range products {
		line := fmt.Sprintf("{\"name\": \"%s\", \"price\": \"%s\", \"avg\": \"%s\", \"reviews\": \"%s\", \"link\": \"%s\"}\n", product.Name, product.Price, product.Avg, product.Reviews, product.Link)
		_, err := json.WriteString(line)
		if err != nil {
			log.Println("Ошибка при записи в файл:", err)
			return
		}
	}

	log.Println("Данные успешно добавлены в файл prducts.json")
}

func findPrice(info s.WebElement) s.WebElement {
	wherePrice, err := info.FindElement(s.ByXPATH, "./div[1]")
	if err != nil {
		log.Println("не знаю где цена: ", err)
		return nil
	}
	
	spanPrice, err := wherePrice.FindElement(s.ByCSSSelector, "span")
	if err != nil {
		log.Println("нет спана с ценой: ", err)
		 return nil
	}
	return spanPrice
}

func findName(info s.WebElement) (s.WebElement, string){
	whereNameNLink, err := info.FindElement(s.ByCSSSelector, "a")
	if err != nil {
		log.Println("не знаю где имя и ссылка: ", err)
		 return nil, ""
	}

	href, _ := whereNameNLink.GetAttribute("href")

	spanName, err := whereNameNLink.FindElement(s.ByCSSSelector, "span")
	if err != nil {
		log.Println("нет спана с ценой: ", err)
		return nil, ""
	}

	return spanName, href
}

func findAvgReviews(info s.WebElement) []s.WebElement {
	whereAvgNReviews, err := info.FindElement(s.ByXPATH, "./div[position()=last()-1]")
	if err != nil {
		log.Println("не знаю где авг и отзывы: ", err)
		return nil
	}

	spans, err := whereAvgNReviews.FindElements(s.ByCSSSelector, "span")
	if err != nil {
		log.Println("нет спанов с авг и отзывами: ", err)
		 return nil
	}
	return spans
}

