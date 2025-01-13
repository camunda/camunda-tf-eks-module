# Since Trivy does not have a pre-commit hook by default, this is a custom hook script
#!/bin/bash
trivy config --config .lint/trivy/trivy.yaml --ignorefile .trivyignore
