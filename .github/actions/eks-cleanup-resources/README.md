# Delete EKS Resources

This GitHub Action automates the deletion of AWS resources using a shell script. It helps you manage and clean up modules of this repository as resources by specifying a target or deleting resources based on age criteria.

## Usage

To use this action, include it in your workflow file (e.g., `.github/workflows/delete-eks-resources.yml`):

```yaml
name: Delete EKS Resources

on:
  workflow_dispatch:

jobs:
  cleanup:
    runs-on: ubuntu-latest
    steps:
      - name: Delete EKS resources
        uses: camunda/camunda-tf-eks-module/eks-cleanup-resources@main
        with:
          tf-bucket: 'your-s3-bucket-name'
          tf-bucket-region: 'your-region'
          max-age-hours: 24
          target: 'all'
          temp-dir: './tmp/eks-cleanup/'
```

## Inputs

The action supports the following input parameters:

| Input Name         | Description                                                                               | Required | Default                    |
|--------------------|-------------------------------------------------------------------------------------------|----------|----------------------------|
| `tf-bucket`        | The S3 bucket containing the resources' state files.                                       | Yes      | -                        |
| `tf-bucket-region` | The region of the S3 bucket containing the resources state files. Falls back to `AWS_REGION` if not set. | No       | AWS_REGION                 |
| `max-age-hours`    | The maximum age (in hours) for resources to be deleted.                                    | No       | "20"                       |
| `target`           | Specifies an ID to destroy specific resources or "all" to destroy all resources.           | No      | "all"                      |
| `temp-dir`         | Temporary directory prefix used for storing resource data during processing.               | No       | "./tmp/eks-cleanup/"       |
