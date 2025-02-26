# syntax=docker/dockerfile:1

# Build the application from source
FROM golang:1.23-alpine3.21 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
COPY secure-connect.zip /app/secure-connect.zip


RUN CGO_ENABLED=0 GOOS=linux go build -o /gqlgen-subscriptions-go

# Run the tests in the container
FROM build-stage AS run-test-stage
RUN go test -v ./...

# Deploy the application binary into a lean image
FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

COPY --from=build-stage /gqlgen-subscriptions-go /gqlgen-subscriptions-go
COPY --from=build-stage /app/secure-connect.zip ./secure-connect.zip

EXPOSE 50051

USER nonroot:nonroot

ENTRYPOINT ["/gqlgen-subscriptions-go"]