package main

import "sydneyqt/sydney"

type CreateConversationRequest struct {
	Cookies string `json:"cookies"`
}

type CreateImageRequest struct {
	Image   sydney.GenerativeImage `json:"image"`
	Cookies string                 `json:"cookies"`
}

type ChatStreamRequest struct {
	Prompt            string   `json:"prompt"`
	WebpageContext    string   `json:"context"`
	Cookies           string   `json:"cookies"`
	ImageURL          string   `json:"imageUrl"`
	NoSearch          bool     `json:"noSearch"`
	UseGPT4Turbo      bool     `json:"gpt4turbo"`
	UseClassic        bool     `json:"classic"`
	ConversationStyle string   `json:"conversationStyle"`
	Plugins           []string `json:"plugins"`
}

// The `content` field can have different types
// Example:
//
//	{
//		"role": "user",
//		"content": "Hello!"
//	}
//
// or
//
//	{
//		"role": "user",
//		"content": [
//			{
//				"type": "text",
//				"text": "What’s in this image?"
//			},
//			{
//				"type": "image_url",
//				"image_url": {
//					"url": "https://upload.wikimedia.org/wikipedia/commons/thumb/d/dd/Gfp-wisconsin-madison-the-nature-boardwalk.jpg/2560px-Gfp-wisconsin-madison-the-nature-boardwalk.jpg"
//				}
//			}
//		]
//	}
type OpenAIMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

type OpenAIMessagesParseResult struct {
	WebpageContext string
	Prompt         string
	ImageURL       string
}

// Most fields are omitted due to limitations of the Bing API
type OpenAIChatCompletionRequest struct {
	Model        string                            `json:"model"`
	Messages     []OpenAIMessage                   `json:"messages"`
	Stream       bool                              `json:"stream"`
	ToolChoice   *interface{}                      `json:"tool_choice"`
	Conversation sydney.CreateConversationResponse `json:"conversation"`
}

type ChoiceDelta struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionChunkChoice struct {
	Index        int         `json:"index"`
	Delta        ChoiceDelta `json:"delta"`
	FinishReason *string     `json:"finish_reason"`
}

type OpenAIChatCompletionChunk struct {
	ID                string                      `json:"id"`
	Object            string                      `json:"object"`
	Created           int64                       `json:"created"`
	Model             string                      `json:"model"`
	SystemFingerprint string                      `json:"system_fingerprint"`
	Choices           []ChatCompletionChunkChoice `json:"choices"`
}

type ChoiceMessage struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}

type UsageStats struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ChatCompletionChoice struct {
	Index        int           `json:"index"`
	Message      ChoiceMessage `json:"message"`
	FinishReason string        `json:"finish_reason"`
}

type OpenAIChatCompletion struct {
	ID                string                 `json:"id"`
	Object            string                 `json:"object"`
	Created           int64                  `json:"created"`
	Model             string                 `json:"model"`
	SystemFingerprint string                 `json:"system_fingerprint"`
	Choices           []ChatCompletionChoice `json:"choices"`
	Usage             UsageStats             `json:"usage"`
}

type OpenAIImageObject struct {
	URL           string `json:"url"`
	RevisedPrompt string `json:"revised_prompt"`
}

type OpenAIImageGeneration struct {
	Created int64               `json:"created"`
	Data    []OpenAIImageObject `json:"data"`
}

type OpenAIImageGenerationRequest struct {
	Prompt string `json:"prompt"`
}
