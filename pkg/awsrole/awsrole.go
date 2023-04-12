package awsrole

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// stsClientAPI is the interface for the sts client, allowing for mocking during testing
type stsClientAPI interface {
	AssumeRole(ctx context.Context, in *sts.AssumeRoleInput, opts ...func(*sts.Options)) (*sts.AssumeRoleOutput, error)
}

// AwsRole is a struct that holds the AWS role information, allows role assumption and returns credentials.
//   - Implements the aws.CredentialsProvider interface.
type AwsRole struct {
	// The AWS account ID
	account string

	// The AWS role name
	role string

	// The aws.Config struct to use for the assumed role
	cfg *aws.Config

	// The assumed role output, set by the AssumeRole function and used
	// by the Retrieve function to return the assumed role credentials
	assumeRoleOutput *sts.AssumeRoleOutput

	// Function is used return a new sts client, allowing for mocking during testing
	//  Defaults to call sts.NewFromConfig with the provided
	//  aws.Config in the NewAwsRole function
	stsNewFromConfig func(cfg aws.Config) stsClientAPI
}

// NewAwsRole returns a new AwsRole struct
func NewAwsRole(cfg aws.Config) *AwsRole {
	return &AwsRole{
		cfg: &cfg,
		stsNewFromConfig: func(cfg aws.Config) stsClientAPI {
			return sts.NewFromConfig(cfg)
		},
	}
}

// UpdateAssumedRole updates the assumed role and returns the credentials
func (a *AwsRole) UpdateAssumedRole(ctx context.Context) (*aws.Config, error) {
	return a.AssumeRole(ctx, a.account, a.role, a.cfg.Region)
}

// AssumeRole assumes the role and sets the credentials in the AwsRole struct
func (a *AwsRole) AssumeRole(ctx context.Context, account, role, region string) (*aws.Config, error) {
	roleArn := fmt.Sprintf("arn:aws:iam::%s:role/%s", account, role)
	a.role = role
	a.account = account

	a.cfg.Region = region

	stsClient := a.stsNewFromConfig(*a.cfg)

	assumeRoleInput := &sts.AssumeRoleInput{
		RoleArn:         aws.String(roleArn),
		RoleSessionName: aws.String(fmt.Sprintf("%s-%s", account, role)),
	}

	assumeRoleOutput, err := stsClient.AssumeRole(ctx, assumeRoleInput)
	if err != nil {
		return nil, fmt.Errorf("assuming role: %w", err)
	}

	a.assumeRoleOutput = assumeRoleOutput

	return a.GetAssumedConfig(), nil
}

// GetAssumedConfig returns a new aws.Config with the assumed role credentials
func (a *AwsRole) GetAssumedConfig() *aws.Config {
	newCfg := a.cfg.Copy()
	newCfg.Credentials = a
	return &newCfg
}

// Retrieve returns the assumed role credentials aws.Credentials
//   - Required for the aws.CredentialsProvider interface.
func (a *AwsRole) Retrieve(ctx context.Context) (aws.Credentials, error) {
	credentials := a.assumeRoleOutput.Credentials
	if credentials == nil {
		return aws.Credentials{}, fmt.Errorf("assume role response missing credentials")
	}
	return aws.Credentials{
		AccessKeyID:     *credentials.AccessKeyId,
		SecretAccessKey: *credentials.SecretAccessKey,
		SessionToken:    *credentials.SessionToken,
		Source:          "AssumeRoleProvider",
	}, nil
}
