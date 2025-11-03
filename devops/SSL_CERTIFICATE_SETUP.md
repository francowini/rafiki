# SSL Certificate Setup for Production

Documentation of SSL certificate setup process for `api.rafiki.lat` using Let's Encrypt and Nginx.

## Infrastructure Overview

- **Domain:** api.rafiki.lat → 178.156.170.37
- **SSL Provider:** Let's Encrypt (via Certbot)
- **Reverse Proxy:** Nginx 1.27-alpine
- **Deployment:** Docker Compose with production profile

## Files Created

```
nginx/nginx.conf              # Nginx reverse proxy configuration with SSL
docker-compose.yml            # Added nginx and certbot services (profile: production)
.env.production              # Production CORS settings (no localhost)
```

## Nginx Configuration

Location: `nginx/nginx.conf`

**Port 80 (HTTP):**
- Handles Let's Encrypt ACME challenges (/.well-known/acme-challenge/)
- Redirects all other traffic to HTTPS

**Port 443 (HTTPS):**
- SSL termination with Let's Encrypt certificates
- TLS 1.2 and 1.3 with strong cipher suites
- HTTP/2 enabled with modern syntax: `http2 on;`
- Rate limiting: 10 req/s per IP (burst 20)
- Security headers: HSTS, X-Content-Type-Options, X-Frame-Options, X-XSS-Protection
- Proxies to `partner-service:3000`
- Health checks (`/v1/liveness`, `/v1/readiness`) exempt from rate limiting

## Certificate Acquisition Steps

1. **Deployed code** including `nginx/nginx.conf` to server
2. **Started backend services** without nginx (postgres, tempo, partner-service)
3. **Ran temporary nginx** on port 80 for ACME challenge
4. **Requested certificate** using certbot with webroot method
5. **Stopped temporary nginx** and all services
6. **Started production stack** with `docker compose --profile production up -d --build`
7. **Verified HTTPS** at https://api.rafiki.lat/v1/liveness

## Certificate Auto-Renewal

The certbot container runs continuously and attempts renewal every 12 hours automatically.

## Production Services

Using Docker Compose profile `production` adds:
- **nginx:** Reverse proxy with SSL (ports 80, 443)
- **certbot:** Certificate renewal daemon

Start production: `docker compose --profile production up -d --build`

## Key Configuration Details

**nginx.conf syntax changes for nginx 1.27+:**
- Old: `listen 443 ssl http2;`
- New: `listen 443 ssl;` + `http2 on;`

**Rate limiting:**
- Health checks have no `limit_req` directive (not rate limited)
- All other endpoints: `limit_req zone=api_limit burst=20 nodelay;`

**CORS:**
- Development: `http://localhost:3000,http://localhost:3001,*`
- Production: `https://app.rafiki.lat,https://rafiki-frontend-*.vercel.app`
