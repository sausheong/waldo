package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/hako/durafmt"
)

func run(tool string, args []string) ([]byte, error) {
	fmt.Println(yellow("executing>"), green(tool), green(strings.Join(args, " ")))
	cmd := exec.Command(tool, args...)
	return cmd.CombinedOutput()
}

func search(model string, query string) error {
	data, result, err := ddg(query)
	if err != nil {
		log.Println("Cannot process query:", err)
		return err
	}
	prompt := `The following search results has come back from a search engine given the query 
that came from a user. Respond to the original query using the search results. Do not add any 
additional information. Assume the person you are explaining to doesn't know anything about 
the answer.	End the response with a list of URLs returned.`
	ctx := `{
	"query" : "` + query + `",
	"search_result" : "` + data + `",
	"urls" : "` + fmt.Sprint(result) + `"
}`
	return predict(model, prompt, ctx, "")
}

func ask(model string, query string) error {
	prompt := `Give immediate, precise and clear answers to questions asked. If you do not know 
the answer, say "I don't know the answer to this.".
`
	return predict(model, prompt, query, "")
}

// predict by calling Ollama with a given model
func predict(model string, prompt string, ctx string, format string) error {
	req := &CompletionRequest{
		Model:  model,
		Prompt: prompt,
		System: ctx,
		Stream: true,
	}
	if format == "json" {
		req.Format = "json"
	}

	reqJson, err := json.Marshal(req)
	if err != nil {
		fmt.Println("err in marshaling:", err)
		return err
	}

	r := bytes.NewReader(reqJson)
	httpResp, err := http.Post("http://localhost:11435/api/generate", "application/json", r)
	if err != nil {
		fmt.Println("err in calling ollama:", err)
		return err
	}
	decoder := json.NewDecoder(httpResp.Body)
	t0 := time.Now()
	for {
		resp := &CompletionResponse{}
		decoder.Decode(&resp)
		fmt.Print(resp.Response)
		if resp.Done {
			elapsed := durafmt.Parse(time.Since(t0)).LimitFirstN(2)
			fmt.Printf(cyan("\n\n(%s)"), elapsed)
			fmt.Println()
			break
		}
	}
	return err
}

// use DuckDuckGo to search the Internet and return the top 5 results
func ddg(query string) (string, []SearchResult, error) {
	queryURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", url.QueryEscape(query))
	userAgent := "Mozilla/5.0 (Macintosh; Intel Mac OS X 14_1) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Safari/605.1.15"
	client := &http.Client{}
	request, err := http.NewRequest("GET", queryURL, nil)
	if err != nil {
		return "", []SearchResult{}, err
	}
	request.Header.Set("User-Agent", userAgent)
	response, err := client.Do(request)
	if err != nil {
		return "", []SearchResult{}, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", []SearchResult{}, fmt.Errorf("status is %d", response.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return "", []SearchResult{}, err
	}

	results := []SearchResult{}
	sel := doc.Find(".web-result")
	// text := ""
	for i := range sel.Nodes {
		if len(results) > 5 {
			break
		}

		node := sel.Eq(i)
		titleNode := node.Find(".result__a")
		info := node.Find(".result__snippet").Text()
		title := titleNode.Text()
		url, _ := node.Find(".result__snippet").Attr("href")
		results = append(results, SearchResult{title, info, url})
	}
	formattedResults := ""

	for _, result := range results {
		formattedResults += fmt.Sprintf("Title: %s\n"+
			"Description: %s\n\n", result.Title, result.Info)
	}
	return formattedResults, results, nil
}
