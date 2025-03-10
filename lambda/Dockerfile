FROM public.ecr.aws/lambda/go:latest

COPY main.go .
RUN go mod init lambda-function-go
RUN go mod tidy
RUN GOOS=linux GOARCH=amd64 go build -o bootstrap main.go

CMD [ "bootstrap" ]