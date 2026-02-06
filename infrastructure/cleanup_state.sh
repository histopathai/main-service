#!/bin/bash
set -e

echo "Cleaning up orphaned IAM bindings from Terraform state..."

# Remove orphaned IAM bindings that reference non-existent topics
terraform state rm 'google_pubsub_topic_iam_member.main_service_publishers["dev-image-processing-request"]' 2>/dev/null || echo "  - dev-image-processing-request binding not in state"
terraform state rm 'google_pubsub_topic_iam_member.main_service_publishers["dev-image-processing-results"]' 2>/dev/null || echo "  - dev-image-processing-results binding not in state"
terraform state rm 'google_pubsub_topic_iam_member.main_service_publishers["dev-image-deletion-requests"]' 2>/dev/null || echo "  - dev-image-deletion-requests binding not in state"
terraform state rm 'google_pubsub_topic_iam_member.main_service_publishers["dev-image-process-dlq"]' 2>/dev/null || echo "  - dev-image-process-dlq binding not in state"

echo "State cleanup complete!"
