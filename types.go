package main

import (
	"time"
)

type SearchResult struct {
	Title string
	Info  string
	Url   string
}

type CommandResponse struct {
	Inputs  []string `json:"inputs"`
	Command string   `json:"command"`
	Outputs []string `json:"outputs"`
}

type CompletionRequest struct {
	Model   string         `json:"model"`
	Prompt  string         `json:"prompt"`
	Images  []string       `images:"system,omitempty"`
	Format  string         `json:"format,omitempty"`
	Options map[string]any `json:"options,omitempty"`
	System  string         `json:"system,omitempty"`
	Context []byte         `json:"context,omitempty"`
	Stream  bool           `json:"stream"`
}

type CompletionResponse struct {
	Model              string        `json:"model"`
	CreatedAt          time.Time     `json:"created_at"`
	Response           string        `json:"response"`
	Done               bool          `json:"done"`
	Context            []int         `json:"context,omitempty"`
	TotalDuration      time.Duration `json:"total_duration,omitempty"`
	LoadDuration       time.Duration `json:"load_duration,omitempty"`
	PromptEvalCount    int           `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration time.Duration `json:"prompt_eval_duration,omitempty"`
	EvalCount          int           `json:"eval_count,omitempty"`
	EvalDuration       time.Duration `json:"eval_duration,omitempty"`
}

type PullResponse struct {
	Status    string `json:"status"`
	Digest    string `json:"digest"`
	Total     int64  `json:"total"`
	Completed int64  `json:"completed"`
}

type ImageQuery struct {
	Query  string   `json:"query"`
	Images []string `json:"images"`
}

type Models struct {
	Models []struct {
		Name       string `json:"name"`
		ModifiedAt string `json:"modified_at"`
		Size       int64  `json:"size"`
	} `json:"models"`
}

// for OpenAI responses
type OpenAIResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int      `json:"created"`
	Model   string   `json:"model"`
	Usage   Usage    `json:"usage"`
	Choices []Choice `json:"choices"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type FinishDetails struct {
	Type string `json:"type"`
	Stop string `json:"stop"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Choice struct {
	Message       Message       `json:"message"`
	FinishDetails FinishDetails `json:"finish_details"`
	Index         int           `json:"index"`
}
