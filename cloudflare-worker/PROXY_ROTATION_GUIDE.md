# 🔄 Proxy Rotation Guide

Complete guide untuk setup proxy rotation di Cloudflare Worker.

---

## 📊 Overview

Cloudflare Worker mendukung proxy rotation dengan 2 metode:

1. **KV Storage** (Recommended) - Dynamic, update tanpa redeploy
2. **Hardcoded** (Simple) - Static list dalam code

---

## 🎯 Method 1: KV Storage (RECOMMENDED)

Dynamic proxy rotation dengan Cloudflare KV.

### Step 1: Create KV Namespace

1. **Login ke Cloudflare Dashboard**

2. **Navigate to KV:**
   - Workers & Pages → KV
   - Click **"Create namespace"**

3. **Name your namespace:**
   ```
   Name: PROXY_CONFIG
   ```
   - Click **"Add"**

4. **Copy Namespace ID:**
   - You'll see: `Namespace ID: abc123...`
   - **Save this ID!**

---

### Step 2: Bind KV to Worker

1. **Go to your Worker:**
   - Workers & Pages → Your Worker → Settings

2. **Add KV Binding:**
   - Variables → KV Namespace Bindings
   - Click **"Add binding"**

3. **Configure binding:**
   ```
   Variable name: PROXY_CONFIG
   KV namespace: (select your namespace)
   ```
   - Click **"Save"**

---

### Step 3: Upload Proxy List to KV

#### Option A: Via Dashboard (Easy)

1. **Go to KV namespace:**
   - Workers & Pages → KV → Your namespace

2. **Add new entry:**
   - Click **"Add entry"**
   - Key: `proxy-list`
   - Value: (paste content from `kvProxyList.json`)
   
3. **Click "Add"**

#### Option B: Via Wrangler CLI

```bash
# Upload proxy list
wrangler kv:key put --binding=PROXY_CONFIG "proxy-list" --path=kvProxyList.json

# Verify
wrangler kv:key get --binding=PROXY_CONFIG "proxy-list"
```

#### Option C: Via API

```bash
# Get Account ID and Namespace ID from dashboard
ACCOUNT_ID="your-account-id"
NAMESPACE_ID="your-namespace-id"
API_TOKEN="your-api-token"

# Upload
curl -X PUT "https://api.cloudflare.com/client/v4/accounts/$ACCOUNT_ID/storage/kv/namespaces/$NAMESPACE_ID/values/proxy-list" \
  -H "Authorization: Bearer $API_TOKEN" \
  -H "Content-Type: application/json" \
  --data @kvProxyList.json
```

---

### Step 4: Deploy Worker with KV Support

1. **Use the rotation-enabled worker:**
   - Copy `worker-with-proxy-rotation.js`
   - Replace current worker code

2. **Deploy:**
   - Click **"Save and Deploy"**

3. **Verify KV binding:**
   - Should show: `PROXY_CONFIG → Your namespace`

---

### Step 5: Configure Proxy List

Edit `kvProxyList.json`:

```json
{
  "proxies": [
    "http://proxy1.example.com:8080",
    "http://proxy2.example.com:3128",
    "http://username:password@proxy3.com:8080"
  ],
  "strategy": "random",
  "config": {
    "enabled": true,
    "maxRetries": 3,
    "timeout": 30000
  }
}
```

**Fields:**
- `proxies`: Array of proxy URLs
- `strategy`: `"random"` or `"round-robin"`
- `config.maxRetries`: How many proxies to try before giving up
- `config.timeout`: Request timeout in milliseconds

---

### Step 6: Test Proxy Rotation

```bash
# Test request
curl "https://your-worker.workers.dev?url=https://jeniusplay.com/test.m3u8"

# Check logs (should show which proxy was used)
wrangler tail your-worker
```

---

## 🔧 Method 2: Hardcoded Proxies (Simple)

Untuk proxy list yang jarang berubah.

### Step 1: Edit Worker Code

Edit `worker-with-proxy-rotation.js`, find `getProxyList()` function:

```javascript
async function getProxyList() {
  // Skip KV, return hardcoded list
  const hardcodedProxies = [
    'http://proxy1.example.com:8080',
    'http://proxy2.example.com:8080',
    'http://proxy3.example.com:3128',
  ];
  
  return hardcodedProxies;
}
```

### Step 2: Deploy

- Copy updated code
- Save and Deploy

**Pros:**
- ✅ Simpler setup
- ✅ No KV needed
- ✅ Faster (no KV reads)

**Cons:**
- ❌ Must redeploy to update proxies
- ❌ Not dynamic

---

## 🎲 Rotation Strategies

### Random Selection

```json
{
  "strategy": "random"
}
```

**How it works:**
- Each request picks random proxy
- Good for load distribution
- Unpredictable pattern

**Use when:**
- Need even load distribution
- Proxies have similar performance

### Round-Robin

```json
{
  "strategy": "round-robin"
}
```

**How it works:**
- Cycles through proxies in order
- Proxy 1 → Proxy 2 → Proxy 3 → Proxy 1...
- Predictable pattern

**Use when:**
- Need fair rotation
- Testing specific proxies
- Debugging issues

---

## 🔄 Update Proxy List (KV)

### Via Dashboard

1. Go to KV namespace
2. Find `proxy-list` key
3. Click **"Edit"**
4. Update JSON
5. Click **"Save"**

**Changes take effect immediately!** No redeploy needed.

### Via Wrangler CLI

```bash
# Update proxy list
wrangler kv:key put --binding=PROXY_CONFIG "proxy-list" --path=kvProxyList.json

# Verify changes
wrangler kv:key get --binding=PROXY_CONFIG "proxy-list"
```

