package lookup

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	"github.com/pcrawfor/golanguk/samples/app/session"

	"context"

	"golang.org/x/net/context/ctxhttp"
)

const apiPath = "http://api.giphy.com/v1/gifs/search"

func afterDeadline(ctx context.Context) bool {
	if deadline, ok := ctx.Deadline(); ok {
		if time.Now().After(deadline) {
			return true
		}
	}

	return false
}

func GifForTerms(ctx context.Context, terms []string, apiKey string) (string, error) {
	if afterDeadline(ctx) {
		return "", ctx.Err()
	}

	rating := "r"
	s, ok := session.FromContext(ctx)
	if ok {
		rating = ratingForUser(s)
	}

	termsString := strings.Join(terms, "+")
	params := map[string]interface{}{"api_key": apiKey, "q": termsString, "rating": rating}
	resp, err := getGiphy(ctx, apiPath, params)
	if err != nil {
		return "", err
	}

	url, perr := parseResponse(resp)
	if perr != nil {
		return "", perr
	}

	fmt.Println("URL:", url)

	return url, nil
}

func getGiphy(ctx context.Context, path string, params map[string]interface{}) (*http.Response, error) {
	// if the params are not nil convert them into a url query params
	var requestUrl string

	if params != nil {
		queryParams := url.Values{}
		for k, v := range params {
			queryParams.Add(k, v.(string))
		}
		requestUrl = path + "?" + queryParams.Encode()
	} else {
		requestUrl = path
	}

	fmt.Println("Request url:", requestUrl)

	client := http.Client{}
	return ctxhttp.Get(ctx, &client, requestUrl)
}

type giphyResponse struct {
	Data []data `json:"data"`
}

type data struct {
	Images imageData `json:"images"`
}

type imageData struct {
	FixedHeight fixedHeightData `json:"fixed_height"`
}

type fixedHeightData struct {
	Url string `json:"url"`
}

func parseResponse(resp *http.Response) (string, error) {
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	gresp := giphyResponse{}
	e := json.Unmarshal(body, &gresp)
	if e != nil {
		return "", e
	}

	result := ""
	if len(gresp.Data) > 0 {
		index := rand.Intn(len(gresp.Data))
		result = gresp.Data[index].Images.FixedHeight.Url
	}

	return result, nil
}

func ratingForUser(s *sessions.Session) string {
	email, ok := session.Email(s)
	if ok && email == "paul@dailyburn.com" {
		return "pg-13"
	}
	return "pg"
}
