# 🎯 Subtitle Enhancement Implementation Plan

**Date:** 2026-03-18  
**Purpose:** Add subtitle download, format conversion, and search features

---

## 📋 Overview

Three new features to enhance subtitle functionality:
1. **Subtitle Download Endpoint** - Proxy and download subtitle files
2. **Format Conversion** - Convert between SRT ↔ VTT formats
3. **Search by Language** - Filter subtitles by language preference

---

## 🎯 Feature 1: Subtitle Download Endpoint

### Purpose
Provide a proxy endpoint to download subtitle files with CORS support and optional format conversion.

### API Design

**Endpoint:** `GET /api/v1/subtitle/download`

**Query Parameters:**
- `url` (required) - Subtitle URL to download
- `format` (optional) - Target format: `srt` or `vtt` (default: keep original)
- `filename` (optional) - Custom filename for download

**Response:**
- Content-Type: `text/plain` or `text/vtt`
- Content-Disposition: `attachment; filename="subtitle.srt"`
- Body: Subtitle file content

**Example:**
```bash
# Download as-is
GET /api/v1/subtitle/download?url=https://g5.wiseacademia.asia/r/xyz.jpg

# Download and convert to VTT
GET /api/v1/subtitle/download?url=https://g5.wiseacademia.asia/r/xyz.jpg&format=vtt

# Custom filename
GET /api/v1/subtitle/download?url=https://g5.wiseacademia.asia/r/xyz.jpg&filename=crime-101-bahasa.srt
```

### Implementation Details

**New Handler Function:**
```go
// internal/handlers/subtitle.go

type SubtitleHandler struct {
    converter *utils.SubtitleConverter
}

func (h *SubtitleHandler) DownloadSubtitle(c *gin.Context) {
    // 1. Get URL from query
    // 2. Fetch subtitle content
    // 3. Detect current format
    // 4. Convert if format param specified
    // 5. Set download headers
    // 6. Return content
}
```

**Features:**
- ✅ CORS support for browser downloads
- ✅ Automatic format detection
- ✅ Optional format conversion
- ✅ Custom filename support
- ✅ Error handling for invalid URLs

---

## 🔄 Feature 2: Format Conversion (SRT ↔ VTT)

### Purpose
Convert subtitle files between SRT (SubRip) and VTT (WebVTT) formats.

### Format Specifications

#### SRT Format (SubRip Text)
```srt
1
00:00:04,515 --> 00:00:31,615
<b>Alih Bahasa: CemonK</b>

2
00:00:32,039 --> 00:00:34,050
Tarik napas dalam-dalam.
```

**Characteristics:**
- Sequence numbers (1, 2, 3...)
- Timestamps: `HH:MM:SS,mmm --> HH:MM:SS,mmm` (comma for milliseconds)
- Blank line separator
- Can contain HTML tags

#### VTT Format (WebVTT)
```vtt
WEBVTT

00:00:04.515 --> 00:00:31.615
<b>Alih Bahasa: CemonK</b>

00:00:32.039 --> 00:00:34.050
Tarik napas dalam-dalam.
```

**Characteristics:**
- Header: `WEBVTT`
- No sequence numbers
- Timestamps: `HH:MM:SS.mmm --> HH:MM:SS.mmm` (dot for milliseconds)
- Blank line separator
- Can contain cues and styling

### Implementation Details

**New Utility Package:**
```go
// internal/utils/subtitle_converter.go

type SubtitleConverter struct{}

// Convert subtitle between formats
func (c *SubtitleConverter) Convert(content string, fromFormat, toFormat string) (string, error)

// SRT to VTT
func (c *SubtitleConverter) SRTToVTT(srtContent string) (string, error)

// VTT to SRT
func (c *SubtitleConverter) VTTToSRT(vttContent string) (string, error)

// Detect format from content
func (c *SubtitleConverter) DetectFormat(content string) string

// Validate subtitle format
func (c *SubtitleConverter) Validate(content string, format string) error
```

**Conversion Logic:**

**SRT → VTT:**
1. Add `WEBVTT` header
2. Remove sequence numbers
3. Replace `,` with `.` in timestamps
4. Keep blank line separators
5. Preserve cue text and tags

