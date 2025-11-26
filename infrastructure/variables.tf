variable "environment" {
    description = "Environment name (prod, dev)"
    type        = string

    validation {
        condition     = contains(["prod", "dev"], var.environment)
        error_message = "Environment must be either 'prod' or 'dev'."
    }
}


# --------------------------------
# Scaling Configuration
# --------------------------------
variable "min_instances" {
  description = "Minimum number of instances for scaling"
  type        = number
  default     = 1
}

variable "max_instances" {
  description = "Maximum number of instances for scaling"
  type        = number
  default     = 1
}

# --------------------------------
# Timeout Configuration
# --------------------------------
variable "read_timeout" {
  description = "Read timeout in seconds"
  type        = string
  default     = "15s"
}

variable "write_timeout" {
  description = "Write timeout in seconds"
  type        = string
  default     = "15s"
}

variable "idle_timeout" {
  description = "Idle timeout in seconds"
  type        = string
  default     = "60s"
}

# --------------------------------
# Resource Configuration
# --------------------------------
variable "cpu_limit" {
    description = "CPU limit for each instance"
    type        = string
    default     = "1"
}

variable "memory_limit" {
    description = "Memory limit for each instance"
    type        = string
    default     = "512Mi"
}

# --------------------------------
# Access Configuration
# --------------------------------
variable "allow_public_access" {
  description = "Allow public access to the service"
  type        = bool
  default     = true
}

# --------------------------------
# Logging Configuration
# --------------------------------
variable "log_levels" {
  description = "Log level (debug, info, warn, error)"
  type        = string
  default     = "info"

  validation {
    condition     = contains(["debug", "info", "warn", "error"], var.log_levels)
    error_message = "Log level must be one of 'debug', 'info', 'warn', or 'error'."
  }
}


variable "log_format" {
  description = "Log format (json, text)"
  type        = string
  default     = "json"

  validation {
    condition     = contains(["json", "text"], var.log_format)
    error_message = "Log format must be either 'json' or 'text'."
  }
}

variable "image_tag" {
  description = "Docker image tag to deploy"
  type        = string
  default     = "latest"
}


variable "tf_state_bucket" {
  description = "GCS bucket name for terraform state"
  type        = string
}

variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "region" {
  description = "GCP Region"
  type        = string
}

variable "artifact_registry_repo" {
  description = "Artifact Registry repository name"
  type        = string
}