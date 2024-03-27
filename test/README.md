## Requirements

```bash
brew update
brew install opentofu
```[README.md](..%2FREADME.md)

Make sure you have awscli installed and configured
Make sure you have an AWS profile setup and a region:
```bash
# install aws cli
brew install awscli

# sso login
aws sso login --profile SystemAdministrator-***

export AWS_DEFAULT_PROFILE=SystemAdministrator-****
export AWS_REGION=eu-central-1
```

test with 
```bash
make test
```

### Troubleshooting

```bash
# make sure you don't have test clusters running since a while

eksctl get clusters 
```

# TODO: https://github.com/gruntwork-io/cloud-nuke every weekend
# => we should have a dedicated tenant for CI
# => sometimes, EKS deletion fails with error: DeleteCluster ResourceInUseException: Cluster has nodegroups attached terraform