#!/bin/bash
trivy config --config .lint/trivy/trivy.yaml --ignorefile .trivyignore .
