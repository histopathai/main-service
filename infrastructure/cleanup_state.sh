#!/bin/bash
set -e

echo "Cleaning up orphaned IAM bindings from Terraform state..."

# Remove orphaned IAM bindings that reference non-existent topics
terraform state rm 'google_pubsub_topic_iam_member.main_service_publishers["dev-image-processing-request"]' 2>/dev/null || echo "  - dev-image-processing-request binding not in state"
terraform state rm 'google_pubsub_topic_iam_member.main_service_publishers["dev-image-processing-results"]' 2>/dev/null || echo "  - dev-image-processing-results binding not in state"
terraform state rm 'google_pubsub_topic_iam_member.main_service_publishers["dev-image-deletion-requests"]' 2>/dev/null || echo "  - dev-image-deletion-requests binding not in state"
terraform state rm 'google_pubsub_topic_iam_member.main_service_publishers["dev-image-process-dlq"]' 2>/dev/null || echo "  - dev-image-process-dlq binding not in state"

echo "Checking if Cloud Run service exists and needs to be imported..."

# Check if Cloud Run service is in state
if ! terraform state show google_cloud_run_v2_service.main_service >/dev/null 2>&1; then
  echo "Cloud Run service not in state, attempting import..."
  
  # Import the existing Cloud Run service
  terraform import \
    -var="image_tag=${TF_VAR_image_tag:-temp}" \
    -var="environment=${TF_VAR_environment:-dev}" \
    -var="tf_state_bucket=${TF_VAR_tf_state_bucket}" \
    -var="project_id=${TF_VAR_project_id}" \
    -var="region=${TF_VAR_region}" \
    -var="artifact_registry_repo=${TF_VAR_artifact_registry_repo}" \
    google_cloud_run_v2_service.main_service \
    "projects/${TF_VAR_project_id}/locations/${TF_VAR_region}/services/main-service-${TF_VAR_environment}" \
    2>&1 || echo "  - Import failed or service doesn't exist (will be created)"
else
  echo "  - Cloud Run service already in state"
fi

echo "State cleanup complete!"
