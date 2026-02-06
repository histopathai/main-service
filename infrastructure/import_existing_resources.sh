#!/bin/bash

# Script to import existing Pub/Sub resources into Terraform state
# This should be run once to handle the 409 errors from existing resources

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Starting import of existing Pub/Sub resources...${NC}"

# Get the project ID and environment from terraform variables
PROJECT_ID="${1:-$(terraform output -raw project_id 2>/dev/null || echo "")}"
ENVIRONMENT="${2:-dev}"

if [ -z "$PROJECT_ID" ]; then
    echo -e "${RED}Error: PROJECT_ID not provided and could not be determined from terraform output${NC}"
    echo "Usage: $0 <PROJECT_ID> [ENVIRONMENT]"
    exit 1
fi

echo -e "${GREEN}Project ID: $PROJECT_ID${NC}"
echo -e "${GREEN}Environment: $ENVIRONMENT${NC}"

# Set prefix based on environment
if [ "$ENVIRONMENT" == "dev" ]; then
    PREFIX="dev-"
else
    PREFIX=""
fi

# Function to import a resource if it doesn't exist in state
import_if_not_exists() {
    local resource_type=$1
    local resource_name=$2
    local resource_id=$3
    
    # Check if resource already exists in state
    if terraform state show "$resource_type.$resource_name" &>/dev/null; then
        echo -e "${YELLOW}Resource $resource_type.$resource_name already in state, skipping...${NC}"
    else
        echo -e "${GREEN}Importing $resource_type.$resource_name...${NC}"
        if terraform import "$resource_type.$resource_name" "$resource_id"; then
            echo -e "${GREEN}✓ Successfully imported $resource_type.$resource_name${NC}"
        else
            echo -e "${RED}✗ Failed to import $resource_type.$resource_name${NC}"
            # Don't exit, continue with other imports
        fi
    fi
}

# Import Pub/Sub Topics
echo -e "\n${YELLOW}Importing Pub/Sub Topics...${NC}"

import_if_not_exists "google_pubsub_topic" "upload_status" \
    "projects/$PROJECT_ID/topics/${PREFIX}upload-status"

import_if_not_exists "google_pubsub_topic" "image_processing_request" \
    "projects/$PROJECT_ID/topics/${PREFIX}image-processing-request"

import_if_not_exists "google_pubsub_topic" "image_processing_request_dlq" \
    "projects/$PROJECT_ID/topics/${PREFIX}image-processing-request-dlq"

import_if_not_exists "google_pubsub_topic" "image_processing_result" \
    "projects/$PROJECT_ID/topics/${PREFIX}image-processing-results"

import_if_not_exists "google_pubsub_topic" "image_processing_result_dlq" \
    "projects/$PROJECT_ID/topics/${PREFIX}image-processing-results-dlq"

import_if_not_exists "google_pubsub_topic" "image_deletion" \
    "projects/$PROJECT_ID/topics/${PREFIX}image-deletion-requests"

import_if_not_exists "google_pubsub_topic" "image_deletion_dlq" \
    "projects/$PROJECT_ID/topics/${PREFIX}image-deletion-requests-dlq"

import_if_not_exists "google_pubsub_topic" "image_process_dlq" \
    "projects/$PROJECT_ID/topics/${PREFIX}image-process-dlq"

# Import Pub/Sub Subscriptions
echo -e "\n${YELLOW}Importing Pub/Sub Subscriptions...${NC}"

import_if_not_exists "google_pubsub_subscription" "upload_status_sub" \
    "projects/$PROJECT_ID/subscriptions/${PREFIX}upload-status-sub"

import_if_not_exists "google_pubsub_subscription" "image_processing_request_sub" \
    "projects/$PROJECT_ID/subscriptions/${PREFIX}image-processing-request-sub"

import_if_not_exists "google_pubsub_subscription" "image_processing_result_sub" \
    "projects/$PROJECT_ID/subscriptions/${PREFIX}image-processing-results-sub"

import_if_not_exists "google_pubsub_subscription" "image_deletion_sub" \
    "projects/$PROJECT_ID/subscriptions/${PREFIX}image-deletion-requests-sub"

import_if_not_exists "google_pubsub_subscription" "image_process_dlq_sub" \
    "projects/$PROJECT_ID/subscriptions/${PREFIX}image-process-dlq-sub"

echo -e "\n${GREEN}Import process completed!${NC}"
echo -e "${YELLOW}Note: Some resources may have failed to import if they don't exist yet.${NC}"
echo -e "${YELLOW}Run 'terraform plan' to see what still needs to be created.${NC}"
