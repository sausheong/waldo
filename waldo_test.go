package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"testing"
)

func TestParseImage(t *testing.T) {
	line := " /Users/sausheong/go/src/github.com/sausheong/multimodal/test-images/uni.jpg what is this image?"
	q, err := parseImageQuery("llama2:13b", line)
	if err != nil {
		t.Error(err)
	}
	fmt.Println("query:", q.Query)
	fmt.Println("images:", q.Images)
}

func TestDisplayInline(t *testing.T) {
	file := "/Users/sausheong/go/src/github.com/sausheong/multimodal/test-images/fruits.jpg"
	bytes, _ := os.ReadFile(file)
	b64 := base64.StdEncoding.EncodeToString(bytes)
	fmt.Printf("\x1b]1337;File=inline=1;width=256px:%s\a\n", b64)
}

func TestCallGPT4Vision(t *testing.T) {
	img1 := "/Users/sausheong/go/src/github.com/sausheong/multimodal/test-images/fruits.jpg"
	img2 := "/Users/sausheong/go/src/github.com/sausheong/multimodal/test-images/uni.jpg"
	file1, _ := os.ReadFile(img1)
	file2, _ := os.ReadFile(img2)

	b64 := []string{base64.StdEncoding.EncodeToString(file1), base64.StdEncoding.EncodeToString(file2)}
	results, err := callGPT4Vision("what are these images about?", b64)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(results)

}
