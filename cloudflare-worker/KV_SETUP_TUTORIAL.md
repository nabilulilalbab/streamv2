# 🗄️ KV Setup Tutorial - Complete Step-by-Step

Tutorial lengkap setup Cloudflare KV untuk proxy rotation.

---

## 📋 Overview

Anda akan:
1. Create KV namespace
2. Bind KV ke Worker
3. Upload proxy list ke KV
4. Update worker code
5. Test

**Time needed:** ~10 minutes

---

## 🚀 Step-by-Step Guide

### Step 1: Create KV Namespace

1. **Login ke Cloudflare Dashboard:**
   ```
   https://dash.cloudflare.com/
   ```

2. **Navigate to Workers & Pages:**
   - Di sidebar kiri, klik **"Workers & Pages"**

3. **Go to KV:**
   - Klik tab **"KV"** (di bagian atas)
   - Atau direct link: https://dash.cloudflare.com/?to=/:account/workers/kv/namespaces

4. **Create Namespace:**
   - Klik tombol biru **"Create namespace"**

5. **Name your namespace:**
   ```
   Namespace name: PROXY_CONFIG
   ```
   - Klik **"Add"**

6. **Namespace Created!**
   - Anda akan lihat namespace baru di list
   - **Copy Namespace ID** (akan digunakan nanti)
   - Format: `1234567890abcdef...`

✅ **Step 1 Complete!**

---

### Step 2: Bind KV to Worker

Sekarang kita akan menghubungkan KV namespace ke Worker.

#### Option A: Via Dashboard (MUDAH)

1. **Go to your Worker:**
   - Workers & Pages → Workers (tab)
   - Klik worker Anda (contoh: `idlix-proxy`)

2. **Open Settings:**
   - Klik tab **"Settings"**

3. **Add Binding:**
   - Scroll ke section **"Bindings"**
   - Klik **"Add"** dropdown
   - Pilih **"KV namespace"**

4. **Configure Binding:**
   
   Di dialog yang muncul:
   
   ```
   Variable name: PROXY_CONFIG
   ```
   - **PENTING:** Nama harus EXACT `PROXY_CONFIG` (case-sensitive)
   - Ini nama variable yang digunakan di worker code
   
   ```
   KV namespace: (pilih namespace yang tadi dibuat)
   ```
   - Select: `PROXY_CONFIG`

5. **Save:**
   - Klik **"Add Binding"** atau **"Save"**

6. **Verify:**
   - Anda akan lihat binding di list:
   ```
   Variable name: PROXY_CONFIG
   KV namespace: PROXY_CONFIG
   Namespace ID: 1234567890abcdef...
   ```

✅ **Step 2 Complete!**

#### Option B: Via wrangler.toml (Advanced)

Edit `wrangler.toml`:

```toml
name = "idlix-proxy"
main = "worker-with-proxy-rotation.js"

# Add KV binding
kv_namespaces = [
  { binding = "PROXY_CONFIG", id = "YOUR_NAMESPACE_ID_HERE" }
]
```

Replace `YOUR_NAMESPACE_ID_HERE` dengan Namespace ID dari Step 1.

---

### Step 3: Upload Proxy List to KV

Sekarang kita akan upload data proxy list ke KV.

#### Option A: Via Dashboard (MUDAH)

1. **Go to KV Namespace:**
   - Workers & Pages → KV
   - Klik namespace **"PROXY_CONFIG"**

2. **Add Entry:**
   - Klik tombol **"Add entry"**

3. **Enter Key-Value:**
   
   **Key:**
   ```
   proxy-list
   ```
   
   **Value:**
   Copy semua isi dari `kvProxyList.json`:
   ```json
   {
     "proxies": [
       "http://proxy1.example.com:8080",
       "http://proxy2.example.com:8080",
       "http://proxy3.example.com:3128"
     ],
     "strategy": "random",
     "config": {
       "enabled": true,
       "maxRetries": 3,
       "timeout": 30000
     }
   }
   ```
   
   **PENTING:** Ganti dengan proxy Anda yang actual!

