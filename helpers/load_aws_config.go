package helpers

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"gopkg.in/yaml.v3"

	"github.com/yaninyzwitty/grpc-products-service/internal/pkg"
)

func FetchFromAWSConfig(ctx context.Context) (*pkg.Config, error) {
	// Load AWS default config (from env, profile, IAM role, etc.)
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %v", err)
	}

	// Create AWS SSM client
	ssmClient := ssm.NewFromConfig(cfg)

	// Fetch YAML config from SSM Parameter Store
	ssmParam, err := ssmClient.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           aws.String("/myapp/config.yaml"), // match your stored key
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch config from AWS SSM: %v", err)
	}

	configData := &pkg.Config{}
	err = yaml.Unmarshal([]byte(strings.TrimSpace(*ssmParam.Parameter.Value)), configData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %v", err)
	}

	// Fetch environment variables
	envParams, err := ssmClient.GetParameters(ctx, &ssm.GetParametersInput{
		Names:          []string{"/myapp/ASTRA_TOKEN", "/myapp/PULSAR_TOKEN"},
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch environment variables: %v", err)
	}

	// Set environment variables in process
	for _, p := range envParams.Parameters {
		switch *p.Name {
		case "/myapp/ASTRA_TOKEN":
			_ = os.Setenv("ASTRA_TOKEN", *p.Value)
		case "/myapp/PULSAR_TOKEN":
			_ = os.Setenv("PULSAR_TOKEN", *p.Value)
		}
	}

	return configData, nil
}
