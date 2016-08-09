package lookup

import (
	"fmt"
	"time"

	"github.com/ajanicij/goduckgo/goduckgo"
	"golang.org/x/net/context"
)

func DuckduckQuery(ctx context.Context, question string) (string, error) {
	if deadline, ok := ctx.Deadline(); ok {
		// if the deadline has passed return
		fmt.Println("DEADLINE:", deadline)
		if time.Now().After(deadline) {
			return "", ctx.Err()
		}
	}

	type responseAndError struct {
		resp string
		err  error
	}

	respChan := make(chan responseAndError)

	go func() {
		resp, err := goduckgo.Query(question)
		fmt.Println("RESP:", resp)
		result := ""

		if resp != nil {
			result = extractInfo(resp)
		}

		respChan <- responseAndError{result, err}
	}()

	select {
	case r := <-respChan:
		return r.resp, r.err
	case <-ctx.Done(): // if the context is cancelled return
		return "", ctx.Err()
	}
}

func extractInfo(resp *goduckgo.Message) string {
	if len(resp.Answer) > 0 {
		fmt.Println("Setting answer:", resp.Answer)
		return resp.Answer
	}
	if len(resp.Definition) > 0 {
		fmt.Println("Setting def:", resp.Definition)
		return resp.Definition
	}
	if len(resp.AbstractText) > 0 {
		fmt.Println("Setting abstract:", resp.Abstract)
		return resp.AbstractText
	}
	if len(resp.Results) > 0 {
		fmt.Println("Setting result:", resp.Abstract)
		//index := rand.Intn(len(resp.Results) - 1)
		return resp.Results[0].Text
	}
	if len(resp.RelatedTopics) > 0 {
		fmt.Println("Setting related topic:", resp.Abstract)
		//index := rand.Intn(len(resp.RelatedTopics) - 1)
		return resp.RelatedTopics[0].Text
	}
	return ""
}
