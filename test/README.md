# Testing

## Requirements

Make sure you have `opentofu` installed:

```bash
brew update
brew install opentofu
```

Ensure you have `awscli` installed and configured with a proper AWS profile and region:

```bash
# install aws cli
brew install awscli

# sso login
aws sso login --profile SystemAdministrator-***

export AWS_DEFAULT_PROFILE=SystemAdministrator-****
export AWS_REGION=eu-central-1
```

If you want to specify a non-random cluster UID:

```bash
export TESTS_CLUSTER_ID="myTest"
```

If you don't want to delete the resources at the end of the test:

```bash
export CLEAN_CLUSTER_AT_THE_END=false
```

Test with:

```bash
make test

# or just test one case
go test -v -timeout 120m -run TestDefaultEKSTestSuite
```

When you run the test, terratest will create a copy of the module to be tested in the `tests/states` directory. You can later navigate to the directory and use its content to manipulate the cluster. You can set the `SKIP_XXX` variable to prevent unique IDs of tests from being generated each time, thus using the same resources instead of deploying new resources with terraform.

## Troubleshooting

Ensure you don't have test clusters running for a while:

```bash
eksctl get clusters
```

You can change the default deployment region:

```bash
export TESTS_CLUSTER_REGION="eu-west-1"
```
