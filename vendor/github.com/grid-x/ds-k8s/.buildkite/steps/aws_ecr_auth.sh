#!/usr/bin/env bash

CREDS=$(aws sts assume-role --role-arn arn:aws:iam::485611583707:role/DevECRPowerUserRole --role-session-name buildkite-${BUILDKITE_BUILD_ID} --output text  | awk '/CREDENTIALS/ { print "AWS_ACCESS_KEY_ID="$2 " AWS_SECRET_ACCESS_KEY="$4 " AWS_SESSION_TOKEN="$5 }')
AWS_ECR_LOGIN="$CREDS aws ecr get-login --no-include-email"

$(eval "$AWS_ECR_LOGIN") 
