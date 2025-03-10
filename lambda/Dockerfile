FROM golang:1.23 as build

WORKDIR /lambda

COPY lambda/go.mod lambda/go.sum ./
COPY lambda/main.go ./
RUN go build -tags lambda.norpc -o main main.go

FROM public.ecr.aws/lambda/go:latest
COPY --from=build /lambda/main ./main
ENTRYPOINT [ "./main" ]