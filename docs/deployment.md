# Production Deployment Guide - 100 Concurrent Streams

## 🎯 Overview

This guide provides detailed server requirements, deployment strategies, and optimization recommendations for recording 100 concurrent HLS streams simultaneously.

## 📊 Resource Requirements Analysis

### **Hardware Requirements**

#### **Minimum Server Specifications**
```yaml
CPU: 16 cores (3.0GHz+)
RAM: 32GB
Storage: 2TB NVMe SSD
Network: 10Gbps connection
OS: Linux (Ubuntu 22.04 LTS recommended)
```

#### **Recommended Server Specifications**
```yaml
CPU: 32 cores (3.2GHz+) Intel Xeon or AMD EPYC
RAM: 64GB DDR4-3200
Storage: 4TB NVMe SSD (RAID 10)
Network: 25Gbps connection with redundancy
OS: Ubuntu 22.04 LTS or RHEL 9
```

#### **Cloud Instance Recommendations**

| Provider | Instance Type | vCPUs | RAM | Storage | Network | Cost/Month* |
|----------|---------------|-------|-----|---------|---------|-------------|
| **AWS** | c6i.8xlarge | 32 | 64GB | 4TB EBS gp3 | 25Gbps | ~$2,100 |
| **GCP** | c2-standard-30 | 30 | 120GB | 4TB SSD | 32Gbps | ~$1,900 |
| **Azure** | F32s_v2 | 32 | 64GB | 4TB Premium | 16Gbps | ~$1,800 |
| **Hetzner** | CCX53 | 32 | 128GB | 4TB NVMe | 10Gbps | ~$400 |

*Approximate pricing, varies by region and commitment

### **Detailed Resource Calculations**

#### **CPU Requirements**
```
Base Calculation:
- 100 streams × 5% CPU per stream = 500% CPU = 5 cores minimum
- With overhead and OS: 5 × 2.5 = 12.5 cores
- For safe headroom: 12.5 × 1.3 = 16 cores minimum

Recommended:
- 32 cores for comfort and burst handling
- High clock speed (3.0GHz+) for single-thread performance
- Modern architecture (Intel Ice Lake or AMD Zen 3+)
```

#### **Memory Requirements**
```
Per Stream Memory Usage:
- Application base: 50MB
- Per stream buffer: 10MB
- Database cache: 5MB per stream
- OS buffers: 5MB per stream
- Total per stream: 20MB

Calculation:
- Base: 50MB
- 100 streams: 100 × 20MB = 2GB
- OS overhead: 4GB
- File system cache: 8GB
- Safety buffer: 2x multiplier
- Total: (50MB + 2GB + 4GB + 8GB) × 2 = 28GB
- Recommended: 32GB minimum, 64GB optimal
```

#### **Storage Requirements**
```
Per Stream Storage (per hour):
- 720p stream: ~800MB/hour
- 1080p stream: ~1.5GB/hour
- 4K stream: ~6GB/hour

100 Streams (24-hour retention):
- Average quality (1080p): 100 × 1.5GB × 24h = 3.6TB/day
- Database overhead: ~500MB/day
- OS and application: ~100GB
- Total active storage needed: ~4TB

Weekly storage with rotation:
- 7 days × 3.6TB = 25TB
- Recommend 4TB with 24-hour retention
```

#### **Network Requirements**
```
Bandwidth per Stream:
- 720p: ~2.5 Mbps
- 1080p: ~5 Mbps
- 4K: ~20 Mbps

100 Concurrent Streams:
- Mixed quality average: ~5 Mbps per stream
- Total: 100 × 5 Mbps = 500 Mbps = 0.5 Gbps
- With overhead and peaks: 0.5 × 2.5 = 1.25 Gbps
- Recommended: 10 Gbps connection minimum
```

## 🏗️ Deployment Architecture

### **Single Server Deployment**

```yaml
# docker-compose.yml
version: '3.8'
services:
  vsr:
    build: .
    ports:
      - "8080:8080"
      - "9090:9090"  # Metrics
    volumes:
      - ./data:/app/files
      - ./db:/app/db
      - ./config:/app/config
    environment:
      - VSR_BIND_TO=:8080
      - VSR_METRICS_PORT=:9090
      - VSR_LOG_LEVEL=info
      - VSR_MAX_STREAMS=100
    restart: unless-stopped
    ulimits:
      nofile: 65536
    deploy:
      resources:
        limits:
          cpus: '30'
          memory: 60G
        reservations:
          cpus: '16'
          memory: 32G
```

### **Load Balanced Deployment**

