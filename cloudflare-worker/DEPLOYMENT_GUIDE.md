# 🚀 Step-by-Step Deployment Guide

Panduan lengkap deploy Cloudflare Worker untuk IDLIX Streaming Proxy.

---

## 📋 Prerequisites

- [ ] Email address (untuk sign up)
- [ ] Browser (Chrome, Firefox, Safari, dll)
- [ ] 5 menit waktu luang

**Tidak perlu:**
- ❌ Credit card (free tier)
- ❌ Domain sendiri (optional)
- ❌ Technical skills (semua via UI)

---

## 🎯 Method 1: Deploy via Dashboard (RECOMMENDED)

Paling mudah, tidak perlu install apapun.

### Step 1: Create Cloudflare Account

1. **Buka browser** dan pergi ke: https://dash.cloudflare.com/sign-up

2. **Isi form registrasi:**
   - Email: your-email@example.com
   - Password: (buat password yang kuat)
   
3. **Klik "Sign Up"**

4. **Verifikasi email:**
   - Buka email Anda
   - Klik link verifikasi dari Cloudflare
   - Kembali ke dashboard

5. **Skip domain setup** (klik "I don't have a domain")

✅ **Account created!**

---

### Step 2: Create Worker

1. **Di Cloudflare Dashboard:**
   - Klik **"Workers & Pages"** di sidebar kiri
   
2. **Create Worker:**
   - Klik tombol biru **"Create Application"**
   - Pilih **"Create Worker"**
   
3. **Name your worker:**
   - Worker name: `idlix-proxy` (atau nama lain yang Anda mau)
   - Klik **"Deploy"**
   
4. **Worker Created!**
   - Anda akan lihat worker URL:
   ```
   https://idlix-proxy.your-account.workers.dev
   ```
   - **SIMPAN URL INI!** Anda akan butuh nanti.

---

### Step 3: Edit Worker Code

1. **Klik tombol "Edit Code"** (di halaman worker)

2. **Hapus semua code default:**
   - Select All (Ctrl+A / Cmd+A)
   - Delete

3. **Copy code dari `worker.js`:**
   - Buka file `cloudflare-worker/worker.js`
   - Select All (Ctrl+A / Cmd+A)
   - Copy (Ctrl+C / Cmd+C)

4. **Paste ke Cloudflare editor:**
   - Paste (Ctrl+V / Cmd+V)

5. **Save and Deploy:**
   - Klik tombol **"Save and Deploy"** di kanan atas
   - Tunggu ~30 detik
   - Status akan berubah "Deployed"

✅ **Worker deployed globally!**

---

### Step 4: Test Worker

1. **Get your worker URL:**
   ```
   https://idlix-proxy.your-account.workers.dev
   ```

2. **Test with curl:**
   ```bash
   curl "https://idlix-proxy.your-account.workers.dev?url=https://jeniusplay.com/cdn/hls/262f89e9b8671281e5423d6fc260f5bf/master.m3u8"
   ```
   
   Atau buka di browser:
   ```
   https://idlix-proxy.your-account.workers.dev?url=https://jeniusplay.com/cdn/hls/262f89e9b8671281e5423d6fc260f5bf/master.m3u8
   ```

3. **Expected result:**
   - Anda akan lihat M3U8 playlist content
   - URLs sudah di-rewrite untuk go through worker

✅ **Worker berfungsi!**

---

### Step 5: Update Player

1. **Edit file `static/player.html`:**

2. **Find function `getProxyURL`:**
   ```javascript
   function getProxyURL(targetURL) {
       const encodedURL = encodeURIComponent(targetURL);
       return `/api/v1/proxy?url=${encodedURL}`;
   }
   ```

