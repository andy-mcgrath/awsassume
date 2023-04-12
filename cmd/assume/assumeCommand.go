package assume

import (
	"context"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/knadh/koanf/v2"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"

	"github.com/mitchellh/cli"
)

type Command struct {
	Ui     cli.Ui
	Action string
	Config *koanf.Koanf
}

func (c *Command) Run(args []string) int {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	cmdFlags := flag.NewFlagSet("assume", flag.ContinueOnError)
	cmdFlags.Usage = func() { c.Ui.Output(c.Help()) }

	var account, region, role string
	var export bool
	cmdFlags.BoolVar(&export, "export", false, "Output export commands instead of setting environment variables")
	cmdFlags.StringVar(&account, "account", "", "AWS account name or number")
	cmdFlags.StringVar(&region, "region", "eu-west-1", "AWS region default 'eu-west-1'")
	cmdFlags.StringVar(&role, "role", "", "AWS role to assume")

	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	if account == "" || role == "" {
		c.Ui.Error("Both account and role are required")
		return 1
	}

	roleArn := fmt.Sprintf("arn:aws:iam::%s:role/%s", account, role)
	c.Ui.Info(fmt.Sprintf("Assuming role: %s", roleArn))

	credentials, err := assumeRole(ctx, roleArn, region)
	if err != nil {
		errStr := fmt.Sprintf("Error assuming role: %v\n", err.Error())
		c.Ui.Error(errStr)
		return 1
	}

	switch export {
	case false:
		os.Setenv("AWS_ACCESS_KEY_ID", *credentials.AccessKeyId)
		os.Setenv("AWS_SECRET_ACCESS_KEY", *credentials.SecretAccessKey)
		os.Setenv("AWS_SESSION_TOKEN", *credentials.SessionToken)
		c.Ui.Info("Environment variables set successfully")
	case true:
		exportCommands := strings.Join([]string{
			fmt.Sprintf("export AWS_ACCESS_KEY_ID=\"%s\"", *credentials.AccessKeyId),
			fmt.Sprintf("export AWS_SECRET_ACCESS_KEY=\"%s\"", *credentials.SecretAccessKey),
			fmt.Sprintf("export AWS_SESSION_TOKEN=\"%s\"", *credentials.SessionToken),
		}, "\n")

		c.Ui.Output(exportCommands)
	}

	return 0
}

func (c *Command) Help() string {
	return `Usage: awsassume assume [options]

  Assume an AWS role and either set environment variables or output them as export commands.

Options:
  -account string
        AWS account name or number
  -role string
        AWS role to assume
  -region string
		AWS region e.g. 'us-east-1' (default "eu-west-1")
  -export bool
        Will output a set of export commands if set, if not assumed role will be applied via environment variables`
}

func (c *Command) Synopsis() string {
	return "Assume an AWS role and set or output temporary credentials"
}

func assumeRole(ctx context.Context, roleArn string, region string) (*types.Credentials, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	client := sts.NewFromConfig(cfg)

	assumedRole, err := client.AssumeRole(ctx, &sts.AssumeRoleInput{
		RoleArn:         aws.String(roleArn),
		RoleSessionName: aws.String("awsassume"),
	})
	if err != nil {
		return nil, fmt.Errorf("assuming role: %w", err)
	}

	return assumedRole.Credentials, nil
}
