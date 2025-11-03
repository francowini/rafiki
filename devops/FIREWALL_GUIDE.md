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

## Option 2: UFW (Uncomplicated Firewall) - Server-based

If you prefer to configure the firewall directly on the server using UFW:

### Step 1: SSH into your server

```bash
ssh root@178.156.170.37
```

### Step 2: Install UFW (if not installed)

```bash
apt update
apt install ufw -y
```

**What to expect:** UFW is installed.

### Step 3: Set default policies

```bash
ufw default deny incoming
ufw default allow outgoing
```

**What to expect:** By default, all incoming connections are blocked, all outgoing allowed.

### Step 4: Allow SSH (IMPORTANT!)

```bash
ufw allow 22/tcp comment 'SSH access'
```

**What to expect:** SSH port 22 is allowed. **DO THIS BEFORE ENABLING UFW** or you'll lock yourself out!

### Step 5: Allow HTTP and HTTPS

```bash
ufw allow 80/tcp comment 'HTTP for Let's Encrypt'
ufw allow 443/tcp comment 'HTTPS API'
```

**What to expect:** Ports 80 and 443 are now allowed.

### Step 6: (Optional) Restrict SSH to your IP only

Get your current IP:

```bash
curl ifconfig.me
```

Suppose your IP is `200.123.45.67`, then:

```bash
ufw delete allow 22/tcp
ufw allow from 200.123.45.67 to any port 22 proto tcp comment 'SSH from my IP'
```

**What to expect:** SSH is now only accessible from your IP address.

### Step 7: Enable UFW

```bash
ufw enable
```

**What to expect:**
- Warning: "Command may disrupt existing ssh connections. Proceed with operation (y|n)?"
- Type `y` and press Enter
- Output: "Firewall is active and enabled on system startup"

### Step 8: Verify firewall status

```bash
ufw status verbose
```

**What to expect:** You should see:

```
Status: active

To                         Action      From
--                         ------      ----
22/tcp                     ALLOW       200.123.45.67    # SSH from my IP
80/tcp                     ALLOW       Anywhere         # HTTP for Let's Encrypt
443/tcp                    ALLOW       Anywhere         # HTTPS API
```

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

## Additional Security Recommendations

### 1. Rate Limiting (Already configured in nginx)

Your nginx config already includes:

```nginx
limit_req_zone $binary_remote_addr zone=api_limit:10m rate=10r/s;
```

This limits clients to 10 requests per second.

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
