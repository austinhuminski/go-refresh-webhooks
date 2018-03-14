package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var wg sync.WaitGroup

// ID's of projects from other program
var projectMap = []string{
	"11111111111111",
	"22222222222222",
	"33333333333333",
	"44444444444444",
	"55555555555555",
}

var client = &http.Client{}

type Webhook struct {
	Id int64 `json:"id"`
}

type WebhookResponse struct {
	Webhooks []Webhook `json:"data"`
}

type DeleteResponse struct {
	Data interface{} `json:"data"`
}

type RequestWebhook struct {
	Data RequestData `json:"data"`
}

type RequestData struct {
	Id       int64           `json:"id"`
	Active   string          `json:"active"`
	Resource RequestResource `json:"resource"`
}

type RequestResource struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

func setAuthHeader(req *http.Request) *http.Request {
	// We need to set this header on all three of our request types
	// Use this to eaisly set it up
	req.Header.Set("Authorization", os.Getenv("ACCESS_TOKEN"))
	return req
}

func err_check(err error) {
	// It feels really cumbersome to do this check all the time so I put it
	// in a function. Is this idiomatic go?
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	fmt.Println("Starting in main...")
	// Executes the Handler func
	lambda.Start(Handler)
}

func Handler() {
	/*
		1) GET request to get all current webhook ID's from asana
		2) DELETE each webhook
		3) Resubmit with POST for each ID in projectMap
	*/
	fmt.Println("Made it to handler")
	start := time.Now()

	// Get ID's of all current webhooks
	req, err := http.NewRequest("GET", os.Getenv("ENDPOINT_URL"), nil)
	err_check(err)
	req = setAuthHeader(req)

	resp, err := client.Do(req)
	err_check(err)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	err_check(err)

	var data WebhookResponse
	jsonErr := json.Unmarshal(body, &data)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	wg.Add(len(data.Webhooks))

	for _, webhook := range data.Webhooks {
		go deleteWebhook(webhook.Id)
	}
	wg.Wait()

	wg.Add(len(projectMap))
	for _, projectId := range projectMap {
		go requestWebhook(projectId)
	}
	wg.Wait()
	fmt.Println("DONE! All webhooks should have been refreshed")
	elapsed := time.Since(start)
	fmt.Printf("Executed in: %s\n", elapsed)
}

func deleteWebhook(id int64) {
	// DELETE request for existing webhooks

	fmt.Printf("Deleting webhook: %d\n", id)
	deleteWebhookUrl := fmt.Sprintf("%s/%s", os.Getenv("DELETE_URL"), strconv.FormatInt(id, 10))

	req, err := http.NewRequest("DELETE", deleteWebhookUrl, nil)
	if err != nil {
		log.Fatal(err)
	}
	req = setAuthHeader(req)
	resp, err := client.Do(req)
	err_check(err)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	err_check(err)

	// If successful in deleting the enpoint, it returns: "data: {}"
	// Not sure how to handle that. Annoying to have to assign everything
	// into a struct. Is there another way to just eaisly see and inspect
	// the returned JSON? `if not data:{} then throw error `
	var data DeleteResponse
	json.Unmarshal([]byte(body), &data)
	fmt.Println("DELETED!")
	wg.Done()

}

func requestWebhook(projID string) {
	// Submit POST request to create webhook
	fmt.Printf("Requesting webhook for %s\n", projID)

	data := url.Values{}
	data.Set("target", os.Getenv("TARGET-ENDPOINT-URL"))
	data.Set("resource", projID)

	req, _ := http.NewRequest("POST", os.Getenv("POST_URL"), strings.NewReader(data.Encode()))
	req = setAuthHeader(req)

	// Requirement on API to used x-www-form-urlencoded
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	fmt.Println(resp.Status)
	err_check(err)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	err_check(err)

	var results RequestWebhook
	json.Unmarshal([]byte(body), &results)

	fmt.Println(results)
	wg.Done()
}
