package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/h2non/filetype"
	"github.com/hako/durafmt"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

func parseImageQuery(model string, ctx string) (ImageQuery, error) {
	query := ImageQuery{}
	prompt := `The input has an instruction or a query, and also one or more image files. Parse and 
return the JSON response  with the following format:
--
{
	"query" : <instruction/query from the user or an empty string "">
	"images": <the path of one or more image files in an array [] or an empty array []>
}
--
If the image files are not provided the images is an empty array []. If there is no question
or instruction the query is an empty string "".
##
`
	req := &CompletionRequest{
		Model:  "llama2:13b",
		Prompt: prompt + ctx,
		System: "",
		Stream: false,
		Format: "json",
	}

	reqJson, err := json.Marshal(req)
	if err != nil {
		fmt.Println("err in marshaling:", err)
		return query, err
	}

	r := bytes.NewReader(reqJson)
	httpResp, err := http.Post("http://localhost:11434/api/generate", "application/json", r)
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
	fmt.Println(resp.Response)
	err = json.Unmarshal([]byte(resp.Response), &query)
	if err != nil {
		log.Println("Cannot unmarshal image query:", err)
	}
	return query, err
}

func askImage(model string, query string, images []string) error {
	prompt := `Answer the question about a given image. Provide clear details in paragraph form, do not answer in point form or with numbered bullets. Only answer what you know, do not add any additional details that you do not have the answer to.`
	switch model {
	case "gpt-4-vision":
		return gptImage("gpt-4-vision-preview", prompt, query, images)
	case "gemini-pro-vision":
		return geminiImage(model, prompt, query, images)
	default:
		return ollamaImage(model, prompt, query, images)
	}

}

// answer questions on images
func ollamaImage(model string, prompt string, ctx string, images []string) error {
	req := &CompletionRequest{
		Model:  model,
		Prompt: prompt,
		System: ctx,
		Images: []string{},
		Stream: true,
	}
	for _, img := range images {
		file, err := os.ReadFile(img)
		if err != nil {
			fmt.Println("err in getting bytes from image file", err, img)
		}
		req.Images = append(req.Images, base64.StdEncoding.EncodeToString(file))
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
			fmt.Printf(cyan("\n\n(%s)\n"), elapsed)
			break
		}
	}
	return err
}

func geminiImage(model string, prompt string, ctx string, images []string) error {
	t0 := time.Now()
	client, err := genai.NewClient(context.Background(), option.WithAPIKey(os.Getenv("GOOGLEAI_API_KEY")))
	if err != nil {
		fmt.Println("cannot create Gemini client:", err)
		return err
	}
	defer client.Close()

	parts := []genai.Part{}
	for _, img := range images {
		imgData, err := os.ReadFile(img)
		if err != nil {
			fmt.Println("cannot read image file:", err)
			return err
		}
		kind, _ := filetype.Match(imgData)
		parts = append(parts, genai.ImageData(kind.MIME.Subtype, imgData))
	}
	parts = append(parts, genai.Text(prompt))
	parts = append(parts, genai.Text(ctx))

	gemini := client.GenerativeModel(model)
	iter := gemini.GenerateContentStream(context.Background(), parts...)
	for {
		resp, err := iter.Next()
		if err == iterator.Done {
			elapsed := durafmt.Parse(time.Since(t0)).LimitFirstN(2)
			fmt.Printf(cyan("\n\n(%s)\n"), elapsed)
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
						s := fmt.Sprint(part)
						if strings.TrimSpace(s) != "" {
							fmt.Print(s)
						}
					}
				}
			}
		}
	}
	return nil
}

func gptImage(model string, prompt string, ctx string, images []string) error {
	t0 := time.Now()
	pmpt := prompt + " ## " + ctx

	b64s := []string{}
	for _, img := range images {
		file, err := os.ReadFile(img)
		if err != nil {
			fmt.Println("cannot get read image file:", err)
			return err
		}
		b64s = append(b64s, base64.StdEncoding.EncodeToString(file))
	}
	response, err := callGPT4Vision(pmpt, b64s)
	if err != nil {
		fmt.Println(red("cannot get response from OpenAI:", err))
	}
	fmt.Print(response)
	elapsed := durafmt.Parse(time.Since(t0)).LimitFirstN(2)
	fmt.Printf(cyan("\n\n(%s)\n"), elapsed)
	return nil
}

func callGPT4Vision(prompt string, imagesb64 []string) (string, error) {
	imagePart := ""
	for _, imgb64 := range imagesb64 {
		p := `{
	"type": "image_url",
	"image_url": {
		"url": "` + fmt.Sprintf("data:image/jpeg;base64,%s", imgb64) + `" 
	}
},`
		imagePart += p
	}
	imagePart = imagePart[0 : len(imagePart)-1]
	requestURL := "https://api.openai.com/v1/chat/completions"
	jsonBody := `{
	"model": "gpt-4-vision-preview",
	"messages": [
		{
		"role": "user",
		"content": [
			{
			"type": "text",
			"text": "` + prompt + `"
			},
			` + imagePart + `
		]
		}
	],
	"max_tokens": 1024
}`

	bodyReader := bytes.NewReader([]byte(jsonBody))
	req, err := http.NewRequest(http.MethodPost, requestURL, bodyReader)
	if err != nil {
		fmt.Printf("client: could not create request: %s\n", err)
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("OPENAI_API_KEY")))
	client := http.Client{
		Timeout: 30 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("client: error making http request: %s\n", err)
		return "", err
	}
	resBody, err := io.ReadAll(res.Body)

	if err != nil {
		fmt.Printf("client: could not read response body: %s\n", err)
		return "", err
	}
	response := OpenAIResponse{}
	err = json.Unmarshal(resBody, &response)
	if err != nil {
		return "", err
	}
	return response.Choices[0].Message.Content, nil
}
