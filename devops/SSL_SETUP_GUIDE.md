# SSL Certificate Setup Guide for api.rafiki.lat

This guide will help you obtain an SSL certificate from Let's Encrypt for your production server.

---

## Prerequisites

- ✅ DNS configured: `api.rafiki.lat` points to your Hetzner server IP
- ✅ SSH access to your Hetzner server
- ✅ Ports 80 and 443 open on firewall
- ✅ Code deployed on server

---

## Step 1: SSH into your Hetzner server

```bash
ssh root@178.156.170.37
```

**What to expect:** You should successfully connect to your server.

---

## Step 2: Navigate to project directory

```bash
cd /path/to/rafiki
```

**What to expect:** You're now in the rafiki project root directory where `docker-compose.yml` exists.

---

## Step 3: Ensure you have the .env file configured

```bash
cat .env
```

**What to expect:** You should see your environment variables including `CORS_ALLOWED_ORIGINS`.

If the file doesn't exist, create it:

```bash
ln -sf .env.production .env
```

---

## Step 4: Create certbot directories

```bash
mkdir -p certbot/www
mkdir -p certbot/conf
```

**What to expect:** Two directories are created for Let's Encrypt files.

---

## Step 5: Stop any running services

```bash
docker compose --profile production down
```

**What to expect:** All containers stop. Output shows containers being removed.

---

## Step 6: Start services WITHOUT nginx

We need the backend running but NOT nginx (chicken-and-egg problem).

```bash
docker compose up -d postgres tempo partner-service
```

**What to expect:** Three containers start (postgres, tempo, partner-service). No nginx yet.

---

## Step 7: Wait for services to be healthy

```bash
sleep 15
```

**What to expect:** Just wait 15 seconds for services to initialize.

Check health:

```bash
docker compose ps
```

**What to expect:** All three services show "Up (healthy)" status.

---

## Step 8: Create temporary nginx configuration

This nginx config ONLY handles Let's Encrypt validation (no SSL yet).

```bash
cat > /tmp/nginx-certbot-init.conf << 'EOF'
server {
    listen 80;
    listen [::]:80;
    server_name api.rafiki.lat;

    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }

    location / {
        return 200 "OK";
        add_header Content-Type text/plain;
    }
}
EOF
```

**What to expect:** A temporary nginx config file is created at `/tmp/nginx-certbot-init.conf`.

---

## Step 9: Start temporary nginx

```bash
docker run -d --rm \
    --name temp-nginx-certbot \
    -p 80:80 \
    -v /tmp/nginx-certbot-init.conf:/etc/nginx/conf.d/default.conf:ro \
    -v "$(pwd)/certbot/www:/var/www/certbot:ro" \
    nginx:1.27-alpine
```

**What to expect:** A temporary nginx container starts on port 80.

Verify it's running:

```bash
docker ps | grep temp-nginx
```

**What to expect:** You should see `temp-nginx-certbot` container running.

Test it:

```bash
curl http://api.rafiki.lat
```

**What to expect:** Response: `OK`

---

## Step 10: Request SSL certificate from Let's Encrypt

This is the main command that gets your SSL certificate.

```bash
docker run --rm \
    -v "$(pwd)/certbot/www:/var/www/certbot:rw" \
    -v "$(pwd)/certbot/conf:/etc/letsencrypt:rw" \
    certbot/certbot certonly \
    --webroot \
    --webroot-path=/var/www/certbot \
    --email francoolsson1995@gmail.com \
    --agree-tos \
    --no-eff-email \
    -d api.rafiki.lat
```

**What to expect:**
- Output shows "Requesting a certificate for api.rafiki.lat"
- Let's Encrypt validates your domain
- Success message: "Successfully received certificate"
- Certificate location: `/etc/letsencrypt/live/api.rafiki.lat/fullchain.pem`

**If it fails:**
- Check DNS is correct: `nslookup api.rafiki.lat`
- Check port 80 is accessible: `curl http://api.rafiki.lat`
- Check firewall allows port 80

---

## Step 11: Stop temporary nginx

```bash
docker stop temp-nginx-certbot
```

**What to expect:** Temporary nginx container stops and is removed.

---

## Step 12: Stop all services

```bash
docker compose down
```

**What to expect:** All containers stop.

---

## Step 13: Start ALL services with production profile

Now we start the REAL nginx with SSL certificates.

```bash
docker compose --profile production up -d --build
```

**What to expect:**
- All containers start: postgres, tempo, grafana, partner-service, **nginx**, **certbot**
- Nginx now has SSL certificates and serves HTTPS

---

## Step 14: Wait for services to start

```bash
sleep 10
```

---

## Step 15: Verify services are running

```bash
docker compose --profile production ps
```

**What to expect:** You should see 6 containers running:
- rafiki-postgres
- rafiki-tempo
- rafiki-grafana
- rafiki-service
- **rafiki-nginx** ← New!
- **rafiki-certbot** ← New!

---

## Step 16: Test your API with HTTPS

```bash
curl https://api.rafiki.lat/v1/liveness
```

**What to expect:**
```json
{
  "status": "ok"
}
```

And you should see a valid SSL certificate (no warnings).

---

## Step 17: Test in browser

Open in your browser:

```
https://api.rafiki.lat/v1/liveness
```

**What to expect:**
- 🔒 Green padlock in address bar
- Valid SSL certificate
- JSON response

---

## Troubleshooting

### If Step 10 fails (certificate request):

Check DNS:
```bash
nslookup api.rafiki.lat
```

Check port 80 is accessible from internet:
```bash
curl -v http://api.rafiki.lat
```

Check temporary nginx logs:
```bash
docker logs temp-nginx-certbot
```

### If Step 16 fails (HTTPS doesn't work):

Check nginx logs:
```bash
docker compose --profile production logs nginx
```

Check certificate exists:
```bash
ls -la certbot/conf/live/api.rafiki.lat/
```

Check nginx is running:
```bash
docker compose --profile production ps
```

---

## Certificate Auto-Renewal

The `certbot` container automatically renews your certificate every 12 hours. You don't need to do anything!

To manually force a renewal:

```bash
docker compose --profile production exec certbot certbot renew
```

---

## Summary

After completing these steps:
- ✅ SSL certificate obtained from Let's Encrypt
- ✅ Nginx serving HTTPS on port 443
- ✅ HTTP (port 80) redirects to HTTPS
- ✅ Auto-renewal configured
- ✅ API accessible at `https://api.rafiki.lat`

---

## Next Steps

1. Configure firewall (see FIREWALL_GUIDE.md)
2. Update CORS for production (see deployment docs)
3. Test your API endpoints
4. Deploy your frontend to Vercel