```yaml
# nginx load balancer
upstream vsr_backend {
    least_conn;
    server vsr1:8080 max_fails=3 fail_timeout=30s;
    server vsr2:8080 max_fails=3 fail_timeout=30s;
    server vsr3:8080 max_fails=3 fail_timeout=30s;
}

server {
    listen 80;
    server_name vsr.example.com;
    
    location / {
        proxy_pass http://vsr_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_buffering off;
        proxy_cache off;
    }
    
    location /health {
        access_log off;
        proxy_pass http://vsr_backend;
    }
}
```

## ⚙️ System Optimization

### **Linux Kernel Tuning**

```bash
# /etc/sysctl.conf
# Network optimization
net.core.rmem_max = 536870912
net.core.wmem_max = 536870912
net.ipv4.tcp_rmem = 4096 65536 536870912
net.ipv4.tcp_wmem = 4096 65536 536870912
net.core.netdev_max_backlog = 10000
net.ipv4.tcp_congestion_control = bbr

# File system optimization
fs.file-max = 1000000
fs.inotify.max_user_watches = 524288
vm.swappiness = 10
vm.dirty_ratio = 15
vm.dirty_background_ratio = 5

# Network connections
net.ipv4.ip_local_port_range = 1024 65535
net.ipv4.tcp_max_syn_backlog = 8192
net.core.somaxconn = 8192
```

### **Application-Level Optimization**

```go
// Optimized configuration
const (
    MaxStreams = 100
    WorkerPoolSize = 200
    HTTPTimeout = 30 * time.Second
    SegmentBufferSize = 2 * 1024 * 1024 // 2MB
    DatabaseBatchSize = 100
    CleanupInterval = 5 * time.Minute
)

// Connection pool optimization
var httpClient = &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        1000,
        MaxIdleConnsPerHost: 100,
        IdleConnTimeout:     90 * time.Second,
        TLSHandshakeTimeout: 10 * time.Second,
        ResponseHeaderTimeout: 10 * time.Second,
        ExpectContinueTimeout: 1 * time.Second,
        MaxConnsPerHost:     100,
    },
    Timeout: HTTPTimeout,
}

// Worker pool for concurrent processing
type WorkerPool struct {
    workers   int
    jobQueue  chan Job
    quit      chan bool
    wg        sync.WaitGroup
}

func NewWorkerPool(workers int) *WorkerPool {
    return &WorkerPool{
        workers:  workers,
        jobQueue: make(chan Job, workers*2),
        quit:     make(chan bool),
    }
}
```

### **Database Optimization**

```go
// BoltDB optimization
func optimizedBoltDB() (*bolt.DB, error) {
    db, err := bolt.Open("stream.db", 0600, &bolt.Options{
        Timeout:         1 * time.Second,
        NoGrowSync:      false,
        NoFreelistSync:  false,
        FreelistType:    bolt.FreelistMapType,
        ReadOnly:        false,
        MmapFlags:       0,
        InitialMmapSize: 100 * 1024 * 1024, // 100MB
    })
    
    if err != nil {
        return nil, err
    }
    
    // Set batch size for bulk operations
    db.MaxBatchSize = 1000
    db.MaxBatchDelay = 10 * time.Millisecond
    
    return db, nil
}
```

## 📈 Monitoring and Alerting

### **Key Metrics to Monitor**

```go
// Prometheus metrics
var (
    activeStreams = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "vsr_active_streams",
        Help: "Number of currently active streams",
    })
    
    segmentsProcessed = prometheus.NewCounterVec(prometheus.CounterOpts{
        Name: "vsr_segments_processed_total",
        Help: "Total number of segments processed",
    }, []string{"stream_id", "quality"})
    
    errorRate = prometheus.NewCounterVec(prometheus.CounterOpts{
        Name: "vsr_errors_total",
        Help: "Total number of errors by type",
    }, []string{"type", "stream_id"})
    
    diskUsage = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "vsr_disk_usage_bytes",
        Help: "Current disk usage in bytes",
    })
    
    memoryUsage = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "vsr_memory_usage_bytes",
        Help: "Current memory usage in bytes",
    })
    
    networkThroughput = prometheus.NewGaugeVec(prometheus.GaugeOpts{
        Name: "vsr_network_throughput_bps",
        Help: "Network throughput in bits per second",
    }, []string{"direction"}) // "in" or "out"
)
```

### **Health Check Endpoints**

