# Configuration Guide

## 🔧 Command Line Options

### **Basic Usage**
```bash
./vsr --url "https://example.com/stream.m3u8" --tail 24 --bind-to ":8080"
```

### **All Available Flags**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--url` | string | *required* | HLS stream URL to record |
| `--tail` | int | 24 | Hours of content to retain |
| `--bind-to` | string | ":8080" | Server bind address |
| `--host` | string | "http://localhos:8080" | Host for M3U8 URLs |
| `--debug` | bool | false | Skip fetcher for testing |
| `--version` | bool | false | Show version information |

### **Examples**

#### **Basic Recording**
```bash
# Record a stream with 24-hour retention
./vsr --url "https://example.com/stream.m3u8"
```

#### **Custom Retention**
```bash
# Keep 48 hours of content
./vsr --url "https://example.com/stream.m3u8" --tail 48
```

#### **Custom Port**
```bash
# Run on port 9090
./vsr --url "https://example.com/stream.m3u8" --bind-to ":9090"
```

#### **Debug Mode**
```bash
# Test server without fetching streams
./vsr --url "https://example.com/stream.m3u8" --debug
```

## 🌍 Environment Variables

### **Configuration via Environment**
```bash
export VSR_URL="https://example.com/stream.m3u8"
export VSR_TAIL=24
export VSR_BIND_TO=":8080"
export VSR_HOST="http://localhost:8080"
export VSR_DEBUG=false
export VSR_DATA_DIR="./files"
export VSR_DB_PATH="./db.db"
```

### **Docker Environment Variables**
```yaml
# docker-compose.yml
environment:
  - VSR_URL=https://example.com/stream.m3u8
  - VSR_TAIL=24
  - VSR_BIND_TO=:8080
  - VSR_LOG_LEVEL=info
  - VSR_MAX_STREAMS=100
```

## 📁 File Structure Configuration

### **Default Directory Layout**
```
./
├── vsr                    # Binary executable
├── files/                 # Segment storage directory
│   ├── 1642678920000000000.ts
│   ├── 1642678922000000000.ts
│   └── ...
├── db.db                  # BoltDB database file
└── logs/                  # Log files (if configured)
    └── vsr.log
```

### **Custom Paths**
```bash
# Custom data directory
./vsr --url "..." --data-dir "/data/vsr/segments"

# Custom database path
./vsr --url "..." --db-path "/data/vsr/database.db"
```

## ⚙️ Advanced Configuration (Future Enhancements)

### **YAML Configuration File**

```yaml
# config/vsr.yaml
server:
  bind_address: "0.0.0.0:8080"
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 60s
  max_header_bytes: 1048576

streams:
  max_concurrent: 100
  default_retention_hours: 24
  segment_timeout: 30s
  retry_attempts: 3
  retry_delay: 5s
  buffer_size: 2097152  # 2MB

storage:
  data_directory: "./files"
  database_path: "./db.db"
  cleanup_interval: 5m
  max_disk_usage_percent: 85

network:
  max_idle_conns: 1000
  max_idle_conns_per_host: 100
  idle_conn_timeout: 90s
  tls_handshake_timeout: 10s
  response_header_timeout: 10s

logging:
  level: "info"           # debug, info, warn, error
  format: "text"          # text, json
  output: "stdout"        # stdout, stderr, file path
  max_size: 100           # MB
  max_backups: 10
  max_age: 30             # days
  compress: true

metrics:
  enabled: true
  bind_address: "0.0.0.0:9090"
  path: "/metrics"

health:
  enabled: true
  bind_address: "0.0.0.0:8081"
  path: "/health"

preprocessing:
  enabled: false
  ytdlp:
    binary_path: "yt-dlp"
    quality: "best"
    format: "best[protocol=m3u8]"
    timeout: 60s
  ffmpeg:
    binary_path: "ffmpeg"
    probe_timeout: 30s

postprocessing:
  enabled: false
  upload:
    provider: "s3"        # s3, gcs, azure
    bucket: "my-streams"
    region: "us-west-2"
    cleanup_after_upload: true
  stitching:
    enabled: false
    interval: 1h
    format: "mp4"
    quality: "1080p"
```

### **JSON Configuration**
```json
{
  "server": {
    "bind_address": "0.0.0.0:8080",
    "read_timeout": "30s",
    "write_timeout": "30s"
  },
  "streams": {
    "max_concurrent": 100,
    "default_retention_hours": 24
  },
  "storage": {
    "data_directory": "./files",
    "database_path": "./db.db"
  },
  "logging": {
    "level": "info",
    "format": "json"
  }
}
```

