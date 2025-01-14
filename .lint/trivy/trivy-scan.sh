#!/bin/bash
trivy config --config .lint/trivy/trivy.yaml --ignorefile .trivyignore ./examples/camunda-8.6
trivy config --config .lint/trivy/trivy.yaml --ignorefile .trivyignore ./examples/camunda-8.6-irsa
trivy config --config .lint/trivy/trivy.yaml --ignorefile .trivyignore ./examples/camunda-8.7
trivy config --config .lint/trivy/trivy.yaml --ignorefile .trivyignore ./examples/camunda-8.7-irsa
trivy config --config .lint/trivy/trivy.yaml --ignorefile .trivyignore ./modules
