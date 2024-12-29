package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/gocolly/colly"
	"github.com/joho/godotenv"
)

func main() {
	_, _ = fmt.Print()
	_ = colly.NewCollector()
	godotenv.Load(".env")
	password := os.Getenv("PASSWORD")
	email := os.Getenv("EMAIL")

	//figure out whether we are getting cookies or scraping

	if os.Args[1] == "cookies" {
		cookies, _ := json.MarshalIndent((getcookies(email, password)), "", "  ")
		f, err := os.OpenFile("cookies.json", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			log.Fatal(err)
		}
		if _, err := f.Write(cookies); err != nil {
			f.Close() // ignore error; Write error takes precedence
			log.Fatal(err)
		}
		if err := f.Close(); err != nil {
			log.Fatal(err)
		}

	}
	if os.Args[1] == "scrape" {
		rodToCollyCookies()
	}

}

func getcookies(email string, password string) []*proto.NetworkCookie {
	browser := rod.New().MustConnect().NoDefaultDevice()

	page := browser.MustConnect().MustPage("https://app.blackbaud.com/signin/?redirecturl=https://gcschool.myschoolapp.com/lms-assignment/assignment-center/student")
	page.MustElementX("/html/body/app-root/skyux-app-shell/div/app-sign-in-and-up-route-index/app-sign-in-and-up/div/div/app-centered-base-template-component/div/div[1]/div/button[2]").MustClick()
	page.MustElementX("//*[@id=\"identifierId\"]").MustInput(email)

	page.MustElementX("//*[@id=\"identifierNext\"]/div/button").MustClick()
	page.MustElementX("//*[@id=\"password\"]/div[1]/div/div[1]/input").MustInput(password)
	page.MustElementX("//*[@id=\"passwordNext\"]/div/button").MustClick()
	page.MustElementX("//*[@id=\"sky-split-view-drawer-1\"]/div[3]").WaitVisible()
	wait := page.MustWaitNavigation()
	wait()
	cookies := browser.MustGetCookies()

	return cookies
}

/*
ok so like basically rod will be used to get past
oauth and get all the users cookies, which is gonna be
wildly inefficent but will work anyways

after that, colly will be used to constantly and efficiently
read assignments from gracenet USING the cookies
rod has gotten

*/

func rodToCollyCookies() []*http.Cookie {
	file, err := os.Open("cookies.json")
	if err != nil {
		log.Fatalf("Failed to open cookie file: %v", err)
	}
	defer file.Close()

	var collyCookies []*http.Cookie
	dec := json.NewDecoder(file)

	for {
		var c any
		if err := dec.Decode(&c); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		fmt.Print(c)
		CTime := (c.(map[string]interface{})["path"].(float64))
		expirationTime := time.Unix(int64(CTime), int64((CTime-float64(int64(CTime)))*1e9))

		collyCookies = append(collyCookies, &http.Cookie{
			Name:    c.(map[string]interface{})["name"].(string),
			Value:   c.(map[string]interface{})["value"].(string),
			Domain:  c.(map[string]interface{})["domain"].(string),
			Path:    c.(map[string]interface{})["path"].(string),
			Expires: expirationTime,
			/*um ok. problem. http.cookie doesn't have all the necessary
			fields that the cookies i got from go have. so we're going to
			have to give gcschool cookies with missing fields??
			the problem with gocolly is that its collector ONLY
			accepts http.Cookie objects. so like i have no idea if this
			will work. missing fields: size,
			*/
			//NEVERMIND  GRACENET DOESNT EVEN CHECK 👅

		})
	}
	return collyCookies
}
