# Production Post-Mortem Guide

## Overview

This document outlines potential production issues for the admoai-test advertisement management system and provides detailed response strategies. Use this as a reference during incidents and for proactive monitoring setup.

## Critical Production Issues

### 1. Database Issues

#### SQLite Database Corruption
**Symptoms:**
- Application crashes with SQLite errors
- GORM connection failures
- Data inconsistency or loss

**Root Causes:**
- Disk space exhaustion
- Improper shutdown during writes
- Concurrent access issues
- Hardware failures

**Immediate Response:**
```bash
# Check disk space
df -h

# Verify database integrity
sqlite3 ./sqliteData/database.db "PRAGMA integrity_check;"

# Check database size
ls -lh ./sqliteData/database.db
```

**Recovery Steps:**
1. Stop the application immediately
2. Backup corrupted database: `cp database.db database.db.corrupted`
3. Restore from latest backup
4. If no backup exists, attempt repair:
   ```bash
   sqlite3 database.db ".recover" > recovered.sql
   sqlite3 new_database.db < recovered.sql
   ```
5. Restart application with recovered database

**Prevention:**
- Implement regular database backups
- Monitor disk space with alerts at 80% capacity
- Use WAL mode for better concurrency
- Implement proper graceful shutdown

#### Database Connection Pool Exhaustion
**Symptoms:**
- HTTP 500 errors
- "too many connections" errors
- Application hanging on database operations

**Immediate Response:**
```bash
# Check current connections
docker exec -it admoai-test-app-1 /bin/sh
ps aux | grep admoai

# Restart application
docker-compose restart
```

**Long-term Fix:**
```go
// Add to initializeDB() function
sqlDB.SetMaxOpenConns(25)
sqlDB.SetMaxIdleConns(25)
sqlDB.SetConnMaxLifetime(5 * time.Minute)
```

### 2. Application Performance Issues

#### Memory Leaks
**Symptoms:**
- Increasing memory usage over time
- Application becoming unresponsive
- OOM kills in Docker logs

**Monitoring:**
```bash
# Check memory usage
docker stats admoai-test-app-1

# Monitor Go runtime metrics
curl http://localhost:8080/metrics | grep go_memstats
```

**Immediate Response:**
1. Restart the application: `docker-compose restart`
2. Check for memory leaks in logs
3. Monitor metrics endpoint for memory patterns

**Investigation:**
```bash
# Enable Go memory profiling
go tool pprof http://localhost:8080/debug/pprof/heap

# Check for goroutine leaks
curl http://localhost:8080/metrics | grep go_goroutines
```

#### High CPU Usage
**Symptoms:**
- Slow response times
- Application timeouts
- High CPU utilization in monitoring

**Immediate Response:**
```bash
# Check CPU usage
top -p $(docker inspect --format '{{.State.Pid}}' admoai-test-app-1)

# Check for CPU-intensive operations
docker logs admoai-test-app-1 | tail -100
```

**Common Causes:**
- Inefficient database queries
- Infinite loops in cron jobs
- High request volume without rate limiting

### 3. Docker Container Issues

#### Container Crashes
**Symptoms:**
- Application unreachable
- Container restart loops
- Error logs in Docker

**Immediate Response:**
```bash
# Check container status
docker ps -a

# View container logs
docker logs admoai-test-app-1 --tail 100

# Check resource usage
docker stats

# Restart container
docker-compose restart
```

**Common Causes:**
- Out of memory conditions
- Panic in Go application
- Disk space issues
- Port conflicts

#### Volume Mount Issues
**Symptoms:**
- Database not persisting
- Permission errors
- SQLite file not found

**Immediate Response:**
```bash
# Check volume mounts
docker inspect admoai-test-app-1 | grep -A 10 "Mounts"

# Verify permissions
ls -la ./data/
sudo chown -R 1000:1000 ./data/
```

### 4. Network and Connectivity Issues

#### Port Conflicts
**Symptoms:**
- Application fails to start
- "Port already in use" errors
- Connection refused errors

**Immediate Response:**
```bash
# Check what's using port 8080
sudo netstat -tulpn | grep :8080
sudo lsof -i :8080

# Kill conflicting process
sudo kill -9 <PID>

# Restart application
docker-compose up -d
```

#### Docker Socket Access Issues
**Symptoms:**
- Log download functionality failing
- "Permission denied" for Docker commands
- `/logs` endpoint returning errors

