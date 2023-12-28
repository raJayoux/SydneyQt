package main

import (
	"errors"
	"strings"
	"sydneyqt/sydney"
	"time"
)

var (
	ErrMissingPrompt   = errors.New("user prompt is missing (last message is not sent by user)")
	FinishReasonStop   = "stop"
	FinishReasonLength = "length"
)

func ParseOpenAIMessages(messages []OpenAIMessage) (OpenAIMessagesParseResult, error) {
	if len(messages) == 0 {
		return OpenAIMessagesParseResult{}, nil
	}

	// last message must be user prompt
	var index int
	var lastMessage OpenAIMessage

	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			index = i
			lastMessage = messages[i]
			break
		}
	}
	// exclude the lastMessage from the array
	messages = append(messages[:index], messages[index+1:]...)
	
	//lastMessage := messages[len(messages)-1]
	prompt, imageUrl := ParseOpenAIMessageContent(lastMessage.Content)
	
	if prompt == "" {
		return OpenAIMessagesParseResult{}, ErrMissingPrompt
	}

	if len(messages) == 1 {
		return OpenAIMessagesParseResult{
			Prompt:   prompt,
			ImageURL: imageUrl,
		}, nil
	}

	// construct context
	var contextBuilder strings.Builder
	contextBuilder.WriteString("\n\n")
	
	for i, message := range messages[:len(messages)] {
		// assert types
		text, _ := ParseOpenAIMessageContent(message.Content)

		// append role to context
		switch message.Role {
		case "user":
			contextBuilder.WriteString("[user](#message)\n")
		case "assistant":
			contextBuilder.WriteString("[assistant](#message)\n")
		case "system":
			contextBuilder.WriteString("[system](#additional_instructions)\n")
		default:
			continue // skip unknown roles
		}

		// append content to context
		contextBuilder.WriteString(text)
		if i != len(messages)-1 {
			contextBuilder.WriteString("\n\n")
		}
	}

	return OpenAIMessagesParseResult{
		Prompt:         prompt,
		WebpageContext: contextBuilder.String(),
		ImageURL:       imageUrl,
	}, nil
}

func ParseOpenAIMessageContent(content interface{}) (text, imageUrl string) {
	switch content := content.(type) {
	case string:
		// content is string, and it automatically becomes prompt
		text = content
	case []map[string]interface{}:
		// content is array of objects, and it contains prompt and optional image url
		for _, content := range content {
			switch content["type"] {
			case "text":
				if contentText, ok := content["text"].(string); ok {
					text = contentText
				}
			case "image_url":
				if url, ok := content["image_url"].(map[string]string); ok {
					imageUrl = url["url"]
				}
			}
		}
	}

	return
}

func NewOpenAIChatCompletion(model, content, finishReason string) *OpenAIChatCompletion {
	return &OpenAIChatCompletion{
		ID:                "chatcmpl-123",
		Object:            "chat.completion",
		Created:           time.Now().Unix(),
		Model:             model,
		SystemFingerprint: "fp_44709d6fcb",
		Choices: []ChatCompletionChoice{
			{
				Index: 0,
				Message: ChoiceMessage{
					Role:    "assistant",
					Content: content,
				},
				FinishReason: finishReason,
			},
		},
		Usage: UsageStats{
			PromptTokens:     1024,
			CompletionTokens: 1024,
			TotalTokens:      2048,
		},
	}
}

func NewOpenAIChatCompletionChunk(model, delta string, finishReason *string) *OpenAIChatCompletionChunk {
	return &OpenAIChatCompletionChunk{
		ID:                "chatcmpl-123",
		Object:            "chat.completion",
		Created:           time.Now().Unix(),
		Model:             model,
		SystemFingerprint: "fp_44709d6fcb",
		Choices: []ChatCompletionChunkChoice{
			{
				Index: 0,
				Delta: ChoiceDelta{
					Role:    "assistant",
					Content: delta,
				},
				FinishReason: finishReason,
			},
		},
	}
}

func ToOpenAIImageGeneration(result sydney.GenerateImageResult) OpenAIImageGeneration {
	var objects []OpenAIImageObject

	for _, url := range result.ImageURLs {
		urlWithoutQuery := strings.Split(url, "?")[0]
		objects = append(objects, OpenAIImageObject{
			URL:           urlWithoutQuery,
			RevisedPrompt: result.Text,
		})
	}

	return OpenAIImageGeneration{
		Created: time.Now().Unix(),
		Data:    objects,
	}
}