### **Configuration Loading Priority**
1. Command line flags (highest priority)
2. Environment variables
3. Configuration file
4. Default values (lowest priority)

```go
// Example configuration loading
func LoadConfig() *Config {
    config := &Config{
        // Default values
        Server: ServerConfig{
            BindAddress: ":8080",
            ReadTimeout: 30 * time.Second,
        },
    }
    
    // Load from config file
    if err := loadConfigFile(config); err != nil {
        log.Printf("Config file error: %v", err)
    }
    
    // Override with environment variables
    loadFromEnv(config)
    
    // Override with command line flags
    loadFromFlags(config)
    
    return config
}
```

## 🔐 Security Configuration

### **TLS/HTTPS Configuration**
```yaml
server:
  tls:
    enabled: true
    cert_file: "/etc/ssl/certs/vsr.crt"
    key_file: "/etc/ssl/private/vsr.key"
    min_version: "1.2"
    cipher_suites:
      - "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"
      - "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305"
```

### **Authentication Configuration**
```yaml
auth:
  enabled: true
  type: "jwt"              # jwt, api_key, basic
  jwt:
    secret: "your-secret-key"
    expiration: 24h
    issuer: "vsr-server"
  api_key:
    keys:
      - "api-key-1"
      - "api-key-2"
  rate_limiting:
    enabled: true
    requests_per_minute: 100
    burst_size: 10
```

### **CORS Configuration**
```yaml
cors:
  enabled: true
  allowed_origins:
    - "https://example.com"
    - "https://app.example.com"
  allowed_methods:
    - "GET"
    - "POST"
    - "PUT"
    - "DELETE"
  allowed_headers:
    - "Authorization"
    - "Content-Type"
  max_age: 3600
```

## 📊 Monitoring Configuration

### **Prometheus Metrics**
```yaml
metrics:
  enabled: true
  bind_address: "0.0.0.0:9090"
  path: "/metrics"
  namespace: "vsr"
  subsystem: "recorder"
  custom_metrics:
    - name: "stream_quality"
      help: "Stream quality metrics"
      type: "histogram"
      buckets: [0.1, 0.5, 1.0, 2.5, 5.0, 10.0]
```

### **Health Check Configuration**
```yaml
health:
  enabled: true
  bind_address: "0.0.0.0:8081"
  path: "/health"
  checks:
    database:
      enabled: true
      timeout: 5s
    disk_space:
      enabled: true
      threshold_percent: 90
    memory:
      enabled: true
      threshold_percent: 85
    active_streams:
      enabled: true
      min_streams: 1
```

### **Logging Configuration**
```yaml
logging:
  level: "info"
  format: "json"
  output: "/var/log/vsr/app.log"
  rotation:
    max_size: 100    # MB
    max_backups: 10
    max_age: 30      # days
    compress: true
  fields:
    service: "vsr"
    version: "1.0.0"
    environment: "production"
```

## 🌐 Network Configuration

### **HTTP Server Settings**
```yaml
server:
  bind_address: "0.0.0.0:8080"
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 60s
  max_header_bytes: 1048576
  keep_alive: true
  
  # Request limits
  max_concurrent_requests: 1000
  max_request_size: 10485760  # 10MB
  
  # Compression
  compression:
    enabled: true
    level: 6
    types:
      - "application/vnd.apple.mpegurl"
      - "application/json"
      - "text/plain"
```

### **HTTP Client Settings**
```yaml
network:
  client:
    timeout: 30s
    max_idle_conns: 1000
    max_idle_conns_per_host: 100
    idle_conn_timeout: 90s
    tls_handshake_timeout: 10s
    response_header_timeout: 10s
    expect_continue_timeout: 1s
    max_conns_per_host: 100
    
  # Retry configuration
  retry:
    attempts: 3
    delay: 5s
    backoff: "exponential"  # linear, exponential
    max_delay: 60s
```

## 💾 Storage Configuration

### **Database Settings**
```yaml
storage:
  database:
    path: "./db.db"
    timeout: 1s
    no_grow_sync: false
    no_freelist_sync: false
    freelist_type: "map"    # array, map
    read_only: false
    initial_mmap_size: 104857600  # 100MB
    max_batch_size: 1000
    max_batch_delay: 10ms
```

### **File System Settings**
```yaml
storage:
  filesystem:
    data_directory: "./files"
    permissions: 0755
    sync_writes: true
    temp_directory: "/tmp/vsr"
    
  # Cleanup settings
  cleanup:
    interval: 5m
    batch_size: 100
    retention_hours: 24
    max_disk_usage_percent: 85
    emergency_cleanup_percent: 95
```