**VTT → SRT:**
1. Remove `WEBVTT` header
2. Add sequence numbers
3. Replace `.` with `,` in timestamps
4. Keep blank line separators
5. Preserve cue text and tags

**Edge Cases to Handle:**
- Missing header in VTT
- Invalid timestamp format
- Empty cues
- Special characters
- Multiple blank lines

---

## 🔍 Feature 3: Search Subtitle by Language

### Purpose
Filter and search subtitle tracks by language preference.

### API Design

**Endpoint:** `GET /api/v1/subtitle/search`

**Query Parameters:**
- `url` (required) - Video URL
- `language` (optional) - Language filter (e.g., "Bahasa", "English")
- `format` (optional) - Format preference: `srt` or `vtt`

**Response:**
```json
{
  "status": true,
  "message": "Subtitles found",
  "data": {
    "video_id": "163426",
    "video_name": "Crime 101 (2026)",
    "subtitles": [
      {
        "language": "Bahasa",
        "url": "https://g5.wiseacademia.asia/r/xyz.jpg",
        "format": "srt",
        "download_url": "/api/v1/subtitle/download?url=https://..."
      }
    ],
    "total": 1,
    "filtered": true
  }
}
```

**Example Requests:**
```bash
# Get all subtitles for a video
GET /api/v1/subtitle/search?url=https://tv12.idlixku.com/movie/crime-101-2026/

# Filter by language
GET /api/v1/subtitle/search?url=https://tv12.idlixku.com/movie/crime-101-2026/&language=Bahasa

# Filter by language (case-insensitive, partial match)
GET /api/v1/subtitle/search?url=https://tv12.idlixku.com/movie/crime-101-2026/&language=eng
```

### Implementation Details

**New Handler Function:**
```go
// internal/handlers/subtitle.go

func (h *SubtitleHandler) SearchSubtitles(c *gin.Context) {
    // 1. Get video URL
    // 2. Get video info (includes all subtitle tracks)
    // 3. Filter by language if specified
    // 4. Add download URLs
    // 5. Return filtered results
}
```

**Search Features:**
- ✅ Case-insensitive search
- ✅ Partial matching (e.g., "eng" matches "English")
- ✅ Multiple results support
- ✅ Auto-generate download URLs
- ✅ Return total count

**Language Matching:**
```go
func matchLanguage(trackLang, searchLang string) bool {
    // Convert to lowercase
    // Partial match support
    // Common language aliases (e.g., "indo" → "Bahasa Indonesia")
}
```

---

## 📁 File Structure

### New Files to Create:

```
internal/
├── handlers/
│   └── subtitle.go          (NEW) - Subtitle-specific handlers
│
├── utils/
│   ├── subtitle_converter.go    (NEW) - Format conversion
│   └── subtitle_converter_test.go (NEW) - Unit tests
│
└── models/
    └── subtitle.go          (UPDATE) - Add new request/response models
```

### Files to Modify:

```
cmd/api/main.go              - Add new routes
docs/                        - Regenerate Swagger
```

---

## 🎯 New Models

### SubtitleDownloadRequest
```go
type SubtitleDownloadRequest struct {
    URL      string `form:"url" binding:"required"`
    Format   string `form:"format"`
    Filename string `form:"filename"`
}
```

### SubtitleSearchRequest
```go
type SubtitleSearchRequest struct {
    URL      string `form:"url" binding:"required"`
    Language string `form:"language"`
    Format   string `form:"format"`
}
```

### SubtitleSearchResponse
```go
type SubtitleSearchResponse struct {
    VideoID    string              `json:"video_id"`
    VideoName  string              `json:"video_name"`
    Subtitles  []SubtitleTrackInfo `json:"subtitles"`
    Total      int                 `json:"total"`
    Filtered   bool                `json:"filtered"`
}

type SubtitleTrackInfo struct {
    Language    string `json:"language"`
    URL         string `json:"url"`
    Format      string `json:"format"`
    DownloadURL string `json:"download_url"`
}
```

---

## 🛣️ API Routes

### New Routes to Add:

```go
// Subtitle endpoints
subtitleGroup := v1.Group("/subtitle")
{
    // Download subtitle with optional conversion
    subtitleGroup.GET("/download", subtitleHandler.DownloadSubtitle)
    
    // Search/filter subtitles
    subtitleGroup.GET("/search", subtitleHandler.SearchSubtitles)
}
```

