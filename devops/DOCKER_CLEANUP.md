# Docker Cleanup Commands for Server

## Quick Nuclear Option (Clean Everything)

```bash
# Stop all running containers
docker stop $(docker ps -aq)

# Remove all containers
docker rm $(docker ps -aq)

# Remove all networks (except default bridge/host/none)
docker network prune -f

# Optional: Remove all images (if you want to rebuild from scratch)
docker image prune -a -f
```

## More Controlled Cleanup (Recommended)

```bash
# 1. Stop all Rafiki containers
cd /opt/rafiki
docker compose -f docker-compose.yml -f docker-compose.prod.yml --profile production down

# 2. If that doesn't work, force stop all containers
docker stop $(docker ps -aq)

# 3. Remove stopped containers
docker rm $(docker ps -aq)

# 4. Remove unused networks
docker network prune -f

# 5. Now try deployment again
sudo ./devops/deploy.sh
```

## Troubleshooting Specific Port Conflicts

```bash
# Find what's using port 5432 (postgres)
netstat -tulpn | grep 5432
lsof -i :5432

# Find what's using port 3000 (partner-service)
netstat -tulpn | grep 3000
lsof -i :3000

# Find what's using port 80 (nginx)
netstat -tulpn | grep :80
lsof -i :80

# Find what's using port 443 (nginx)
netstat -tulpn | grep :443
lsof -i :443

# Kill specific process by PID
kill -9 <PID>
```

## Safe Cleanup (Preserves Database)

```bash
cd /opt/rafiki

# Stop containers but keep volumes (database preserved)
docker compose -f docker-compose.yml -f docker-compose.prod.yml --profile production down

# Remove only unused networks
docker network prune -f

# Start fresh
sudo ./devops/deploy.sh
```

## Check What's Running

```bash
# List all running containers
docker ps

# List all containers (including stopped)
docker ps -a

# List all networks
docker network ls

# Inspect specific network
docker network inspect rafiki_rafiki-network
```

## Common Issues and Solutions

### Issue: "network with name rafiki_rafiki-network already exists"
```bash
docker network rm rafiki_rafiki-network
# Then retry deployment
```

### Issue: "port is already allocated"
```bash
# Find and stop the container using the port
docker ps | grep <port_number>
docker stop <container_id>
# Or force remove
docker rm -f <container_id>
```

### Issue: Old containers still running
```bash
# Force remove all rafiki-related containers
docker ps -a | grep rafiki | awk '{print $1}' | xargs docker rm -f
```

## After Cleanup

Once you've cleaned up, run deployment:

```bash
cd /opt/rafiki
sudo ./devops/deploy.sh
```

## Emergency Reset (DANGER: Deletes Database!)

**⚠️ WARNING: This deletes ALL data including database!**

Only use if you're setting up from scratch:

```bash
cd /opt/rafiki

# Stop and remove EVERYTHING including volumes
docker compose -f docker-compose.yml -f docker-compose.prod.yml --profile production down -v

# Remove all networks
docker network prune -f

# Remove all containers
docker rm -f $(docker ps -aq)

# Now you need to run first-time setup again
sudo ./devops/deploy.sh
```
