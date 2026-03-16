# Cloudflare Bypass Analysis: Python vs Go

**Date:** 2026-03-17  
**Purpose:** Detailed comparison of Cloudflare bypass implementation

---

## 🔍 Python Implementation Analysis

### curl_cffi Configuration (Line 52-56)

```python
self.request = cffi_requests.Session(
    impersonate=random.choice(["chrome124", "chrome119", "chrome104"]),
    headers=self.BASE_STATIC_HEADERS,
    debug=False,
)
```

### Key Features:

1. **Browser Impersonation:**
   - Random selection: chrome124, chrome119, chrome104
   - curl_cffi provides full Chrome browser fingerprint

2. **Static Headers (Lines 27-42):**
```python
BASE_STATIC_HEADERS = {
    "Host": "tv12.idlixku.com",
    "Connection": "keep-alive",
    "sec-ch-ua": 'Not)A;Brand;v=99, Google Chrome;v=127, Chromium;v=127',
    "sec-ch-ua-mobile": "?0",
    "sec-ch-ua-platform": "Windows",
    "Upgrade-Insecure-Requests": "1",
    "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36",
    "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
    "Sec-Fetch-Site": "same-origin",
    "Sec-Fetch-Mode": "navigate",
    "Sec-Fetch-User": "?1",
    "Sec-Fetch-Dest": "document",
    "Referer": "https://tv12.idlixku.com/",
    "Accept-Language": "en-US,en;q=0.9,id;q=0.8"
}
```

3. **Critical Headers:**
   - ✅ `sec-ch-ua` - Client hints (Chrome version)
   - ✅ `sec-ch-ua-mobile` - Mobile detection
   - ✅ `sec-ch-ua-platform` - OS platform
   - ✅ `Sec-Fetch-*` - Security headers
   - ✅ `User-Agent` - Browser identification

---

## 🔍 Go Implementation Analysis

### tls-client Configuration

```go
options := []tls_client.HttpClientOption{
    tls_client.WithTimeoutSeconds(int(config.Timeout.Seconds())),
    tls_client.WithClientProfile(profiles.Chrome_124),
    tls_client.WithRandomTLSExtensionOrder(),
    tls_client.WithNotFollowRedirects(),
}

client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
```

### Current Headers (doGet method):

```go
req.Header.Set("User-Agent", c.getRandomUserAgent())
req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
req.Header.Set("Accept-Language", "en-US,en;q=0.9,id;q=0.8")
req.Header.Set("Accept-Encoding", "gzip, deflate, br")
req.Header.Set("Connection", "keep-alive")
req.Header.Set("Upgrade-Insecure-Requests", "1")
req.Header.Set("Sec-Fetch-Dest", "document")
req.Header.Set("Sec-Fetch-Mode", "navigate")
req.Header.Set("Sec-Fetch-Site", "none")
req.Header.Set("Sec-Fetch-User", "?1")
req.Header.Set("Cache-Control", "max-age=0")
```

---

## ⚠️ MISSING CRITICAL HEADERS IN GO

### 1. **sec-ch-ua** Headers (CRITICAL!)

**Python has:**
```python
"sec-ch-ua": 'Not)A;Brand;v=99, Google Chrome;v=127, Chromium;v=127'
"sec-ch-ua-mobile": "?0"
"sec-ch-ua-platform": "Windows"
```

**Go missing:** ❌ These are **CRITICAL** for Cloudflare detection!

### 2. **Host Header**

**Python has:**
```python
"Host": "tv12.idlixku.com"
```

**Go missing:** ❌ Not set explicitly

### 3. **Referer Header**

**Python has:**
```python
"Referer": "https://tv12.idlixku.com/"
```

**Go missing:** ❌ Not set for GET requests

---

## 🔧 Issues Found

