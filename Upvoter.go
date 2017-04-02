package main

import (
	"net/http"
	"net/url"
	"strings"
	"time"
	"io/ioutil"
	"encoding/json"
	"fmt"
	"strconv"
)

var AccessToken string = ""


// This is my first ever useful go application.
// It upvotes imgur posts that have the tag "TeamHuman"
// in order to aide humans in the april fools contest.
func main() {
	bytes, err := ioutil.ReadFile("AccessToken.txt")
	if err != nil {
		panic("oops")
	}

	AccessToken = string(bytes)
	for ;; {
		// upvote all available pages
		for i := 1; i < 10; i++ {
			votePage(i, false)
		}
		println("Entering sleep for a quick five minutes")
		time.Sleep(5 * time.Minute)
	}
}

func votePage(num int, retry bool) {
	println("Page: " + strconv.Itoa(num))

	resp := request("GET", "gallery/t/teamhuman/time/" + string(num), nil)

	if resp == nil {
		println("No response!" )

		if !retry {
			println("Retrying...")
			votePage(num, true)
		} else {
			println("No more retries.")
			return
		}
	}

	var posts map[string] interface{} // make a map of key:string value:interface{}
	err := json.Unmarshal(resp, &posts)
	if err != nil {
		panic(err)
	}

	// type: map of arrays of maps
	items := (posts["data"]).(map[string] interface{})["items"].([]interface{})

	for _, elem := range items {
		item, ok := elem.(map[string] interface{})
		if ok {
			if item["vote"] != "up" {
				itemResp := request("POST", fmt.Sprintf("gallery/%s/vote/up", item["id"]), nil)

				// if time out, retry
				if itemResp == nil {
					fmt.Printf("No response for %s, retrying!", item["id"])
					itemResp = request("POST", fmt.Sprintf("gallery/%s/vote/up", item["id"]), nil)

					if itemResp == nil {
						println("Giving up on retrying.")
						continue
					}
				}

				var itemJson map[string] interface{}
				err := json.Unmarshal(itemResp, &itemJson)
				if err != nil {
					fmt.Println("Error parsing: " + string(itemResp))
				}
				println("Voted up: imgur.com/" + item["id"].(string))
			}
		}
	}
}

func request(method string, endpoint string, params url.Values) []byte {

	defer func() {
		err := recover()

		if err != nil {
			println("Error: " + err.(error).Error())
		}
	}()

	req, err := http.NewRequest(method, "https://api.imgur.com/3/" + endpoint, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Add("Authorization", "Bearer " + AccessToken)
	req.Header.Add("Accept", "application/json")

	if params != nil {
		if strings.ToUpper(method) == "GET" {
			req.URL.RawQuery = params.Encode()
		}

		if strings.ToUpper(method) == "POST" {
			req.PostForm = params
		}
	}

	client := http.Client{Timeout: 5 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	header := resp.Header.Get("X-RateLimit-UserRemaining")

	if header == "" {
		header = resp.Header.Get("X-RateLimit-ClientRemaining")
	}

	if header != "" {
		rateLimitRemaining, err := strconv.ParseInt(header, 10, 0)

		if err != nil {
			panic(err)
		}
		if rateLimitRemaining < 1000 {
			println("Rate limit getting low: " + string(rateLimitRemaining))
		}
	}

	header = resp.Header.Get("X-RateLimit-UserRemaining")

	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	return bytes
}