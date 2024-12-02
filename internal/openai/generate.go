package openai

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest -alias-types -generate types,client,spec -response-type-suffix Msg -o gen/openai.gen.go -package gen https://raw.githubusercontent.com/openai/openai-openapi/refs/heads/master/openapi.yaml
