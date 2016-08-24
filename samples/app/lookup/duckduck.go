package lookup

import (
	"fmt"

	"context"

	"github.com/ajanicij/goduckgo/goduckgo"
)

func DuckduckQuery(ctx context.Context, question string) ([]string, error) {
	type responseAndError struct {
		resp []string
		err  error
	}

	respChan := make(chan responseAndError)

	go func() {
		resp, err := goduckgo.Query(question)

		fmt.Println("RESP:", resp)
		var result []string

		if resp != nil {
			result = combineResults(resp)
		}

		respChan <- responseAndError{result, err}
		return
	}()

	select {
	case r := <-respChan:
		return r.resp, r.err
	case <-ctx.Done(): // if the context is cancelled return
		return nil, ctx.Err()
	}
}

func combineResults(resp *goduckgo.Message) []string {
	results := []string{}
	switch {
	case len(resp.Answer) > 0:
		results = append(results, resp.Answer)
	case len(resp.Definition) > 0:
		results = append(results, resp.Definition)
	case len(resp.AbstractText) > 0:
		results = append(results, resp.AbstractText)
	case len(resp.Results) > 0:
		for _, v := range resp.Results {
			if len(v.Text) > 0 {
				results = append(results, v.Text)
			}
		}
	case len(resp.RelatedTopics) > 0:
		for _, v := range resp.RelatedTopics {
			if len(v.Text) > 0 {
				results = append(results, v.Text)
			}
		}
	}

	return results
}
