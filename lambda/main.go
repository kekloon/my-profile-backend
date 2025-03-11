package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

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

func main() {
	if os.Getenv("BUCKET_NAME") == "" {
		_ = godotenv.Load()
	}

	lambda.Start(handleRequest)
}

func handleRequest(ctx context.Context, request Request) (string, error) {
	var BUCKET_NAME string = os.Getenv("BUCKET_NAME")

	fmt.Println("BUCKET_NAME", BUCKET_NAME)
	fmt.Println("Hello Sent From Lambda!", request)

	// Check the emotional type of the message with OpenAI
	emotionalType := checkMessageEmotionalType(request.Message)
	fmt.Println("Emotional Type:", emotionalType)

	// Create a JSON file with the data
	jsonData := map[string]string{
		"name":           request.Name,
		"email":          request.Email,
		"message":        request.Message,
		"emotional_type": emotionalType,
	}

	jsonFile, err := os.Create("/tmp/output.json")
	if err != nil {
		return "", fmt.Errorf("failed to create json file: %v", err)
	}
	defer jsonFile.Close()

	encoder := json.NewEncoder(jsonFile)
	if err := encoder.Encode(jsonData); err != nil {
		return "", fmt.Errorf("failed to encode json data: %v", err)
	}

	// Upload the JSON file to S3
	sdkConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		fmt.Println("Couldn't load default configuration. Have you set up your AWS account?")
		fmt.Println(err)
		return "", fmt.Errorf("failed to load default configuration: %v", err)
	}

	s3Client := s3.NewFromConfig(sdkConfig)
	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(BUCKET_NAME),
		Key:    aws.String(fmt.Sprintf("%s-%s", uuid.New().String(), jsonFile.Name())),
		Body:   jsonFile,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %v", err)
	}

	return "Successfully submitted!", nil
}

func checkMessageEmotionalType(Message string) string {
	var OPENAI_API_KEY string = os.Getenv("OPENAI_API_KEY")

	client := openai.NewClient(
		option.WithAPIKey(OPENAI_API_KEY),
	)
	chatCompletion, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(Message),
		}),
		Tools: openai.F([]openai.ChatCompletionToolParam{
			{
				Type: openai.F(openai.ChatCompletionToolTypeFunction),
				Function: openai.F(openai.FunctionDefinitionParam{
					Name:        openai.String("check_message_emotional_type"),
					Description: openai.String(`Check the emotional type of the message only return "positive" or "negative"`),
					Parameters: openai.F(openai.FunctionParameters{
						"type": "object",
						"properties": map[string]interface{}{
							"emotional_type": map[string]string{
								"type": "string",
							},
						},
						"required": []string{"emotional_type"},
					}),
				}),
			},
		}),
		Model: openai.F(openai.ChatModelGPT4oMini),
	})
	if err != nil {
		panic(err.Error())
	}
	return chatCompletion.Choices[0].Message.Content
}