4. **Add:**
   - Klik tombol **"Add"**

5. **Verify:**
   - Anda akan lihat entry baru:
   ```
   Key: proxy-list
   Value: {JSON content}
   ```

✅ **Step 3 Complete!**

#### Option B: Via Wrangler CLI

```bash
# 1. Install wrangler
npm install -g wrangler

# 2. Login
wrangler login

# 3. Upload proxy list
wrangler kv:key put --binding=PROXY_CONFIG "proxy-list" \
  --path=kvProxyList.json \
  --namespace-id=YOUR_NAMESPACE_ID

# 4. Verify
wrangler kv:key get --binding=PROXY_CONFIG "proxy-list" \
  --namespace-id=YOUR_NAMESPACE_ID
```

---

### Step 4: Update Worker Code

Sekarang deploy worker yang support KV.

1. **Go to Worker:**
   - Workers & Pages → Your Worker

2. **Edit Code:**
   - Klik **"Edit code"**

3. **Replace with Rotation-Enabled Worker:**
   
   - **Delete all** existing code (Ctrl+A, Delete)
   
   - **Copy** semua isi dari `worker-with-proxy-rotation.js`
   
   - **Paste** ke editor

4. **Important Code Sections:**

   Look for this in the code:
   ```javascript
   // This line uses KV
   if (typeof PROXY_CONFIG !== 'undefined') {
     const kvData = await PROXY_CONFIG.get('proxy-list', 'json');
   ```
   
   `PROXY_CONFIG` harus match dengan binding name di Step 2!

5. **Save and Deploy:**
   - Klik **"Save and Deploy"**
   - Wait ~30 seconds for global deployment

✅ **Step 4 Complete!**

---

### Step 5: Test KV Integration

Test apakah KV berfungsi.

#### Test 1: Check Binding

1. **Go to Worker Settings:**
   - Settings → Bindings

2. **Verify you see:**
   ```
   KV namespace bindings
   Variable name: PROXY_CONFIG
   ```

#### Test 2: Test Request

```bash
# Replace with your worker URL
WORKER_URL="https://idlix-proxy.your-account.workers.dev"

# Test request
curl "$WORKER_URL?url=https://jeniusplay.com/cdn/hls/test/master.m3u8"
```

#### Test 3: Check Logs

```bash
# Install wrangler if not already
npm install -g wrangler

# Login
wrangler login

# Tail logs
wrangler tail idlix-proxy
```

Look for:
```
Loaded 3 proxies from KV
Attempt 1: Using proxy http://proxy1.example.com:8080
```

If you see this, KV is working! ✅

---

## 🎯 Complete Setup Verification

Check all these:

- [ ] KV namespace created ✅
- [ ] KV bound to worker (Variable: PROXY_CONFIG) ✅
- [ ] Proxy list uploaded to KV (Key: proxy-list) ✅
- [ ] Worker deployed with rotation code ✅
- [ ] Test request successful ✅
- [ ] Logs show proxy usage ✅

---

## 🔄 Update Proxy List (Later)

Kapanpun Anda ingin update proxy list:

### Via Dashboard (MUDAH):

1. Workers & Pages → KV → PROXY_CONFIG
2. Click key: `proxy-list`
3. Click **"Edit"**
4. Update JSON
5. Click **"Save"**

**Changes take effect immediately!** No redeploy needed.

### Via Wrangler:

```bash
wrangler kv:key put --binding=PROXY_CONFIG "proxy-list" \
  --path=kvProxyList.json
```

---

## 📝 Example: Edit Your Proxy List

1. **Open `kvProxyList.json`**

2. **Edit proxies:**
   ```json
   {
     "proxies": [
       "http://YOUR_PROXY_1:8080",
       "http://YOUR_PROXY_2:3128",
       "http://username:password@YOUR_PROXY_3:8080"
     ],
     "strategy": "random"
   }
   ```

