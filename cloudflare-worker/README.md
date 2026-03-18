# Cloudflare Worker - IDLIX Streaming Proxy

Proxy untuk streaming M3U8/HLS melalui Cloudflare Workers untuk bypass CORS restrictions.

## 🌟 Features

- ✅ CORS headers otomatis
- ✅ M3U8 playlist URL rewriting
- ✅ Caching dengan Cloudflare CDN
- ✅ Global edge deployment (200+ locations)
- ✅ Free tier: 100,000 requests/day
- ✅ Zero maintenance
- ✅ Auto-scaling

## 📋 Prerequisites

1. Akun Cloudflare (gratis): https://dash.cloudflare.com/sign-up
2. Domain (opsional, untuk custom domain)

## 🚀 Deployment Steps

### Step 1: Setup Cloudflare Account

1. Buka: https://dash.cloudflare.com/sign-up
2. Daftar dengan email Anda
3. Verifikasi email

### Step 2: Create Worker

1. Login ke Cloudflare Dashboard
2. Klik **Workers & Pages** di sidebar kiri
3. Klik **Create Application**
4. Klik **Create Worker**
5. Beri nama worker, misal: `idlix-proxy`
6. Klik **Deploy**

### Step 3: Edit Worker Code

1. Setelah deploy, klik **Edit Code**
2. Hapus semua code default
3. Copy semua isi dari `worker.js`
4. Paste ke code editor
5. Klik **Save and Deploy**

### Step 4: Get Worker URL

Worker URL Anda akan seperti:
```
https://idlix-proxy.your-account.workers.dev
```

Contoh penggunaan:
```
https://idlix-proxy.your-account.workers.dev?url=https://jeniusplay.com/cdn/hls/xxx/master.m3u8
```

### Step 5: Update Player

Edit `static/player.html`, ubah fungsi `getProxyURL()`:

```javascript
// Before (menggunakan Go proxy)
function getProxyURL(targetURL) {
    const encodedURL = encodeURIComponent(targetURL);
    return `/api/v1/proxy?url=${encodedURL}`;
}

// After (menggunakan Cloudflare Worker)
function getProxyURL(targetURL) {
    const encodedURL = encodeURIComponent(targetURL);
    return `https://idlix-proxy.your-account.workers.dev?url=${encodedURL}`;
}
```

### Step 6: Test

1. Buka http://localhost:8080/player
2. Fetch video info
3. Play video - seharusnya streaming melalui Cloudflare Worker

## 🔧 Configuration

Edit bagian `CONFIG` di `worker.js`:

```javascript
const CONFIG = {
  // Allowed origins
  allowedOrigins: '*',  // Untuk production, ganti dengan domain spesifik
  
  // Cache TTL
  cacheTTL: {
    m3u8: 60,    // 1 menit untuk playlist
    ts: 3600,    // 1 jam untuk TS segments
  },
  
  // Headers
  upstreamHeaders: {
    'User-Agent': 'Mozilla/5.0...',
    'Referer': 'https://tv12.idlixku.com/',
  }
};
```

## 🌐 Custom Domain (Opsional)

Jika punya domain, bisa setup custom domain:

1. Di Cloudflare Dashboard, pilih worker Anda
2. Klik **Settings** → **Triggers**
3. Klik **Add Custom Domain**
4. Masukkan subdomain, misal: `proxy.yourdomain.com`
5. Cloudflare akan auto-configure DNS

Hasilnya:
```
https://proxy.yourdomain.com?url=...
```

## 📊 Monitoring

### View Metrics

1. Buka Cloudflare Dashboard
2. Pilih worker Anda
3. Klik **Metrics** untuk melihat:
   - Request count
   - Success rate
   - CPU time
   - Errors

### View Logs (Real-time)

```bash
# Install Wrangler CLI
npm install -g wrangler

# Login
wrangler login

# Tail logs
wrangler tail idlix-proxy
```

## 💰 Pricing

### Free Tier (sudah cukup untuk personal use)
- ✅ 100,000 requests/day
- ✅ 10ms CPU time per request
- ✅ Unlimited bandwidth
- ✅ Global deployment

### Paid Plan ($5/month)
- 10 million requests/month
- 50ms CPU time per request
- Advanced features

## 🔒 Security Best Practices

1. **Restrict Origins** (Production):
   ```javascript
   allowedOrigins: 'https://yourdomain.com'
   ```

2. **Domain Whitelist**:
   Already implemented - only allows jeniusplay.com, idlixku.com

3. **Rate Limiting**:
   Cloudflare automatically handles this

4. **DDoS Protection**:
   Built-in with Cloudflare

## 🐛 Troubleshooting

### Worker not responding
- Check deployment status di dashboard
- Check logs dengan `wrangler tail`

### CORS errors
- Verify `allowedOrigins` in CONFIG
- Check browser console for exact error

### Slow streaming
- Check cache hit rate di Metrics
- Adjust `cacheTTL` values

### 403 Forbidden
- URL domain tidak di whitelist
- Tambahkan domain ke `isAllowedDomain()` function

## 📈 Performance Tips

1. **Enable Caching**:
   Already enabled in worker.js

2. **Use Custom Domain**:
   Better for DNS resolution

3. **Adjust Cache TTL**:
   - M3U8: 30-60 seconds (dynamic)
   - TS: 1 hour+ (static)

4. **Monitor Metrics**:
   Check cache hit rate regularly

## 🔄 Updates

Update worker code:

1. Edit `worker.js`
2. Copy new code
3. Paste di Cloudflare editor
4. Click **Save and Deploy**

Changes are deployed globally in ~30 seconds.

## 📚 Resources

- [Cloudflare Workers Docs](https://developers.cloudflare.com/workers/)
- [Wrangler CLI](https://developers.cloudflare.com/workers/wrangler/)
- [Workers Examples](https://developers.cloudflare.com/workers/examples/)

## ✅ Checklist

Sebelum production:

- [ ] Deploy worker ke Cloudflare
- [ ] Test dengan berbagai video
- [ ] Update `allowedOrigins` untuk production
- [ ] (Optional) Setup custom domain
- [ ] Monitor metrics untuk 1-2 hari
- [ ] Setup alerts untuk errors
- [ ] Dokumentasikan worker URL untuk tim

## 🎯 Next Steps

1. Deploy worker ✅
2. Test thoroughly ✅
3. Update player.html ✅
4. (Optional) Remove Go proxy endpoint
5. Monitor dan optimize

---

**Status**: ✅ Production Ready
**Deployment Time**: ~5 minutes
**Maintenance**: Zero (serverless)
