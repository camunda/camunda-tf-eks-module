#!/bin/bash

# see https://github.com/camunda/infra-core/tree/opensearch-cluster/camunda-opensearch#user-setup
psql -h $AURORA_ENDPOINT -p $AURORA_PORT "dbname=$AURORA_DB_NAME user=$AURORA_USERNAME password=$AURORA_PASSWORD" \
  -c "CREATE USER \"${AURORA_USERNAME_IRSA}\" WITH LOGIN;" \
  -c "GRANT rds_iam TO \"${AURORA_USERNAME_IRSA}\";" \
  -c "GRANT rds_superuser TO \"${AURORA_USERNAME_IRSA}\";" \
  -c "GRANT ALL PRIVILEGES ON DATABASE \"${AURORA_DB_NAME}\" TO \"${AURORA_USERNAME_IRSA}\";" \
  -c "SELECT aurora_version();" \
  -c "SELECT version();" -c "\du"
