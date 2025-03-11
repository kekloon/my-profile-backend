# Edmond Wong My Profile Backend

## Overview

**My Profile Backend** is a serverless application designed to show my profile and store user messages with emotional analysis. It leverages AWS services such as Lambda, S3, and API Gateway, along with OpenAI's API for emotion detection.

## Features

- **Message Storage**: Store user messages in an S3 bucket.
- **Emotion Analysis**: Analyze the emotional tone of messages using OpenAI's API.
- **API Gateway**: Expose RESTful endpoints for message submission and retrieval.

## Architecture

The application is built using Terraform for infrastructure as code, deploying resources on AWS with Dockerfile. It includes:

- **AWS Lambda**: Processes incoming requests and interacts with S3 and OpenAI.
- **Amazon S3**: Stores messages in JSON format.
- **Amazon API Gateway**: Provides RESTful API endpoints.
- **Amazon ECR**: Hosts Docker images for the Lambda function.

## Prerequisites

- AWS account with necessary permissions.
- Docker installed for building and deploying the Lambda function.
- OpenAI API key.

## Installation

1. **Clone the repository**:
   ```bash
   git clone <repository-url>
   cd my-profile-backend
   ```

2. **Set up environment variables**:
   Create a `.env` file with the following variables:
   ```plaintext
   AWS_ACCESS_KEY_ID=your-access-key-id
   AWS_SECRET_ACCESS_KEY=your-secret-access-key
   OPENAI_API_KEY=your-openai-api-key
   ```

3. **Initialize Terraform**:
   ```bash
   make init
   ```

4. **Deploy the infrastructure**:
   ```bash
   make apply
   ```

## Usage

- **Submit a Message**: Send a POST request to `/message` with a JSON body containing `name`, `email`, and `message`.
- **Retrieve Messages**: Send a GET request to `/message` to retrieve the last 20 messages.
