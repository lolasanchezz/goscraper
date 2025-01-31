package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/stealth"
	"github.com/gocolly/colly"
	"github.com/joho/godotenv"
)

func main() {
	_, _ = fmt.Print()
	_ = colly.NewCollector()
	godotenv.Load(".env")
	password := os.Getenv("PASSWORD")
	email := os.Getenv("EMAIL")
	link := os.Getenv("LINK")

	//figure out whether we are getting cookies or scraping

	if os.Args[1] == "cookies" {
		cookies, _ := json.MarshalIndent((getcookies(email, password, link)), "", "  ")
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
		fmt.Print("scraping")

		getXHRreq(email, password, os.Getenv("LINK"))
	}

}

func getcookies(email string, password string, link string) []*proto.NetworkCookie {
	browser := rod.New().MustConnect().NoDefaultDevice()
	defer browser.MustClose()
	page := stealth.MustPage(browser)
	page.MustNavigate(link)
	page.MustElementX("/html/body/app-root/skyux-app-shell/div/app-sign-in-and-up-route-index/app-sign-in-and-up/div/div/app-centered-base-template-component/div/div[1]/div/button[2]").MustClick()
	page.MustElementX("//*[@id=\"identifierId\"]").MustInput(email)
	fmt.Print("at email")

	page.MustElementR("button", "Next").MustWaitVisible().MustClick()
	page.MustScreenshot("a.png")
	fmt.Print("clicked")

	page.MustElementR("div", "Enter your password").MustParent().MustElement(":first-child").MustInput(password)
	fmt.Print("at password")
	page.MustElementR("button", "Next").MustWaitVisible().MustClick()
	page.MustElementX("//*[@id=\"sky-split-view-drawer-1\"]/div[3]").WaitVisible()
	fmt.Print("made it to blackbaud")
	page.MustScreenshot("a.png")

	page.MustScreenshot("b.png")

	fmt.Print("wait function")
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

func rodToCookies() ([]*http.Cookie, []*proto.NetworkCookie) {
	file, err := os.Open("cookies.json")
	if err != nil {
		log.Fatalf("Failed to open cookie file: %v", err)
	}
	defer file.Close()

	type cookietype struct {
		Name         string  `json:"name"`
		Value        string  `json:"value"`
		Domain       string  `json:"domain"`
		Path         string  `json:"path"`
		Expires      float64 `json:"expires"`
		Size         int     `json:"size"`
		HttpOnly     bool    `json:"httpOnly"`
		Secure       bool    `json:"secure"`
		Session      bool    `json:"session"`
		Priority     string  `json:"priority"`
		SameParty    bool    `json:"sameParty"`
		SourceScheme string  `json:"sourceScheme"`
		SourcePort   int     `json:"sourcePort"`
	}
	var collyCookies []*http.Cookie
	var rodCookies []*proto.NetworkCookie
	dec := json.NewDecoder(file)
	//opening bracket
	if _, err := dec.Token(); err != nil {
		log.Fatal(err)
	}

	for dec.More() {
		//decoding part
		var c cookietype
		var nc proto.NetworkCookie
		if err := dec.Decode(&c); err != nil {
			log.Fatal(err)
		}
		if err = dec.Decode(&nc); err != nil {
			log.Fatal("decoding protonetwork error: ", err)
		} else {
			rodCookies = append(rodCookies, &nc)
		}

		//CTime := c.Expires
		//expirationTime := time.Unix(int64(CTime), int64((CTime-float64(int64(CTime)))*1e9))
		//fmt.Print(c)
		collyCookies = append(collyCookies, &http.Cookie{
			Name:   c.Name,
			Value:  c.Value,
			Domain: c.Domain,
			//hopefully i dont need these.
			/*

				Path:    c.Path,
				Expires: expirationTime,
			*/

			/*um ok. problem. http.cookie doesn't have all the necessary
			fields that the cookies i got from go have. so we're going to
			have to give gcschool cookies with missing fields??
			the problem with gocolly is that its collector ONLY
			accepts http.Cookie objects. so like i have no idea if this
			will work. missing fields: size,
			*/
			//NEVERMIND  GRACENET DOESNT EVEN CHECK 👅
			//oh. also i have to change to a for loop because
			// apparently decode treats the entire array like one json decoded thing.
		})
		//now making proto network cookies

	}

	//fmt.Print(collyCookies[0])
	return collyCookies, rodCookies
}

// using colly!!!
func getassignments(cookies []*http.Cookie, link string) {
	co := colly.NewCollector()
	fmt.Print(*(cookies[0]))
	if err := co.SetCookies(link, cookies); err != nil {
		log.Fatalf("Couldn't set cookies: %v", err)
	}

	co.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting:", r.URL.String())
	})

	co.OnError(func(r *colly.Response, err error) {
		log.Printf("Request URL: %s failed with status: %d, error: %v", r.Request.URL, r.StatusCode, err)
	})

	co.OnHTML("div", func(e *colly.HTMLElement) {
		fmt.Println("found", e.Text)
		fmt.Println("eh")
	})
	co.OnResponse(func(r *colly.Response) {
		fmt.Println("recieved", r.StatusCode)
	})

	if err := co.Visit(link); err != nil {
		log.Fatalf("Couldn't visit %s: %v", link, err)
	}
}