3. **Upload to KV:**
   - Via Dashboard: Copy-paste into KV entry
   - Via Wrangler: `wrangler kv:key put ...`

---

## 🎲 Rotation Strategies

Change strategy in `kvProxyList.json`:

### Random (Default):
```json
{
  "strategy": "random"
}
```
- Each request picks random proxy
- Good for load distribution

### Round-Robin:
```json
{
  "strategy": "round-robin"
}
```
- Cycles through proxies in order
- Predictable pattern

---

## 🐛 Troubleshooting

### Problem: "PROXY_CONFIG is not defined"

**Solution:**
- Check binding name in Settings → Bindings
- Must be exactly: `PROXY_CONFIG` (case-sensitive)
- Redeploy worker after fixing

### Problem: "Failed to load from KV"

**Solution:**
- Verify KV entry exists:
  - Key: `proxy-list`
  - Value: Valid JSON
- Check namespace ID matches

### Problem: Proxies not being used

**Solution:**
1. Check logs: `wrangler tail`
2. Look for: "Loaded X proxies from KV"
3. If 0 proxies, check KV data format

### Problem: All proxies failing

**Solution:**
- Worker will fallback to direct fetch
- Check logs for: "falling back to direct fetch"
- Verify proxy URLs are correct
- Test proxies manually with curl

---

## 💰 KV Pricing

### Free Tier (Your Plan):
```
✅ 100,000 reads/day
✅ 1,000 writes/day  
✅ 1 GB storage
```

### Your Usage (Estimated):
```
Reads: ~12/hour (with 5min cache)
      = ~288/day
      
Writes: 1/day (when you update)

Storage: <1 KB
```

**Verdict:** Way below free tier! ✅

---

## ✅ Quick Reference

### KV Operations via Dashboard:

**Read entry:**
- KV → PROXY_CONFIG → Click key

**Update entry:**
- KV → PROXY_CONFIG → Click key → Edit → Save

**Delete entry:**
- KV → PROXY_CONFIG → Select key → Delete

### KV Operations via Wrangler:

```bash
# Read
wrangler kv:key get --binding=PROXY_CONFIG "proxy-list"

# Write
wrangler kv:key put --binding=PROXY_CONFIG "proxy-list" \
  --path=kvProxyList.json

# Delete
wrangler kv:key delete --binding=PROXY_CONFIG "proxy-list"

# List all keys
wrangler kv:key list --binding=PROXY_CONFIG
```

---

## 🎯 Next Steps After Setup

1. **Test thoroughly:**
   - Try multiple requests
   - Check logs for proxy usage
   - Monitor success rate

2. **Add real proxies:**
   - Replace example proxies
   - Test each proxy manually first
   - Add 3-5 reliable proxies

3. **Monitor performance:**
   - Check Worker metrics
   - Track KV read count
   - Monitor error rate

4. **Optimize:**
   - Adjust cache TTL if needed
   - Change rotation strategy
   - Remove slow proxies

---

## 📞 Need Help?

**Cloudflare Docs:**
- KV: https://developers.cloudflare.com/kv/
- Workers: https://developers.cloudflare.com/workers/

**Community:**
- Discord: https://discord.gg/cloudflaredev
- Forum: https://community.cloudflare.com/

**This Setup:**
- Check other .md files in this folder
- Review worker code comments
- Test with sample data first

---

## 🎉 Congratulations!

Jika semua step sudah selesai:

✅ KV namespace created  
✅ Binding configured  
✅ Proxy list uploaded  
✅ Worker deployed  
✅ Rotation working  

Your Cloudflare Worker now has dynamic proxy rotation! 🚀

You can update proxies anytime without redeploying!

---

**Setup Time:** ~10 minutes  
**Maintenance:** Update proxy list as needed  
**Cost:** FREE (within limits)