```go
// Health check implementation
func healthCheck(w http.ResponseWriter, r *http.Request) {
    health := &HealthStatus{
        Status:    "healthy",
        Timestamp: time.Now(),
        Checks:    make(map[string]interface{}),
    }
    
    // Database health
    if err := checkDatabase(); err != nil {
        health.Status = "unhealthy"
        health.Checks["database"] = map[string]interface{}{
            "status": "error",
            "error":  err.Error(),
        }
    } else {
        health.Checks["database"] = map[string]interface{}{
            "status": "ok",
        }
    }
    
    // Disk space check
    diskFree, diskTotal := checkDiskSpace()
    diskUsagePercent := float64(diskTotal-diskFree) / float64(diskTotal) * 100
    
    if diskUsagePercent > 90 {
        health.Status = "unhealthy"
    }
    
    health.Checks["disk"] = map[string]interface{}{
        "usage_percent": diskUsagePercent,
        "free_bytes":    diskFree,
        "total_bytes":   diskTotal,
    }
    
    // Memory check
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    health.Checks["memory"] = map[string]interface{}{
        "alloc_bytes":     m.Alloc,
        "total_alloc":     m.TotalAlloc,
        "sys_bytes":       m.Sys,
        "num_gc":          m.NumGC,
        "goroutines":      runtime.NumGoroutine(),
    }
    
    // Active streams
    health.Checks["streams"] = map[string]interface{}{
        "active_count": getActiveStreamCount(),
        "max_streams":  MaxStreams,
    }
    
    if health.Status == "unhealthy" {
        w.WriteHeader(http.StatusServiceUnavailable)
    }
    
    json.NewEncoder(w).Encode(health)
}
```

### **Alerting Rules**

```yaml
# Prometheus alerting rules
groups:
- name: vsr_alerts
  rules:
  - alert: HighErrorRate
    expr: rate(vsr_errors_total[5m]) > 0.1
    for: 2m
    labels:
      severity: warning
    annotations:
      summary: "High error rate detected"
      description: "Error rate is {{ $value }} errors/sec"
      
  - alert: DiskSpaceHigh
    expr: (vsr_disk_usage_bytes / disk_total_bytes) > 0.9
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: "Disk space critically low"
      description: "Disk usage is {{ $value }}%"
      
  - alert: MemoryUsageHigh
    expr: vsr_memory_usage_bytes > 50 * 1024^3  # 50GB
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "High memory usage"
      description: "Memory usage is {{ $value | humanize }}B"
      
  - alert: StreamsDown
    expr: vsr_active_streams < 90  # Less than 90 out of 100
    for: 2m
    labels:
      severity: warning
    annotations:
      summary: "Stream count low"
      description: "Only {{ $value }} streams active"
```

## 🔧 Configuration Management

### **Production Configuration**

```yaml
# config/production.yaml
server:
  bind_address: "0.0.0.0:8080"
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 60s
  max_header_bytes: 1048576

streams:
  max_concurrent: 100
  segment_timeout: 30s
  retry_attempts: 3
  retry_delay: 5s
  buffer_size: 2097152  # 2MB

storage:
  data_directory: "/data/segments"
  database_path: "/data/vsr.db"
  retention_hours: 24
  cleanup_interval: 5m
  max_disk_usage_percent: 85

network:
  max_idle_conns: 1000
  max_idle_conns_per_host: 100
  idle_conn_timeout: 90s
  tls_handshake_timeout: 10s

logging:
  level: "info"
  format: "json"
  output: "/var/log/vsr/app.log"
  max_size: 100  # MB
  max_backups: 10
  max_age: 30  # days

metrics:
  enabled: true
  bind_address: "0.0.0.0:9090"
  path: "/metrics"

health:
  bind_address: "0.0.0.0:8081"
  path: "/health"
```

### **Environment Variables**

```bash
# Production environment variables
export VSR_ENV=production
export VSR_CONFIG_FILE=/etc/vsr/config.yaml
export VSR_LOG_LEVEL=info
export VSR_MAX_STREAMS=100
export VSR_DATA_DIR=/data/vsr
export VSR_DB_PATH=/data/vsr/vsr.db
export VSR_RETENTION_HOURS=24
export VSR_BIND_TO=0.0.0.0:8080
export VSR_METRICS_PORT=9090
```

## 🚀 Deployment Strategies

### **Blue-Green Deployment**

```bash
#!/bin/bash
# blue-green-deploy.sh

NEW_VERSION=$1
CURRENT_COLOR=$(curl -s http://localhost:8080/health | jq -r '.version.color // "blue"')

if [ "$CURRENT_COLOR" = "blue" ]; then
    NEW_COLOR="green"
    NEW_PORT="8081"
    OLD_PORT="8080"
else
    NEW_COLOR="blue"
    NEW_PORT="8080"
    OLD_PORT="8081"
fi

echo "Deploying $NEW_VERSION to $NEW_COLOR environment on port $NEW_PORT"

# Start new version
docker run -d \
    --name vsr-$NEW_COLOR \
    --port $NEW_PORT:8080 \
    -e VSR_VERSION_COLOR=$NEW_COLOR \
    vsr:$NEW_VERSION

# Health check
for i in {1..30}; do
    if curl -f http://localhost:$NEW_PORT/health; then
        echo "Health check passed"
        break
    fi
    sleep 10
done

# Update load balancer
nginx -s reload

# Stop old version
docker stop vsr-$CURRENT_COLOR
docker rm vsr-$CURRENT_COLOR
```