3. **Replace dengan:**
   ```javascript
   function getProxyURL(targetURL) {
       const encodedURL = encodeURIComponent(targetURL);
       // Replace dengan worker URL Anda
       return `https://idlix-proxy.your-account.workers.dev?url=${encodedURL}`;
   }
   ```

4. **Save file**

---

### Step 6: Test Streaming

1. **Start Go server:**
   ```bash
   cd IdlixDownloader/idlix-api
   go run cmd/api/*.go
   ```

2. **Open browser:**
   ```
   http://localhost:8080/player
   ```

3. **Test Crime 101:**
   - URL: https://tv12.idlixku.com/movie/crime-101-2026/
   - Click "Fetch Video Info"
   - Video should play!
   - Try switching quality options

4. **Check Network tab (F12):**
   - M3U8 requests should go to:
   ```
   https://idlix-proxy.your-account.workers.dev?url=...
   ```
   - TS segment requests also through worker
   - CORS headers present

✅ **Streaming works!**

---

## 🔧 Method 2: Deploy via Wrangler CLI (Advanced)

Untuk developer yang prefer command line.

### Prerequisites

- Node.js installed
- npm or yarn

### Steps

1. **Install Wrangler:**
   ```bash
   npm install -g wrangler
   ```

2. **Login to Cloudflare:**
   ```bash
   wrangler login
   ```
   - Browser akan terbuka
   - Login dengan akun Cloudflare Anda
   - Authorize Wrangler

3. **Navigate to worker directory:**
   ```bash
   cd IdlixDownloader/idlix-api/cloudflare-worker
   ```

4. **Edit wrangler.toml:**
   - Uncomment `account_id` line
   - Fill dengan Account ID Anda (dari dashboard)

5. **Deploy:**
   ```bash
   wrangler deploy
   ```

6. **Get URL:**
   ```bash
   wrangler deployments list
   ```

✅ **Worker deployed via CLI!**

---

## 📊 Monitoring

### View Metrics

1. **Di Cloudflare Dashboard:**
   - Workers & Pages → Your Worker
   - Click **"Metrics"**

2. **Metrics available:**
   - Request count (per hour/day)
   - Success rate (%)
   - CPU time (ms)
   - Errors count

### Real-time Logs

```bash
wrangler tail idlix-proxy
```

Output:
```
Listening for logs...
GET /proxy?url=... - 200 OK - 45ms
GET /proxy?url=... - 200 OK - 23ms
```

---

## 🌐 Optional: Custom Domain

Jika Anda punya domain, bisa setup custom domain.

### Prerequisites

- Domain registered di Cloudflare
- atau domain di registrar lain dengan DNS pointing ke Cloudflare

### Steps

1. **Add Custom Domain:**
   - Workers & Pages → Your Worker
   - Settings → Triggers
   - Custom Domains → **Add Custom Domain**

2. **Enter subdomain:**
   ```
   proxy.yourdomain.com
   ```

3. **Cloudflare auto-configures:**
   - DNS record created
   - SSL certificate issued
   - Worker bound to domain

4. **Wait 1-2 minutes**

5. **Test:**
   ```
   https://proxy.yourdomain.com?url=...
   ```

6. **Update player.html:**
   ```javascript
   function getProxyURL(targetURL) {
       const encodedURL = encodeURIComponent(targetURL);
       return `https://proxy.yourdomain.com?url=${encodedURL}`;
   }
   ```

✅ **Custom domain active!**

---

## 🔒 Security Configuration

### 1. Restrict Origins (Production)

Edit `worker.js`:

```javascript
const CONFIG = {
  // Change from '*' to your domain
  allowedOrigins: 'https://yourdomain.com',
  // ...
};
```

### 2. Add More Allowed Domains

Edit `isAllowedDomain()` function:

```javascript
const allowedDomains = [
  'jeniusplay.com',
  'tv12.idlixku.com',
  'idlix.com',
  'yourdomain.com',  // Add your domains
];
```

### 3. Enable Rate Limiting (Optional)

Via Cloudflare Dashboard:
- Security → WAF → Rate limiting rules
- Create rule for worker route

---

## 🐛 Troubleshooting

### Issue: Worker returns 404

**Solution:**
- Check worker URL is correct
- Check worker is deployed (green status in dashboard)
- Try redeploying

### Issue: CORS error in browser

**Solution:**
- Check `allowedOrigins` in CONFIG
- Should be `'*'` for development
- Or your exact domain for production

### Issue: Video won't play

**Solution:**
1. Open browser console (F12)
2. Check Network tab for errors
3. Check worker logs:
   ```bash
   wrangler tail idlix-proxy
   ```
4. Verify M3U8 URLs are being proxied

### Issue: Slow streaming

**Solution:**
- Check cache hit rate in Metrics
- Increase `cacheTTL` values
- Check if too many regions are being used

---

## 💰 Cost Estimate

### Free Tier (sudah cukup!)

**Limits:**
- 100,000 requests/day
- 1,000 requests/minute burst
- 10ms CPU time/request

**Estimated usage:**
- 1 movie = ~500 requests (M3U8 + TS segments)
- 100k requests = ~200 movies/day
- **MORE than enough for personal use!**

### Paid Plan ($5/month)

If you exceed free tier:
- 10 million requests/month
- 50ms CPU time/request
- Advanced features

---

## ✅ Post-Deployment Checklist

After successful deployment:

- [ ] Worker deployed and accessible
- [ ] Test URL works in browser
- [ ] player.html updated with worker URL
- [ ] Streaming works in player
- [ ] All quality options work
- [ ] CORS headers present
- [ ] Metrics showing requests
- [ ] (Optional) Custom domain configured
- [ ] (Optional) Production security enabled

---

## 📞 Support

**Cloudflare Resources:**
- Documentation: https://developers.cloudflare.com/workers/
- Community: https://community.cloudflare.com/
- Discord: https://discord.gg/cloudflaredev

**This Project:**
- Check README.md
- Review worker.js comments
- Test with sample URLs

---

## 🎉 Success!

Your Cloudflare Worker is now:
- ✅ Deployed globally (200+ locations)
- ✅ Handling streaming with CORS
- ✅ Caching for performance
- ✅ Auto-scaling
- ✅ Zero maintenance
- ✅ FREE!

Enjoy streaming! 🍿🎬
