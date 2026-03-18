/**
 * Cloudflare Worker - IDLIX Streaming Proxy
 * 
 * This worker acts as a CORS proxy for streaming M3U8 playlists and TS segments
 * from JeniusPlay to bypass browser CORS restrictions.
 * 
 * Features:
 * - CORS headers for all responses
 * - M3U8 playlist URL rewriting
 * - Caching for better performance
 * - Error handling
 * 
 * Deploy: https://dash.cloudflare.com/workers
 */

// Configuration
const CONFIG = {
  // Allowed origins (use '*' for development, specific domain for production)
  allowedOrigins: '*',
  
  // Cache TTL (Time To Live) in seconds
  cacheTTL: {
    m3u8: 60,      // 1 minute for playlists
    ts: 3600,      // 1 hour for TS segments
  },
  
  // Headers to forward to upstream
  upstreamHeaders: {
    'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
    'Accept': '*/*',
    'Referer': 'https://tv12.idlixku.com/',
  }
};

/**
 * Main worker handler
 */
addEventListener('fetch', event => {
  event.respondWith(handleRequest(event.request));
});

/**
 * Handle incoming requests
 */
async function handleRequest(request) {
  // Handle CORS preflight
  if (request.method === 'OPTIONS') {
    return handleCORSPreflight();
  }

  // Only allow GET requests
  if (request.method !== 'GET') {
    return createErrorResponse('Method not allowed', 405);
  }

  // Parse URL and get target parameter
  const url = new URL(request.url);
  const targetURL = url.searchParams.get('url');

  // Validate target URL
  if (!targetURL) {
    return createErrorResponse('Missing url parameter. Usage: ?url=<target_url>', 400);
  }

  // Validate target URL is from allowed domain
  if (!isAllowedDomain(targetURL)) {
    return createErrorResponse('Target domain not allowed', 403);
  }

  try {
    // Fetch from upstream
    const response = await fetchUpstream(targetURL);

    // Check if response is successful
    if (!response.ok) {
      return createErrorResponse(`Upstream error: ${response.status} ${response.statusText}`, response.status);
    }

    // Process based on content type
    const contentType = response.headers.get('Content-Type') || '';
    const isM3U8 = contentType.includes('mpegURL') || targetURL.endsWith('.m3u8');

    if (isM3U8) {
      // Process M3U8 playlist
      return await processM3U8(response, targetURL, url.origin);
    } else {
      // Pass through (TS segments, etc.)
      return createProxyResponse(response);
    }

  } catch (error) {
    console.error('Proxy error:', error);
    return createErrorResponse(`Proxy error: ${error.message}`, 500);
  }
}

/**
 * Fetch from upstream server
 */
async function fetchUpstream(targetURL) {
  const cacheKey = new Request(targetURL);
  const cache = caches.default;

  // Check cache first
  let response = await cache.match(cacheKey);
  
  if (!response) {
    // Not in cache, fetch from upstream
    response = await fetch(targetURL, {
      headers: CONFIG.upstreamHeaders,
      cf: {
        // Cloudflare-specific options
        cacheTtl: targetURL.endsWith('.m3u8') ? CONFIG.cacheTTL.m3u8 : CONFIG.cacheTTL.ts,
        cacheEverything: true,
      }
    });

    // Clone response for caching
    const clonedResponse = response.clone();
    
    // Cache the response
    const cacheResponse = new Response(clonedResponse.body, {
      status: clonedResponse.status,
      statusText: clonedResponse.statusText,
      headers: clonedResponse.headers
    });
    
    event.waitUntil(cache.put(cacheKey, cacheResponse));
  }

  return response;
}

/**
 * Process M3U8 playlist - rewrite URLs to go through this worker
 */
async function processM3U8(response, originalURL, workerOrigin) {
  // Read playlist content
  const content = await response.text();
  
  // Rewrite URLs in playlist
  const rewrittenContent = rewriteM3U8URLs(content, originalURL, workerOrigin);
  
  // Create new response with rewritten content
  return new Response(rewrittenContent, {
    status: response.status,
    statusText: response.statusText,
    headers: addCORSHeaders(new Headers({
      'Content-Type': 'application/x-mpegURL',
      'Cache-Control': `public, max-age=${CONFIG.cacheTTL.m3u8}`,
    }))
  });
}

/**
 * Rewrite URLs in M3U8 playlist to go through this worker
 */
function rewriteM3U8URLs(content, originalURL, workerOrigin) {
  const lines = content.split('\n');
  const baseURL = getBaseURL(originalURL);
  
  const rewrittenLines = lines.map(line => {
    const trimmed = line.trim();
    
    // Skip empty lines and comments
    if (!trimmed || trimmed.startsWith('#')) {
      return line;
    }
    
    // This is a URL line
    let absoluteURL;
    
    if (trimmed.startsWith('http://') || trimmed.startsWith('https://')) {
      // Already absolute URL
      absoluteURL = trimmed;
    } else if (trimmed.startsWith('/')) {
      // Absolute path
      const urlObj = new URL(originalURL);
      absoluteURL = `${urlObj.protocol}//${urlObj.host}${trimmed}`;
    } else {
      // Relative path
      absoluteURL = baseURL + trimmed;
    }
    
    // Rewrite to go through this worker
    return `${workerOrigin}?url=${encodeURIComponent(absoluteURL)}`;
  });
  
  return rewrittenLines.join('\n');
}

/**
 * Get base URL from full URL (everything before the filename)
 */
function getBaseURL(url) {
  const lastSlash = url.lastIndexOf('/');
  return url.substring(0, lastSlash + 1);
}

/**
 * Create proxy response with CORS headers
 */
function createProxyResponse(response) {
  const headers = new Headers(response.headers);
  addCORSHeaders(headers);
  
  return new Response(response.body, {
    status: response.status,
    statusText: response.statusText,
    headers: headers
  });
}

/**
 * Add CORS headers to response
 */
function addCORSHeaders(headers) {
  headers.set('Access-Control-Allow-Origin', CONFIG.allowedOrigins);
  headers.set('Access-Control-Allow-Methods', 'GET, OPTIONS');
  headers.set('Access-Control-Allow-Headers', 'Content-Type');
  headers.set('Access-Control-Max-Age', '86400'); // 24 hours
  return headers;
}

/**
 * Handle CORS preflight requests
 */
function handleCORSPreflight() {
  return new Response(null, {
    status: 204,
    headers: addCORSHeaders(new Headers())
  });
}

/**
 * Create error response
 */
function createErrorResponse(message, status = 500) {
  return new Response(JSON.stringify({
    error: true,
    message: message,
    status: status
  }), {
    status: status,
    headers: addCORSHeaders(new Headers({
      'Content-Type': 'application/json'
    }))
  });
}

/**
 * Check if domain is allowed
 */
function isAllowedDomain(url) {
  try {
    const urlObj = new URL(url);
    const allowedDomains = [
      'jeniusplay.com',
      'tv12.idlixku.com',
      'idlix.com',
      'idlixku.com'
    ];
    
    return allowedDomains.some(domain => 
      urlObj.hostname === domain || urlObj.hostname.endsWith('.' + domain)
    );
  } catch {
    return false;
  }
}
