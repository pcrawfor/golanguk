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

	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"
)

const apiPath = "http://api.giphy.com/v1/gifs/search"

type Giphy struct {
	apiKey string
}

type GiphyResponse struct {
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

func NewGiphy(apiKey string) *Giphy {
	return &Giphy{apiKey: apiKey}
}

func (g *Giphy) GifForTerms(ctx context.Context, terms []string) (string, error) {
	if deadline, ok := ctx.Deadline(); ok {
		// if the deadline has passed return
		fmt.Println("DEADLINE:", deadline)
		if time.Now().After(deadline) {
			return "", ctx.Err()
		}
	}

	termsString := strings.Join(terms, "+")
	params := map[string]interface{}{"api_key": g.apiKey, "q": termsString}
	resp, err := g.get(ctx, apiPath, params)
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

func (g *Giphy) get(ctx context.Context, path string, params map[string]interface{}) (*http.Response, error) {
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

func parseResponse(resp *http.Response) (string, error) {
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	gresp := GiphyResponse{}
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