### Via API

```bash
curl -X PUT "https://api.cloudflare.com/client/v4/accounts/$ACCOUNT_ID/storage/kv/namespaces/$NAMESPACE_ID/values/proxy-list" \
  -H "Authorization: Bearer $API_TOKEN" \
  -H "Content-Type: application/json" \
  --data @kvProxyList.json
```

---

## 🔍 Proxy List Format

### HTTP Proxy

```
http://host:port
http://username:password@host:port
```

Example:
```json
{
  "proxies": [
    "http://proxy.example.com:8080",
    "http://user:pass@proxy2.com:3128"
  ]
}
```

### HTTPS Proxy

```
https://host:port
https://username:password@host:port
```

### SOCKS5 Proxy

```
socks5://host:port
socks5://username:password@host:port
```

**Note:** Cloudflare Workers has limitations with SOCKS5. HTTP/HTTPS proxies work best.

---

## ⚠️ Important Notes

### Cloudflare Worker Limitations

1. **No native proxy support:**
   - Workers can't use CONNECT method
   - Need HTTP proxy that accepts URL parameter
   - Example: `http://proxy.com/?url=target`

2. **Workaround:**
   - Use HTTP proxy services
   - Or use Cloudflare as CDN/proxy itself
   - Or use external proxy relay service

### Alternative Approach

Instead of proxying through external proxies, consider:

**Option A: Cloudflare as Proxy**
- Worker → Cloudflare CDN → JeniusPlay
- No external proxies needed
- Cloudflare's global network acts as proxy

**Option B: Residential Proxy Services**
- Use API-based proxy services
- Example: BrightData, Oxylabs, Smartproxy
- They provide HTTP API endpoints

**Option C: Your Own Proxy Pool**
- Deploy simple HTTP proxies
- Accept requests via URL parameter
- Worker rotates between your proxies

---

## 🎯 Recommended Proxy Services

### For Streaming

1. **Residential Proxies:**
   - BrightData (luminati.io)
   - Smartproxy (smartproxy.com)
   - Oxylabs (oxylabs.io)

2. **Datacenter Proxies:**
   - ProxyMesh (proxymesh.com)
   - Storm Proxies (stormproxies.com)
   - Rayobyte (rayobyte.com)

3. **Free Proxies (NOT recommended for production):**
   - free-proxy-list.net
   - geonode.com/free-proxy-list
   - proxy-list.download

---

## 📊 Monitoring

### Check Proxy Usage

View logs:
```bash
wrangler tail your-worker
```

Look for:
```
Attempt 1: Using proxy http://proxy1.com:8080
Success with proxy http://proxy1.com:8080
```

### Metrics to Monitor

1. **Success Rate:**
   - How many requests succeed per proxy
   - Remove proxies with low success rate

2. **Latency:**
   - Measure response time per proxy
   - Prefer faster proxies

3. **Reliability:**
   - Track downtime
   - Remove unreliable proxies

---

## 🐛 Troubleshooting

### Proxies Not Working

1. **Check KV binding:**
   ```bash
   wrangler kv:key get --binding=PROXY_CONFIG "proxy-list"
   ```

2. **Verify proxy format:**
   - Must be valid URL
   - Include protocol (http://)

3. **Test proxy manually:**
   ```bash
   curl -x http://proxy.com:8080 https://jeniusplay.com
   ```

### All Proxies Failing

1. **Check fallback:**
   - Worker should fallback to direct fetch
   - Look for "falling back to direct fetch" in logs

2. **Verify proxy authentication:**
   - Check username/password
   - Test with curl

3. **Check proxy connectivity:**
   - Ensure proxies are reachable
   - Check firewall rules

---

## 💰 Cost Estimate

### KV Storage (Free Tier)

**Included:**
- 100,000 reads/day
- 1,000 writes/day
- 1 GB storage

**Typical usage:**
- 1 request = 1 KV read (cached 5 min)
- With caching: ~12 reads/hour
- **Way below free tier limit!**

### Paid Plan ($5/month)

If you exceed free tier:
- 10 million reads/month
- 1 million writes/month
- More storage

---

## ✅ Best Practices

1. **Cache Proxy List:**
   - Worker caches for 5 minutes
   - Reduces KV reads
   - Improves performance

2. **Test Proxies:**
   - Verify each proxy before adding
   - Remove dead proxies promptly

3. **Monitor Performance:**
   - Track success rates
   - Remove slow proxies

4. **Use Authentication:**
   - More reliable than free proxies
   - Better performance

5. **Rotate Regularly:**
   - Update proxy list weekly
   - Add new proxies
   - Remove bad ones

6. **Set Reasonable Limits:**
   - Max retries: 3
   - Timeout: 30 seconds
   - Don't hammer proxies

---

## 🚀 Quick Start Checklist

For KV-based rotation:

- [ ] Create KV namespace
- [ ] Bind KV to worker (name: PROXY_CONFIG)
- [ ] Edit kvProxyList.json with your proxies
- [ ] Upload to KV
- [ ] Deploy worker-with-proxy-rotation.js
- [ ] Test with sample request
- [ ] Monitor logs
- [ ] Adjust rotation strategy if needed

---

## 📞 Support

**Cloudflare KV Docs:**
https://developers.cloudflare.com/workers/runtime-apis/kv/

**Wrangler CLI:**
https://developers.cloudflare.com/workers/wrangler/

**Worker Examples:**
https://developers.cloudflare.com/workers/examples/

---

**Status:** ✅ Production Ready
**Deployment:** ~10 minutes
**Maintenance:** Update proxy list as needed
