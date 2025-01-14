#!/bin/bash
cd "$(git rev-parse --show-toplevel)"
trivy config --config .lint/trivy/trivy.yaml --ignorefile .trivyignore .
