# API Flow Test Plan

**Date:** 2026-03-17  
**Purpose:** Comprehensive testing of complete API flow

---

## 📋 Expected Flow (From Planning)

### Complete Video Info Flow:

```
User Request
    ↓
POST /api/v1/video/info
    ↓
[Step 1] GetVideoData(url)
    → Parse movie page
    → Extract: video_id, video_name, poster
    ↓
[Step 2] GetEmbedURL(video_id)
    → POST to wp-admin/admin-ajax.php
    → Get encrypted response
    → Decrypt with CryptoJS
    → Return: embed_url
    ↓
[Step 3] ExtractEmbedHash(embed_url)
    → Parse URL for hash
    → Return: hash
    ↓
[Step 4] GetVideoSource(hash)
    → POST to jeniusplay.com/player/index.php
    → Parse JSON response
    → Return: videoSource (MP4 URL)
    ↓
[Step 5] ConvertMP4ToM3U8(videoSource)
    → Replace .mp4 → .m3u8
    → Return: m3u8_url
    ↓
[Step 6] ParseMasterPlaylist(m3u8_url)
    → Download M3U8 file
    → Parse with grafov/m3u8
    → Extract variants (resolutions)
    → Return: variants[]
    ↓
[Step 7] GetSubtitleURL(hash) [Optional]
    → POST to jeniusplay.com
    → Regex extract subtitle URL
    → Return: subtitle_url
    ↓
Response JSON with complete video info
```

---

## 🎯 Test Cases

### Test 1: Health Check
**Endpoint:** `GET /api/v1/health`  
**Expected:**
```json
{
  "status": "ok",
  "version": "1.0.0",
  "message": "IDLIX API is running"
}
```

---

### Test 2: Featured Movies
**Endpoint:** `GET /api/v1/featured`  
**Expected:**
```json
{
  "status": true,
  "message": "Featured movies retrieved successfully",
  "data": {
    "movies": [
      {
        "url": "https://tv12.idlixku.com/movie/...",
        "title": "Movie Title",
        "year": "2024",
        "type": "movie",
        "poster": "https://..."
      }
    ]
  }
}
```

**Validation:**
- ✅ Status code: 200
- ✅ Array of movies (expect 10-20)
- ✅ Each movie has: url, title, poster
- ✅ No TV series included

---

### Test 3: Complete Video Info (CRITICAL)
**Endpoint:** `POST /api/v1/video/info`  
**Request:**
```json
{
  "url": "https://tv12.idlixku.com/movie/crime-101-2026/"
}
```

**Expected Response:**
```json
{
  "status": true,
  "message": "Video info retrieved successfully",
  "data": {
    "video_id": "163426",
    "video_name": "Crime 101 (2026)",
    "poster": "https://image.tmdb.org/...",
    "embed_url": "https://jeniusplay.com/video/...",
    "m3u8_url": "https://jeniusplay.com/cdn/hls/.../master.m3u8",
    "is_variant_playlist": true,
    "variants": [
      {
        "id": "0",
        "resolution": "1920x804",
        "bandwidth": 1510000,
        "uri": "https://jeniusplay.com/hls/..."
      }
    ],
    "subtitle": {
      "available": true,
      "url": "https://...",
      "format": "vtt"
    }
  }
}
```

**Step-by-Step Validation:**

#### Step 1: Video Data ✅
- ✅ video_id exists and numeric
- ✅ video_name matches movie title
- ✅ poster URL valid

#### Step 2: Embed URL ✅
- ✅ embed_url decrypted successfully
- ✅ Format: https://jeniusplay.com/video/{hash}

#### Step 3: Embed Hash ✅
- ✅ Hash extracted from URL
- ✅ Hash is alphanumeric (32 chars)

#### Step 4: Video Source ✅
- ✅ JeniusPlay API returns videoSource
- ✅ videoSource is MP4 URL

#### Step 5: M3U8 URL ✅
- ✅ M3U8 URL created from MP4
- ✅ Format: .../master.m3u8

#### Step 6: Variants ✅
- ✅ M3U8 parsed successfully
- ✅ is_variant_playlist = true
- ✅ variants array has 2+ items
- ✅ Each variant has: id, resolution, bandwidth, uri
- ✅ Resolutions sorted (highest first)

#### Step 7: Subtitle ✅
- ✅ subtitle.available = boolean
- ✅ If available, URL is valid
- ✅ Format is "vtt"

---

## 🚨 Error Cases to Test

### Test 4: Invalid URL
**Request:**
```json
{"url": "https://invalid-site.com/movie/test/"}
```
**Expected:**
```json
{
  "status": false,
  "message": "Failed to get video info",
  "error": {
    "code": "VIDEO_INFO_ERROR",
    "details": "..."
  }
}
```

### Test 5: Missing URL
**Request:**
```json
{}
```
**Expected:**
```json
{
  "status": false,
  "message": "Invalid request body",
  "error": {
    "code": "INVALID_REQUEST",
    "details": "..."
  }
}
```

### Test 6: Non-existent Movie
**Request:**
```json
{"url": "https://tv12.idlixku.com/movie/does-not-exist-12345/"}
```
**Expected:**
```json
{
  "status": false,
  "message": "Failed to get video info",
  "error": {...}
}
```

---

## 🔍 Integration Points to Verify

### 1. Cloudflare Bypass
- ✅ All requests pass Cloudflare
- ✅ No "connection reset by peer"
- ✅ No 403 Forbidden
- ✅ Client Hints headers working

### 2. AES Decryption
- ✅ Embed URL decrypted correctly
- ✅ Passphrase generation working
- ✅ No padding errors

### 3. JeniusPlay API
- ✅ POST request successful
- ✅ videoSource returned
- ✅ Subtitle extraction working

### 4. M3U8 Parsing
- ✅ M3U8 file downloaded
- ✅ Playlist parsed
- ✅ Variants extracted
- ✅ URIs made absolute

### 5. Response Format
- ✅ JSON structure correct
- ✅ All fields present
- ✅ Data types correct
- ✅ No null for required fields

---

## 📊 Performance Metrics to Check

| Metric | Target | Acceptable |
|--------|--------|------------|
| Health endpoint | < 50ms | < 100ms |
| Featured endpoint | < 1s | < 2s |
| Video info endpoint | < 5s | < 10s |
| Memory per request | < 50MB | < 100MB |

---

## ✅ Success Criteria

**API is considered fully functional if:**

1. ✅ All 3 endpoints return 200 OK
2. ✅ Featured returns 10+ movies
3. ✅ Video info completes all 7 steps
4. ✅ Variants array populated
5. ✅ Subtitle extracted (if available)
6. ✅ No errors in any step
7. ✅ Response format matches spec
8. ✅ Error handling works correctly

---

## 🧪 Test Execution Plan

### Phase 1: Basic Endpoints (2 minutes)
1. Test health endpoint
2. Test featured endpoint

### Phase 2: Complete Flow (5 minutes)
3. Test video info with valid URL
4. Verify all 7 steps completed
5. Check response structure

### Phase 3: Edge Cases (3 minutes)
6. Test with invalid URL
7. Test with missing data
8. Test error responses

### Phase 4: Integration (5 minutes)
9. Test multiple requests
10. Check for memory leaks
11. Verify Cloudflare bypass stable

**Total Estimated Time:** 15 minutes

---

**Next Action:** Execute tests
