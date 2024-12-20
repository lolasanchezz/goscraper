package main

import (
	"os"

	"github.com/go-rod/rod"
)

func main() {
	password := os.Getenv("PASSWORD")
	email := os.Getenv("EMAIL")
	//EVENTUALLY - get the link here as an environment variable, so the signin flow can be faster
	getcookies(email, password)
}

func getcookies(email string, password string) {
	page := rod.New().MustConnect().MustPage("https://app.blackbaud.com/signin/?redirecturl=https://gcschool.myschoolapp.com/lms-assignment/assignment-center/student")
	page.MustElementR("div", "Continue with Google").MustClick()

}
