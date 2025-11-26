
# Service Configuration
service_name = "main-service"
min_instances = 0
max_instances = 5
cpu_limit     = "1000m"
memory_limit  = "512Mi"

# Pub/Sub Configuration
pubsub_topics = [
  "user-events-dev",
  "order-events-dev",
  "notification-events-dev"
]

# Cloud Run Configuration
timeout_seconds           = 300
max_concurrent_requests   = 80
cpu_throttling           = true
startup_cpu_boost        = false

# Logging and Monitoring
log_level = "debug"

# Environment Variables (non-sensitive)
env_vars = {
  ENV                = "development"
  LOG_LEVEL         = "debug"
  ENABLE_PROFILING  = "true"
  MAX_WORKERS       = "10"
}

# Tags
labels = {
  environment = "dev"
  managed_by  = "terraform"
  team        = "backend"
}

# Allow unauthenticated access (for development only)
allow_unauthenticated = true