func getXHRreq(email string, password string, link string) {
	browser := rod.New().MustConnect().NoDefaultDevice()

	defer browser.MustClose()
	page := stealth.MustPage(browser)
	page.MustNavigate(link)

	router := browser.HijackRequests()
	router.MustAdd("*assignments", func(ctx *rod.Hijack) {
		fmt.Print("h4ello")
		headers := ctx.Request.Headers()
		payload := ctx.Request.Body()

		f, _ := os.OpenFile("headers.txt", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)

		f.Write([]byte(headersTocUrl(headers, payload)))
		defer f.Close()
	})
	page.MustNavigate(link)
	page.MustElementX("/html/body/app-root/skyux-app-shell/div/app-sign-in-and-up-route-index/app-sign-in-and-up/div/div/app-centered-base-template-component/div/div[1]/div/button[2]").MustClick()
	page.MustElementX("//*[@id=\"identifierId\"]").MustInput(email)
	fmt.Print("at email")

	page.MustElementR("button", "Next").MustWaitVisible().MustClick()
	page.MustScreenshot("a.png")
	fmt.Print(password)
	element := page.MustElementR("div", "Enter your password")
	if element == nil {
		fmt.Println("Password prompt element not found")
	} else {
		parent := element.MustParent()
		if parent == nil {
			fmt.Println("Parent element not found")
		} else {
			input := parent.MustElement(":first-child")
			if input == nil {
				fmt.Println("Input element not found")
			} else {
				input.MustInput(password)
				fmt.Print("at password")
			}
		}
	}
	fmt.Print("at password")
	page.MustElementR("button", "Next").MustWaitVisible().MustClick()
	fmt.Print("navigating")
	page.MustElementX("//*[@id=\"sky-split-view-drawer-1\"]/div[3]").WaitVisible()
	fmt.Print("made it")

	page.WaitLoad()
	go router.Run()
	fmt.Print("router running")
	page.MustNavigate("https://gcschool.myschoolapp.com/app/student#calendar")

	fmt.Print("reloading")

	//now timne to unmarshal the json and put it into a thihng that cna be curled!!!!!!!

}

func headersTocUrl(headers proto.NetworkHeaders, body string) string {
	finalStr := "'" + os.Getenv("APILINK1") + `' \`
	for key, value := range headers {
		finalStr = finalStr + "-H" + "'" + key + ": " + value.Str() + `' \`
	}
	finalStr = finalStr + `--data-raw '{"bulkURL":"mycalendar/assignments","startDate":"01/25/2025","endDate":"01/26/2025","filterString":"7379045_15500_26935166_4,7379045_15500_26934337_4,7379045_15500_26935165_4,7379045_15500_26935101_4,7379045_15500_26935206_4,7379045_15500_26934223_4,7379045_15500_26934151_4,7379045_15500_27434581_4,7379045_15500_26934351_4,7379045_15500_26934368_4,7379045_15500_26934366_4,7379045_15500_26934171_4,7379045_15500_26934141_4,7379045_15500_26934117_4,7379045_15500_26934304_4,7379045_15500_26934259_4,7379045_15500_27360846_4,7379045_15500_26935218_4","recentSave":false}'`
	return finalStr
}
