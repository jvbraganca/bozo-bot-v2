package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

func searchAndRetweet(client *twitter.Client) {

	log.Print("Running ", time.Now().String())
	tweets, _, err := client.Search.Tweets(&twitter.SearchTweetParams{
		Query:     "fora bolsonaro",
		TweetMode: "omitempty",
		Count:     180,
	})

	if err != nil {
		log.Fatal(err)
	}

	for _, val := range tweets.Statuses {
		if doesTweetContainSearchWords(val.Text) {
			retweet(client.Statuses, val.ID)
		}
	}
}

func retweetHandler(w http.ResponseWriter, r *http.Request, client *twitter.Client) {
	searchAndRetweet(client)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := make(map[string]string)
	resp["message"] = "Bot ran successfully!"
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("Error while marshalling response: %v", err)
	}
	w.Write(jsonResp)
}

func retweet(client *twitter.StatusService, tweetId int64) {
	_, _, err := client.Retweet(tweetId, nil)
	if err != nil {
		log.Print("Unable to retweet: ", tweetId)
	} else {
		log.Print("Retweeted: ", tweetId)
	}
}

func doesTweetContainSearchWords(text string) bool {
	match, _ := regexp.MatchString("(?mi)fora.bolsonaro|forabolsonaro|fora..bolsonaro", text)
	return match
}

func main() {
	httpInvokerPort, exists := os.LookupEnv("FUNCTIONS_CUSTOMHANDLER_PORT")
	if exists {
		fmt.Println("FUNCTIONS_HTTPWORKER_PORT: " + httpInvokerPort)
	} else {
		httpInvokerPort = "8080"
	}
	config := oauth1.NewConfig(os.Getenv("TWITTER_API_KEY"), os.Getenv("TWITTER_API_SECRET"))
	token := oauth1.NewToken(os.Getenv("TWITTER_ACCESS_TOKEN"), os.Getenv("TWITTER_ACCESS_TOKEN_SECRET"))
	httpClient := config.Client(oauth1.NoContext, token)

	twitterClient := twitter.NewClient(httpClient)

	customHandler := func(w http.ResponseWriter, r *http.Request) {
		retweetHandler(w, r, twitterClient)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/bozoBot", customHandler)
	log.Println("Go server Listening...on httpInvokerPort:", httpInvokerPort)
	log.Fatal(http.ListenAndServe(":"+httpInvokerPort, mux))
}
