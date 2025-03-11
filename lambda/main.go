package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"

	"github.com/joho/godotenv"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type Request struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

type Message struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Message     string `json:"message"`
	Time        int64  `json:"time"`
	EmotionType string `json:"emotion_type"`
}

var (
	BucketName = os.Getenv("BUCKET_NAME")
	FileKey    = "messages.json"
)

func main() {
	lambda.Start(handleRequest)
}

func handleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Load environment variables for local development
	if os.Getenv("BUCKET_NAME") == "" {
		_ = godotenv.Load()
	}

	sdkConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return serverError("Failed to load AWS config")
	}
	s3Client := s3.NewFromConfig(sdkConfig)

	// Route API requests
	switch {
	case request.HTTPMethod == "POST" && strings.HasPrefix(request.Path, "/message"):
		return handleStoreMessage(ctx, request, s3Client)
	case request.HTTPMethod == "GET" && strings.HasPrefix(request.Path, "/message"):
		return handleGetMessages(ctx, s3Client)
	default:
		return clientError("Invalid route")
	}
}

func handleStoreMessage(ctx context.Context, request events.APIGatewayProxyRequest, s3Client *s3.Client) (events.APIGatewayProxyResponse, error) {
	var req Request
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return clientError("Invalid JSON request")
	}

	emotionType := checkMessageEmotionalType(req.Message)

	newMessage := Message{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Email:       req.Email,
		Message:     req.Message,
		EmotionType: emotionType,
		Time:        unixTimestamp(),
	}

	// Fetch existing messages
	messages, err := fetchMessagesFromS3(ctx, s3Client)
	if err != nil {
		return serverError("Failed to fetch messages")
	}

	// Append new message & keep only last 20
	messages = append(messages, newMessage)
	if len(messages) > 20 {
		messages = messages[len(messages)-20:]
	}

	// Upload updated messages back to S3
	err = uploadMessagesToS3(ctx, s3Client, messages)
	if err != nil {
		return serverError("Failed to store message")
	}

	return successResponse("Message stored successfully")
}

func handleGetMessages(ctx context.Context, s3Client *s3.Client) (events.APIGatewayProxyResponse, error) {
	messages, err := fetchMessagesFromS3(ctx, s3Client)
	if err != nil {
		return serverError("Failed to retrieve messages")
	}

	jsonData, _ := json.Marshal(messages)
	return createResponse(200, string(jsonData))
}

func fetchMessagesFromS3(ctx context.Context, s3Client *s3.Client) ([]Message, error) {
	resp, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(BucketName),
		Key:    aws.String(FileKey),
	})
	if err != nil {
		log.Printf("File not found in S3, creating new one...")
		return []Message{}, nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var messages []Message
	if err := json.Unmarshal(body, &messages); err != nil {
		return nil, err
	}

	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Time > messages[j].Time
	})

	return messages, nil
}

func uploadMessagesToS3(ctx context.Context, s3Client *s3.Client, messages []Message) error {
	jsonData, err := json.MarshalIndent(messages, "", "  ")
	if err != nil {
		return err
	}

	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(BucketName),
		Key:    aws.String(FileKey),
		Body:   bytes.NewReader(jsonData),
	})
	return err
}

func createResponse(statusCode int, body string) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "OPTIONS, GET, POST",
			"Access-Control-Allow-Headers": "Content-Type",
		},
		Body: body,
	}, nil
}

func successResponse(msg string) (events.APIGatewayProxyResponse, error) {
	return createResponse(200, msg)
}

func clientError(msg string) (events.APIGatewayProxyResponse, error) {
	return createResponse(400, msg)
}

func serverError(msg string) (events.APIGatewayProxyResponse, error) {
	return createResponse(500, msg)
}

func unixTimestamp() int64 {
	return time.Now().Unix()
}

func checkMessageEmotionalType(message string) string {
	var OPENAI_API_KEY string = os.Getenv("OPENAI_API_KEY")

	client := openai.NewClient(
		option.WithAPIKey(OPENAI_API_KEY),
	)
	chatCompletion, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(message),
		}),
		Tools: openai.F([]openai.ChatCompletionToolParam{
			{
				Type: openai.F(openai.ChatCompletionToolTypeFunction),
				Function: openai.F(openai.FunctionDefinitionParam{
					Name:        openai.String("check_message_emotional_type"),
					Description: openai.String(`Check the emotional type of the message and return only one from: "happy", "love", "angry", "sad", "afraid", "bored", or "calm".`),
					Parameters: openai.F(openai.FunctionParameters{
						"type": "object",
						"properties": map[string]interface{}{
							"emotional_types": map[string]interface{}{
								"type": "array",
								"items": map[string]interface{}{
									"type": "string",
									"enum": []string{"happy", "love", "angry", "sad", "afraid", "bored", "calm"},
								},
							},
						},
						"required": []string{"emotional_types"},
					}),
				}),
			},
		}),
		Model: openai.F(openai.ChatModelGPT4oMini),
	})

	if err != nil {
		log.Printf("OpenAI API error: %v", err)
		return "unknown"
	}

	if len(chatCompletion.Choices) == 0 || chatCompletion.Choices[0].Message.ToolCalls == nil {
		log.Printf("No tool call detected in OpenAI response")
		return "unknown"
	}

	toolResponse := chatCompletion.Choices[0].Message.ToolCalls[0].Function.Arguments
	var result struct {
		EmotionalTypes []string `json:"emotional_types"`
	}
	err = json.Unmarshal([]byte(toolResponse), &result)
	if err != nil {
		log.Printf("Failed to parse emotional types: %v", err)
		return "unknown"
	}

	if len(result.EmotionalTypes) == 0 {
		return "unknown"
	}

	return result.EmotionalTypes[0]
}
