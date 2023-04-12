package awsrole

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
	"testing"
)

// mockSTSClient is a mock implementation of the STS client used for testing.
// Implements the stsClientAPI interface.
type mockSTSClient struct {
	AssumeRoleFn func(context.Context, *sts.AssumeRoleInput, ...func(*sts.Options)) (*sts.AssumeRoleOutput, error)
}

func (m *mockSTSClient) AssumeRole(ctx context.Context, in *sts.AssumeRoleInput, opts ...func(*sts.Options)) (*sts.AssumeRoleOutput, error) {
	return m.AssumeRoleFn(ctx, in, opts...)
}

func TestAwsRole_AssumeRole(t *testing.T) {
	cfg := aws.Config{
		Region: "us-west-2",
		//Logger: logging.DefaultEntry(),
	}
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		role := "test-role"
		account := "123456789012"

		awsRole := &AwsRole{
			cfg: &cfg,
			stsNewFromConfig: func(cfg aws.Config) stsClientAPI {
				return &mockSTSClient{
					AssumeRoleFn: func(ctx context.Context, in *sts.AssumeRoleInput, opts ...func(*sts.Options)) (*sts.AssumeRoleOutput, error) {
						return &sts.AssumeRoleOutput{
							Credentials: &types.Credentials{
								AccessKeyId:     aws.String("access_key"),
								SecretAccessKey: aws.String("secret_key"),
								SessionToken:    aws.String("session_token"),
							},
						}, nil
					},
				}
			},
		}

		newCfg, err := awsRole.AssumeRole(ctx, account, role, cfg.Region)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		assumedCredentials, err := awsRole.Retrieve(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if newCfg == nil {
			t.Fatal("newCfg should not be nil")
		}

		if assumedCredentials.AccessKeyID != "access_key" ||
			assumedCredentials.SecretAccessKey != "secret_key" ||
			assumedCredentials.SessionToken != "session_token" {
			t.Error("assumed credentials do not match expected values")
		}
	})

	t.Run("error", func(t *testing.T) {
		awsRole := &AwsRole{
			cfg: &cfg,
			stsNewFromConfig: func(cfg aws.Config) stsClientAPI {
				return &mockSTSClient{
					AssumeRoleFn: func(ctx context.Context, in *sts.AssumeRoleInput, opts ...func(*sts.Options)) (*sts.AssumeRoleOutput, error) {
						return nil, errors.New("assume role error")
					},
				}
			},
		}

		_, err := awsRole.AssumeRole(ctx, "123456789012", "test-role", cfg.Region)
		if err == nil {
			t.Fatal("expected an error but got none")
		}
	})
}

func TestAwsRole_GetAssumedConfig(t *testing.T) {
	// Test if GetAssumedConfig returns a non-nil aws.Config instance
	cfg := aws.Config{
		Region: "us-west-2",
	}
	ctx := context.Background()
	awsRole := NewAwsRole(cfg)

	// Set the assumeRoleOutput with mocked data
	awsRole.assumeRoleOutput = &sts.AssumeRoleOutput{
		Credentials: &types.Credentials{
			AccessKeyId:     aws.String("access_key"),
			SecretAccessKey: aws.String("secret_key"),
			SessionToken:    aws.String("session_token"),
		},
	}

	// Call GetAssumedConfig to get the new configuration
	newCfg := awsRole.GetAssumedConfig()

	if newCfg == nil {
		t.Fatal("newCfg should not be nil")
	}

	// Call Retrieve to get the assumed credentials
	assumedCredentials, err := awsRole.Retrieve(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check if the assumed credentials are set correctly
	if assumedCredentials.AccessKeyID != "access_key" ||
		assumedCredentials.SecretAccessKey != "secret_key" ||
		assumedCredentials.SessionToken != "session_token" {
		t.Error("assumed credentials do not match expected values")
	}
}