### Complete API Structure:

```
/api/v1/
├── health              GET     - Health check
├── featured            GET     - Featured movies
├── video/
│   └── info            POST    - Video info
├── proxy               GET     - M3U8/TS proxy
└── subtitle/           (NEW)
    ├── download        GET     - Download subtitle
    └── search          GET     - Search subtitles
```

---

## 🧪 Testing Plan

### 1. Unit Tests

**Subtitle Converter:**
```go
// internal/utils/subtitle_converter_test.go

func TestSRTToVTT(t *testing.T) {
    // Test basic conversion
    // Test with HTML tags
    // Test with multiple cues
    // Test edge cases
}

func TestVTTToSRT(t *testing.T) {
    // Similar tests
}

func TestDetectFormat(t *testing.T) {
    // Test SRT detection
    // Test VTT detection
    // Test invalid format
}
```

### 2. Integration Tests

**Download Endpoint:**
- ✅ Download original format
- ✅ Download with SRT→VTT conversion
- ✅ Download with VTT→SRT conversion
- ✅ Custom filename
- ✅ Invalid URL handling
- ✅ CORS headers present

**Search Endpoint:**
- ✅ Get all subtitles
- ✅ Filter by exact language
- ✅ Filter by partial language
- ✅ Case-insensitive search
- ✅ No results found
- ✅ Invalid video URL

### 3. Manual Testing

**Test Cases:**
```bash
# 1. Download Bahasa subtitle as SRT
curl "http://localhost:8080/api/v1/subtitle/download?url=...&format=srt" -o subtitle.srt

# 2. Download and convert to VTT
curl "http://localhost:8080/api/v1/subtitle/download?url=...&format=vtt" -o subtitle.vtt

# 3. Search for English subtitles
curl "http://localhost:8080/api/v1/subtitle/search?url=...&language=english"

# 4. Get all subtitles for a movie
curl "http://localhost:8080/api/v1/subtitle/search?url=..."
```

---

## 📊 Implementation Steps

### Step 1: Subtitle Converter Utility
1. Create `internal/utils/subtitle_converter.go`
2. Implement `SRTToVTT()` function
3. Implement `VTTToSRT()` function
4. Implement `DetectFormat()` function
5. Write unit tests
6. Test with real subtitle files

### Step 2: Subtitle Handler
1. Create `internal/handlers/subtitle.go`
2. Implement `DownloadSubtitle()` handler
3. Implement `SearchSubtitles()` handler
4. Add new models to `internal/models/`
5. Test handlers manually

### Step 3: Routes & Integration
1. Add routes to `cmd/api/main.go`
2. Update Swagger annotations
3. Regenerate Swagger docs
4. Integration testing

### Step 4: Documentation
1. Update `USAGE.md` with new endpoints
2. Create examples in documentation
3. Add testing guide

---

## ⚠️ Considerations

### Security
- ✅ Validate subtitle URLs (prevent SSRF)
- ✅ Limit file size (prevent DoS)
- ✅ Sanitize filenames
- ✅ Add rate limiting

### Performance
- ✅ Cache converted subtitles (optional)
- ✅ Stream large files (don't load all in memory)
- ✅ Set timeouts for external requests

### Error Handling
- ✅ Invalid URL format
- ✅ Unreachable subtitle URL
- ✅ Invalid subtitle format
- ✅ Conversion errors
- ✅ Empty results

---

## 📈 Expected Benefits

### For Users
- 🎯 Direct subtitle download
- 🎯 Format compatibility (any player)
- 🎯 Easy language selection
- 🎯 No CORS issues

### For Developers
- 🎯 Clean API design
- 🎯 Reusable converter utility
- 🎯 Well-documented endpoints
- 🎯 Comprehensive tests

---

## 🎯 Success Criteria

- ✅ All 3 endpoints working
- ✅ Format conversion accurate (SRT ↔ VTT)
- ✅ Search returns correct results
- ✅ Unit tests passing
- ✅ Integration tests passing
- ✅ Swagger documentation updated
- ✅ No breaking changes to existing API

---

**Ready to implement?** Let's start with Step 1: Subtitle Converter Utility! 🚀