### **Rolling Deployment (Multi-Instance)**

```yaml
# kubernetes deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: vsr-deployment
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: vsr
  template:
    metadata:
      labels:
        app: vsr
    spec:
      containers:
      - name: vsr
        image: vsr:latest
        ports:
        - containerPort: 8080
        resources:
          requests:
            cpu: 5000m
            memory: 10Gi
          limits:
            cpu: 10000m
            memory: 20Gi
        env:
        - name: VSR_MAX_STREAMS
          value: "33"  # 100 / 3 instances
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

## 💰 Cost Optimization

### **Cloud Cost Analysis (Monthly)**

| Component | AWS | GCP | Azure | Hetzner |
|-----------|-----|-----|-------|---------|
| **Compute** | $1,500 | $1,400 | $1,300 | $300 |
| **Storage** | $400 | $320 | $350 | $80 |
| **Network** | $200 | $150 | $180 | $50 |
| **Total** | $2,100 | $1,870 | $1,830 | $430 |

### **Cost Optimization Strategies**

1. **Reserved Instances**: 30-50% savings with 1-3 year commitments
2. **Spot Instances**: 60-90% savings for fault-tolerant workloads
3. **Storage Optimization**: Use lifecycle policies for old segments
4. **Network Optimization**: CDN for popular content delivery
5. **Auto-scaling**: Scale down during low usage periods

### **Storage Lifecycle Management**

```yaml
# AWS S3 lifecycle policy
{
  "Rules": [{
    "Id": "VSRLifecycle",
    "Status": "Enabled",
    "Transitions": [
      {
        "Days": 1,
        "StorageClass": "STANDARD_IA"
      },
      {
        "Days": 7,
        "StorageClass": "GLACIER"
      },
      {
        "Days": 30,
        "StorageClass": "DEEP_ARCHIVE"
      }
    ],
    "Expiration": {
      "Days": 365
    }
  }]
}
```

## 🔒 Security Considerations

### **Network Security**

```bash
# Firewall rules (iptables)
# Allow HTTP traffic
iptables -A INPUT -p tcp --dport 8080 -j ACCEPT

# Allow metrics (restricted)
iptables -A INPUT -p tcp --dport 9090 -s 10.0.0.0/8 -j ACCEPT

# Allow SSH (restricted)
iptables -A INPUT -p tcp --dport 22 -s YOUR_IP -j ACCEPT

# Block everything else
iptables -A INPUT -j DROP
```

### **Application Security**

```go
// Rate limiting
func rateLimitMiddleware(next http.Handler) http.Handler {
    limiter := rate.NewLimiter(100, 200) // 100 req/sec, burst 200
    
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !limiter.Allow() {
            http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
            return
        }
        next.ServeHTTP(w, r)
    })
}

// Input validation
func validateStreamURL(url string) error {
    parsed, err := url.Parse(url)
    if err != nil {
        return err
    }
    
    if parsed.Scheme != "http" && parsed.Scheme != "https" {
        return errors.New("invalid scheme")
    }
    
    if parsed.Host == "" {
        return errors.New("missing host")
    }
    
    return nil
}
```

## 📋 Deployment Checklist

### **Pre-Deployment**
- [ ] Hardware/instance provisioned and configured
- [ ] Operating system updated and hardened
- [ ] Network configuration completed
- [ ] Storage mounted and configured
- [ ] Monitoring and logging setup
- [ ] Security configurations applied
- [ ] Load balancer configured (if applicable)

### **Deployment**
- [ ] Application binary deployed
- [ ] Configuration files in place
- [ ] Database initialized
- [ ] Service started and enabled
- [ ] Health checks passing
- [ ] Monitoring alerts configured
- [ ] Backup procedures tested

### **Post-Deployment**
- [ ] Performance testing completed
- [ ] Load testing with 100 streams
- [ ] Failover procedures tested
- [ ] Documentation updated
- [ ] Team training completed
- [ ] Incident response procedures ready

## 🎯 Performance Benchmarks

### **Expected Performance (100 Streams)**
- **CPU Usage**: 60-80% under normal load
- **Memory Usage**: 40-50GB
- **Disk I/O**: 200-300 MB/s write, 100-150 MB/s read
- **Network**: 1-2 Gbps throughput
- **Latency**: <100ms response time for API calls
- **Throughput**: 1000+ concurrent HTTP connections

### **Scaling Thresholds**
- **Scale Up**: CPU >85% for 5 minutes
- **Scale Down**: CPU <30% for 15 minutes
- **Alert**: Memory >90% or disk >95%
- **Critical**: Error rate >5% or streams <50

This deployment guide provides a comprehensive foundation for running 100 concurrent streams reliably and efficiently in production.
