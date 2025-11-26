
environment     = "prod"

# ------------------------------
# Scaling Configuration
# ------------------------------
min_instances = 1
max_instances = 2

# ------------------------------
# Resource Configuration
# ------------------------------
cpu_limit       = "1"
memory_limit    = "512Mi"

# ------------------------------
# Access Configuration
# ------------------------------
allow_public_access = false

# ------------------------------
# Logging Configuration
# ------------------------------
log_level = "info"
 

# ------------------------------
# Timeout Configuration
# ------------------------------
read_timeout  = "15s"
write_timeout = "60s"
idle_timeout  = "120s"