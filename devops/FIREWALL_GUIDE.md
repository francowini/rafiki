# Firewall Configuration Guide for Hetzner Server

This guide shows you how to configure the firewall on your Hetzner server to allow HTTP/HTTPS traffic while keeping your server secure.

---

## Understanding the Ports

| Port | Service | Purpose | Public Access |
|------|---------|---------|---------------|
| 22   | SSH     | Remote access | ⚠️ Restricted (your IP only) |
| 80   | HTTP    | Let's Encrypt validation, redirect to HTTPS | ✅ Public |
| 443  | HTTPS   | Main API (SSL/TLS) | ✅ Public |
| 3000 | Go API  | Backend service | ❌ Internal only (docker network) |
| 3010 | Debug   | Metrics/profiling | ⚠️ Restricted or internal only |
| 3100 | Grafana | Observability | ⚠️ Optional (restrict or tunnel) |

---

## Option 1: Hetzner Cloud Firewall (Recommended)

Hetzner Cloud provides a managed firewall through their control panel. This is the easiest and recommended approach.

### Step 1: Access Hetzner Cloud Console

1. Go to https://console.hetzner.cloud/
2. Select your project
3. Click on "Firewalls" in the left sidebar

### Step 2: Create a new Firewall

Click "Create Firewall" and name it `rafiki-production`

### Step 3: Configure Inbound Rules

Add these rules:

```
Rule 1: SSH
- Protocol: TCP
- Port: 22
- Source: Your IP address (e.g., 200.123.45.67/32)
- Description: SSH access from my IP

Rule 2: HTTP
- Protocol: TCP
- Port: 80
- Source: 0.0.0.0/0 (anywhere)
- Description: HTTP for Let's Encrypt and redirect

Rule 3: HTTPS
- Protocol: TCP
- Port: 443
- Source: 0.0.0.0/0 (anywhere)
- Description: HTTPS API access

Rule 4: ICMP (optional)
- Protocol: ICMP
- Source: 0.0.0.0/0
- Description: Allow ping
```

### Step 4: Apply to your server

- Select your server from the list
- Click "Apply Firewall"

**What to expect:** Firewall rules are applied immediately. Your server is now protected.

---

## Testing Firewall Configuration

### Test 1: Check SSH still works

```bash
ssh root@178.156.170.37
```

**What to expect:** You can still connect via SSH.

### Test 2: Check HTTP is accessible

From your local machine:

```bash
curl -v http://api.rafiki.lat
```

**What to expect:** Connection succeeds (even if content is redirected).

### Test 3: Check HTTPS is accessible

From your local machine:

```bash
curl -v https://api.rafiki.lat/v1/liveness
```

**What to expect:** Connection succeeds with SSL certificate, JSON response.

### Test 4: Verify internal ports are NOT accessible

From your local machine (this should FAIL):

```bash
curl http://178.156.170.37:3000
```

**What to expect:** Connection timeout or "Connection refused". This is GOOD - internal ports should not be publicly accessible.

---

### 2. Fail2Ban (Optional but recommended)

Install Fail2Ban to automatically ban IPs that try to brute-force SSH:

```bash
apt install fail2ban -y
systemctl enable fail2ban
systemctl start fail2ban
```

**What to expect:** Fail2Ban monitors SSH login attempts and bans IPs after repeated failures.

### 3. Disable Password Authentication (Use SSH keys only)

Edit SSH config:

```bash
nano /etc/ssh/sshd_config
```

Change these lines:

```
PasswordAuthentication no
PubkeyAuthentication yes
```

Restart SSH:

```bash
systemctl restart sshd
```

**What to expect:** Only SSH key authentication is allowed. Make sure you have your key configured BEFORE doing this!

---

## Troubleshooting

### "I locked myself out of SSH!"

If you used Hetzner Cloud Firewall:
1. Go to Hetzner Console
2. Use the web-based console (VNC)
3. Fix firewall rules

If you used UFW:
1. Go to Hetzner Console
2. Use rescue mode or web console
3. Run: `ufw disable`
4. Reconfigure properly

### "I can't access HTTPS"

Check firewall allows port 443:

```bash
ufw status | grep 443
```

Test from server itself:

```bash
curl -v https://localhost/v1/liveness
```

Check nginx is running:

```bash
docker compose --profile production ps | grep nginx
```

### "Let's Encrypt fails during certificate request"

Port 80 must be open:

```bash
ufw status | grep 80
```

Test from outside:

```bash
curl http://api.rafiki.lat/.well-known/acme-challenge/test
```

---

## Summary

After configuring firewall:
- ✅ Ports 80 and 443 are publicly accessible
- ✅ SSH is restricted (optional but recommended)
- ✅ Internal ports (3000, 3010, 3100) are NOT exposed
- ✅ Nginx handles rate limiting
- ✅ Server is secure

---

## Next Steps

1. Follow SSL_SETUP_GUIDE.md to get SSL certificate
2. Test your API endpoints
3. Update CORS for production
4. Deploy frontend to Vercel