| Issue | Severity | Impact |
|-------|----------|--------|
| Missing `sec-ch-ua` headers | 🔴 **CRITICAL** | Cloudflare can detect bot |
| Missing `Host` header | 🟡 **MEDIUM** | May trigger suspicion |
| Missing `Referer` for same-origin | 🟡 **MEDIUM** | Breaks navigation flow |
| `Sec-Fetch-Site` = "none" | 🟡 **MEDIUM** | Should be "same-origin" |
| User-Agent not matching Chrome 127 | 🟠 **LOW** | Minor inconsistency |

---

## ✅ What Go Does Correctly

1. ✅ TLS fingerprinting (Chrome 124 profile)
2. ✅ Random TLS extension order
3. ✅ Proper Accept headers
4. ✅ Sec-Fetch-* headers (but wrong values)
5. ✅ Accept-Encoding with brotli
6. ✅ Connection keep-alive

---

## 🚨 ROOT CAUSE ANALYSIS

### Why Cloudflare Blocks Go Implementation:

1. **Missing Client Hints** (`sec-ch-ua*`)
   - Modern browsers ALWAYS send these
   - Cloudflare checks for their presence
   - Missing = immediate bot detection

2. **Wrong `Sec-Fetch-Site` Value**
   - Go: "none" (indicates direct navigation)
   - Python: "same-origin" (indicates navigation within site)
   - For subsequent requests, should be "same-origin"

3. **No Referer Chain**
   - Real browsers maintain referer chain
   - Go doesn't set Referer for same-site navigation

---

## 📋 Required Fixes

### Priority 1: Add Client Hints (CRITICAL)

```go
req.Header.Set("sec-ch-ua", `"Not)A;Brand";v="99", "Google Chrome";v="127", "Chromium";v="127"`)
req.Header.Set("sec-ch-ua-mobile", "?0")
req.Header.Set("sec-ch-ua-platform", `"Windows"`)
```

### Priority 2: Fix Sec-Fetch-Site

```go
// For homepage (first request)
req.Header.Set("Sec-Fetch-Site", "none")

// For subsequent requests (same-origin navigation)
req.Header.Set("Sec-Fetch-Site", "same-origin")
```

### Priority 3: Add Referer

```go
// For same-origin requests
req.Header.Set("Referer", c.baseURL)
```

### Priority 4: Add Host Header

```go
// Extract from URL
parsedURL, _ := url.Parse(targetURL)
req.Header.Set("Host", parsedURL.Host)
```

---

## 🧪 Test Plan

### Test 1: Without Fixes
- Expected: Connection reset / 403 Forbidden

### Test 2: With Client Hints Only
- Expected: Should improve, may still block

### Test 3: With All Fixes
- Expected: Should bypass successfully

---

## 📊 Comparison Matrix

| Feature | Python (curl_cffi) | Go (tls-client) | Status |
|---------|-------------------|-----------------|--------|
| TLS Fingerprinting | ✅ Chrome 124/119/104 | ✅ Chrome 124 | ✅ OK |
| Client Hints | ✅ sec-ch-ua* | ❌ Missing | 🔴 CRITICAL |
| User-Agent | ✅ Chrome 127 | ✅ Random | ✅ OK |
| Sec-Fetch-* | ✅ Correct values | ⚠️ Wrong values | 🟡 FIX |
| Host Header | ✅ Set | ❌ Not set | 🟡 FIX |
| Referer | ✅ Set | ❌ Not set | 🟡 FIX |
| Accept Headers | ✅ Complete | ✅ Complete | ✅ OK |
| Connection | ✅ keep-alive | ✅ keep-alive | ✅ OK |

---

## 🎯 Conclusion

**Current Status:** ❌ **INCOMPLETE BYPASS**

**Why it's failing:**
1. 🔴 Missing critical `sec-ch-ua` headers (main cause)
2. 🟡 Wrong `Sec-Fetch-Site` value
3. 🟡 Missing `Referer` chain
4. 🟡 Missing explicit `Host` header

**Fix Difficulty:** ⭐⭐ (Easy - just add headers)

**Estimated Time:** 5-10 minutes

---

**Next Action:** Apply fixes to `httpclient.go`