### **Cloud Storage Settings**
```yaml
storage:
  cloud:
    enabled: false
    provider: "s3"
    config:
      bucket: "my-vsr-bucket"
      region: "us-west-2"
      access_key_id: "${AWS_ACCESS_KEY_ID}"
      secret_access_key: "${AWS_SECRET_ACCESS_KEY}"
      endpoint: ""  # Custom S3 endpoint
      force_path_style: false
    
    # Upload settings
    upload:
      workers: 10
      timeout: 5m
      retry_attempts: 3
      chunk_size: 5242880  # 5MB
```

## 🔄 Stream Configuration

### **HLS Settings**
```yaml
streams:
  hls:
    segment_duration: 2s
    playlist_length: 5      # Number of segments in live playlist
    target_duration: 10s
    version: 4
    
  # Fetching settings
  fetching:
    interval: 3s
    timeout: 30s
    user_agent: "VSR/1.0"
    follow_redirects: true
    max_redirects: 5
    
  # Quality settings
  quality:
    preferred: "1080p"
    fallback: "720p"
    min_bitrate: 1000000    # 1 Mbps
    max_bitrate: 10000000   # 10 Mbps
```

### **Multi-Stream Configuration**
```yaml
streams:
  multiple:
    enabled: true
    max_concurrent: 100
    per_stream_memory_limit: 10485760  # 10MB
    load_balancing: "round_robin"      # round_robin, least_loaded
    
  # Stream groups
  groups:
    high_priority:
      max_streams: 50
      memory_limit: 524288000  # 500MB
      retention_hours: 48
      
    low_priority:
      max_streams: 50
      memory_limit: 262144000  # 250MB
      retention_hours: 24
```

## 🚀 Performance Tuning

### **Memory Configuration**
```yaml
performance:
  memory:
    gc_percent: 100
    max_heap_size: 8589934592  # 8GB
    buffer_pool_size: 1000
    segment_buffer_size: 2097152  # 2MB
    
  # Goroutine settings
  goroutines:
    max_workers: 1000
    worker_queue_size: 10000
    idle_timeout: 5m
```

### **CPU Configuration**
```yaml
performance:
  cpu:
    max_procs: 0  # Use all available CPUs
    gomaxprocs: -1
    
  # Processing settings
  processing:
    batch_size: 100
    flush_interval: 1s
    compression_level: 6
```

## 📝 Configuration Validation

### **Validation Rules**
```go
type ConfigValidator struct {
    rules []ValidationRule
}

type ValidationRule struct {
    Field    string
    Required bool
    Min      interface{}
    Max      interface{}
    Pattern  *regexp.Regexp
}

func ValidateConfig(config *Config) error {
    validator := &ConfigValidator{
        rules: []ValidationRule{
            {Field: "server.bind_address", Required: true},
            {Field: "streams.max_concurrent", Min: 1, Max: 1000},
            {Field: "storage.retention_hours", Min: 1, Max: 8760},
        },
    }
    
    return validator.Validate(config)
}
```

### **Configuration Examples by Use Case**

#### **Development Environment**
```yaml
# config/development.yaml
server:
  bind_address: "localhost:8080"
  
streams:
  max_concurrent: 5
  default_retention_hours: 2

storage:
  data_directory: "./dev-data"
  
logging:
  level: "debug"
  format: "text"
  output: "stdout"

metrics:
  enabled: false
```

#### **Testing Environment**
```yaml
# config/testing.yaml
server:
  bind_address: "0.0.0.0:8080"
  
streams:
  max_concurrent: 10
  default_retention_hours: 1

storage:
  data_directory: "/tmp/vsr-test"
  
logging:
  level: "info"
  format: "json"
  
health:
  enabled: true
```

#### **Production Environment**
```yaml
# config/production.yaml
server:
  bind_address: "0.0.0.0:8080"
  tls:
    enabled: true
    cert_file: "/etc/ssl/certs/vsr.crt"
    key_file: "/etc/ssl/private/vsr.key"
  
streams:
  max_concurrent: 100
  default_retention_hours: 24

storage:
  data_directory: "/data/vsr/segments"
  database_path: "/data/vsr/db.db"
  cloud:
    enabled: true
    provider: "s3"
  
logging:
  level: "info"
  format: "json"
  output: "/var/log/vsr/app.log"
  
metrics:
  enabled: true
  
health:
  enabled: true
  
auth:
  enabled: true
  type: "jwt"
```

This configuration guide provides comprehensive options for customizing VSR behavior across different environments and use cases.
