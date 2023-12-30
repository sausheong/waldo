package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/generative-ai-go/genai"
	"github.com/hako/durafmt"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
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

func predict(model string, prompt string, ctx string, format string) error {
	switch model {
	case "gpt-3.5-turbo":
		return gpt(model, prompt, ctx, format)
	case "gpt-4":
		return gpt(model, prompt, ctx, format)
	case "gpt-4-turbo":
		return gpt("gpt-4-1106-preview", prompt, ctx, format)
	case "gpt-4-vision":
		return gpt("gpt-4-vision-preview", prompt, ctx, format)
	case "gemini-pro":
		return gemini(model, prompt, ctx, format)
	default:
		return ollama(model, prompt, ctx, format)
	}
}

// Call OpenAI APIs to predict
// uses langchaingo
func gpt(model string, prompt string, ctx string, format string) error {
	t0 := time.Now()
	llm, err := openai.NewChat(openai.WithModel(model))
	if err != nil {
		return err
	}
	c := context.Background()
	_, err = llm.Call(c, []schema.ChatMessage{
		schema.SystemChatMessage{Content: prompt},
		schema.HumanChatMessage{Content: ctx},
	}, llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
		fmt.Print(string(chunk))
		return nil
	}), llms.WithMinLength(1024),
	)
	if err != nil {
		return err
	}
	elapsed := durafmt.Parse(time.Since(t0)).LimitFirstN(2)
	fmt.Printf(cyan("\n\n(%s)"), elapsed)
	fmt.Println()

	return nil
}

// call Gemini API to predict
func gemini(model string, prompt string, ctx string, format string) error {
	t0 := time.Now()
	client, err := genai.NewClient(context.Background(), option.WithAPIKey(os.Getenv("GOOGLEAI_API_KEY")))
	if err != nil {
		fmt.Println("cannot create Gemini client:", err)
		return err
	}
	defer client.Close()

	gemini := client.GenerativeModel(model)
	iter := gemini.GenerateContentStream(c, genai.Text(prompt), genai.Text(ctx))
	for {
		resp, err := iter.Next()
		if err == iterator.Done {
			elapsed := durafmt.Parse(time.Since(t0)).LimitFirstN(2)
			fmt.Printf(cyan("\n\n(%s)"), elapsed)
			fmt.Println()
			break
		}
		if err != nil {
			fmt.Println("cannot generate content:", err)
			return err
		}
		for _, cand := range resp.Candidates {
			if cand.Content != nil {
				for _, part := range cand.Content.Parts {
					if part != nil {
						fmt.Print(part)
					}
				}
			}
		}
	}
	return nil
}

// predict by calling Ollama with a given model
func ollama(model string, prompt string, ctx string, format string) error {
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

func qa(model string, ctx string, filepath string) error {
	prompt := `You are a helpful AI image assistant who can answer questions about
a given image.`

	req := &CompletionRequest{
		Model:  model,
		Prompt: prompt,
		System: ctx,
		Stream: true,
	}

	// Read the entire file into a byte slice
	file, err := os.ReadFile(filepath)
	if err != nil {
		fmt.Println("err in getting bytes from image file", err, filepath)
	}
	req.Images = []string{base64.StdEncoding.EncodeToString(file)}

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
		fmt.Print(yellow(resp.Response))
		if resp.Done {
			elapsed := durafmt.Parse(time.Since(t0)).LimitFirstN(2)
			fmt.Printf(cyan("\n\n(%s)"), elapsed)
			break
		}
	}
	return err
}

func query(model string, ctx string) (ImageQuery, error) {
	query := ImageQuery{}
	prompt := `You are a helpful AI image assistant who can answer questions about
a given image. When the user asks a question about the image Return the JSON response 
with the following format:
--
{
	"query" : <query from the user>
	"filepath": <the path of the file to be queried upon>
}
`
	req := &CompletionRequest{
		Model:  model,
		Prompt: prompt,
		System: ctx,
		Stream: false,
		Format: "json",
	}

	reqJson, err := json.Marshal(req)
	if err != nil {
		fmt.Println("err in marshaling:", err)
		return query, err
	}

	r := bytes.NewReader(reqJson)
	httpResp, err := http.Post("http://localhost:11435/api/generate", "application/json", r)
	if err != nil {
		fmt.Println("err in calling ollama:", err)
		return query, err
	}
	decoder := json.NewDecoder(httpResp.Body)
	resp := CompletionResponse{}
	err = decoder.Decode(&resp)
	if err != nil {
		log.Println("Cannot decode completion response:", err)
		return query, err
	}
	err = json.Unmarshal([]byte(resp.Response), &query)
	if err != nil {
		log.Println("Cannot unmarshal image query:", err)
	}
	return query, err
}