**Immediate Response:**
```bash
# Check Docker socket permissions
ls -la /var/run/docker.sock

# Fix permissions
sudo chmod 666 /var/run/docker.sock

# Restart container
docker-compose restart
```

### 5. Application Logic Issues

#### Cron Job Failures
**Symptoms:**
- Ads not expiring automatically
- Memory warnings not appearing
- Application logs showing cron errors

**Immediate Response:**
```bash
# Check cron job logs
docker logs admoai-test-app-1 | grep -i cron

# Manually trigger ad expiration
curl -X POST http://localhost:8080/admin/expire-ads
```

**Investigation:**
1. Check database connectivity in cron context
2. Verify cron syntax and timing
3. Look for panic conditions in cron functions

#### Metrics Collection Issues
**Symptoms:**
- `/metrics` endpoint returning errors
- Missing or incorrect metrics
- Prometheus scraping failures

**Immediate Response:**
```bash
# Test metrics endpoint
curl http://localhost:8080/metrics

# Check for data races
go run -race main.go
```

### 6. Data Consistency Issues

#### Concurrent Modification Issues
**Symptoms:**
- Ads in inconsistent states
- Data races in logs
- Unexpected ad status changes

**Immediate Response:**
1. Check for race conditions in logs
2. Verify database transaction integrity
3. Review concurrent access patterns

**Long-term Fix:**
```go
// Add proper locking to critical sections
var adMutex sync.RWMutex

func updateAdStatus(db *gorm.DB, id uint, status string) error {
    adMutex.Lock()
    defer adMutex.Unlock()
    
    return db.Model(&models.Ad{}).Where("id = ?", id).Update("status", status).Error
}
```

## Monitoring and Alerting Setup

### Key Metrics to Monitor

```bash
# Application Health
curl http://localhost:8080/ads?placement=homepage&status=active

# Memory Usage
curl http://localhost:8080/metrics | grep go_memstats_alloc_bytes

# Request Rate
curl http://localhost:8080/metrics | grep http_requests_total

# Database Connections
curl http://localhost:8080/metrics | grep go_goroutines
```

### Recommended Alerts

1. **Memory Usage > 80%**
   - Check for memory leaks
   - Consider scaling up

2. **Response Time > 5 seconds**
   - Check database performance
   - Review slow queries

3. **Error Rate > 5%**
   - Check application logs
   - Verify database connectivity

4. **Disk Usage > 90%**
   - Clean up old logs
   - Implement log rotation

5. **Active Ads > Threshold**
   - Review ad expiration logic
   - Check for stuck ads

## Backup and Recovery Procedures

### Daily Backup Script
```bash
#!/bin/bash
# backup-db.sh
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/backups"
mkdir -p $BACKUP_DIR

# Backup SQLite database
cp ./data/database.db $BACKUP_DIR/database_$DATE.db

# Compress backup
gzip $BACKUP_DIR/database_$DATE.db

# Keep only last 7 days
find $BACKUP_DIR -name "database_*.db.gz" -mtime +7 -delete

echo "Backup completed: database_$DATE.db.gz"
```

### Recovery Procedure
```bash
# Stop application
docker-compose down

# Restore from backup
gunzip -c /backups/database_20240613_120000.db.gz > ./data/database.db

# Restart application
docker-compose up -d

# Verify recovery
curl http://localhost:8080/ads?placement=homepage&status=active
```

## Incident Response Checklist

### Immediate Actions (0-15 minutes)
- [ ] Assess impact and scope
- [ ] Check application logs: `docker logs admoai-test-app-1`
- [ ] Verify service status: `curl http://localhost:8080/metrics`
- [ ] Check system resources: `docker stats`
- [ ] Notify stakeholders if critical

### Investigation (15-60 minutes)
- [ ] Collect logs and metrics
- [ ] Identify root cause
- [ ] Document timeline of events
- [ ] Determine if data corruption occurred

### Resolution (60+ minutes)
- [ ] Implement fix
- [ ] Test functionality
- [ ] Monitor for stability
- [ ] Update monitoring/alerting
- [ ] Document lessons learned

# Quick health check
curl -f http://localhost:8080/metrics > /dev/null && echo "OK" || echo "FAIL"

# Database backup
docker exec admoai-test-app-1 sqlite3 /sqliteData/database.db .dump > backup.sql

# Log analysis
docker logs admoai-test-app-1 2>&1 | grep -i error | tail -20

# Memory analysis
docker exec admoai-test-app-1 cat /proc/meminfo

# Disk space check
docker exec admoai-test-app-1 df -h
```