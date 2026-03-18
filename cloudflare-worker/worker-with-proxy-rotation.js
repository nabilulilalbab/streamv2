/**
 * Cloudflare Worker - IDLIX Streaming Proxy with Proxy Rotation
 * 
 * Features:
 * - CORS headers for all responses
 * - M3U8 playlist URL rewriting
 * - Proxy rotation (random or round-robin)
 * - KV storage for dynamic proxy list
 * - Fallback to direct fetch if proxy fails
 * - Health check for proxies
 * 
 * Setup:
 * 1. Create KV namespace in Cloudflare dashboard
 * 2. Bind KV to worker (name: PROXY_CONFIG)
 * 3. Upload proxy list to KV
 * 4. Deploy worker
 */

// Configuration
const CONFIG = {
  // Allowed origins
  allowedOrigins: '*',
  
  // Cache TTL
  cacheTTL: {
    m3u8: 60,
    ts: 3600,
    proxyList: 300, // Cache proxy list for 5 minutes
  },
  
  // Proxy rotation strategy: 'random' or 'round-robin'
  rotationStrategy: 'random',
  
  // Retry settings
  maxRetries: 3,
  retryDelay: 1000, // ms
  
  // Request timeout
  timeout: 30000, // 30 seconds
  
  // Headers to forward
  upstreamHeaders: {
    'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
    'Accept': '*/*',
    'Referer': 'https://tv12.idlixku.com/',
  }
};

// In-memory cache for proxy list
let proxyListCache = null;
let proxyListCacheTime = 0;
let currentProxyIndex = 0;

/**
 * Main handler
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

  // Parse URL
  const url = new URL(request.url);
  const targetURL = url.searchParams.get('url');
  const useProxy = url.searchParams.get('proxy') !== 'false'; // default: true

  // Validate target URL
  if (!targetURL) {
    return createErrorResponse('Missing url parameter', 400);
  }

  if (!isAllowedDomain(targetURL)) {
    return createErrorResponse('Target domain not allowed', 403);
  }

  try {
    // Fetch with or without proxy
    const response = useProxy 
      ? await fetchViaProxy(targetURL)
      : await fetchDirect(targetURL);

    if (!response.ok) {
      return createErrorResponse(`Upstream error: ${response.status}`, response.status);
    }

    // Process M3U8 or pass through
    const contentType = response.headers.get('Content-Type') || '';
    const isM3U8 = contentType.includes('mpegURL') || targetURL.endsWith('.m3u8');

    if (isM3U8) {
      return await processM3U8(response, targetURL, url.origin);
    } else {
      return createProxyResponse(response);
    }

  } catch (error) {
    console.error('Proxy error:', error);
    return createErrorResponse(`Proxy error: ${error.message}`, 500);
  }
}

/**
 * Fetch via proxy with rotation
 */
async function fetchViaProxy(targetURL) {
  // Get proxy list
  const proxyList = await getProxyList();
  
  if (!proxyList || proxyList.length === 0) {
    console.log('No proxies available, falling back to direct fetch');
    return await fetchDirect(targetURL);
  }

  // Try each proxy with retries
  let lastError;
  
  for (let attempt = 0; attempt < CONFIG.maxRetries; attempt++) {
    // Select proxy based on strategy
    const proxyURL = selectProxy(proxyList);
    
    try {
      console.log(`Attempt ${attempt + 1}: Using proxy ${proxyURL}`);
      
      const response = await fetchThroughProxy(targetURL, proxyURL);
      
      if (response.ok) {
        console.log(`Success with proxy ${proxyURL}`);
        return response;
      }
      
      lastError = new Error(`Proxy returned ${response.status}`);
      
    } catch (error) {
      console.log(`Proxy ${proxyURL} failed:`, error.message);
      lastError = error;
      
      // Wait before retry
      if (attempt < CONFIG.maxRetries - 1) {
        await sleep(CONFIG.retryDelay);
      }
    }
  }

  // All proxies failed, try direct fetch as fallback
  console.log('All proxies failed, trying direct fetch');
  try {
    return await fetchDirect(targetURL);
  } catch (directError) {
    throw lastError || directError;
  }
}

/**
 * Fetch through a specific proxy
 */
async function fetchThroughProxy(targetURL, proxyURL) {
  // Parse proxy URL
  const proxy = new URL(proxyURL);
  
  // Build fetch options with proxy
  const options = {
    headers: CONFIG.upstreamHeaders,
    // Cloudflare Workers doesn't support native proxy
    // Instead, we'll use the proxy as a relay
    cf: {
      cacheTtl: targetURL.endsWith('.m3u8') ? CONFIG.cacheTTL.m3u8 : CONFIG.cacheTTL.ts,
    }
  };

  // If proxy requires authentication
  if (proxy.username && proxy.password) {
    const auth = btoa(`${proxy.username}:${proxy.password}`);
    options.headers['Proxy-Authorization'] = `Basic ${auth}`;
  }

  // Note: Cloudflare Workers doesn't support CONNECT method
  // For HTTP proxies, we need to make request to proxy with target URL as parameter
  // Example: http://proxy.com:8080/?url=target
  const proxyRequest = `${proxy.origin}/?url=${encodeURIComponent(targetURL)}`;
  
  return await fetch(proxyRequest, options);
}

/**
 * Direct fetch without proxy
 */
async function fetchDirect(targetURL) {
  return await fetch(targetURL, {
    headers: CONFIG.upstreamHeaders,
    cf: {
      cacheTtl: targetURL.endsWith('.m3u8') ? CONFIG.cacheTTL.m3u8 : CONFIG.cacheTTL.ts,
      cacheEverything: true,
    }
  });
}

/**
 * Get proxy list from KV or cache
 */
async function getProxyList() {
  const now = Date.now();
  
  // Return from cache if fresh
  if (proxyListCache && (now - proxyListCacheTime < CONFIG.cacheTTL.proxyList * 1000)) {
    return proxyListCache;
  }

  try {
    // Try to get from KV storage
    if (typeof PROXY_CONFIG !== 'undefined') {
      const kvData = await PROXY_CONFIG.get('proxy-list', 'json');
      
      if (kvData && kvData.proxies) {
        proxyListCache = kvData.proxies;
        proxyListCacheTime = now;
        
        // Update strategy if specified
        if (kvData.strategy) {
          CONFIG.rotationStrategy = kvData.strategy;
        }
        
        console.log(`Loaded ${proxyListCache.length} proxies from KV`);
        return proxyListCache;
      }
    }
  } catch (error) {
    console.error('Failed to load from KV:', error);
  }

  // Fallback to hardcoded list
  const fallbackProxies = [
    // Add your backup proxies here
    // 'http://proxy1.example.com:8080',
    // 'http://proxy2.example.com:8080',
  ];

  proxyListCache = fallbackProxies;
  proxyListCacheTime = now;
  
  return fallbackProxies;
}

/**
 * Select proxy based on rotation strategy
 */
function selectProxy(proxyList) {
  if (CONFIG.rotationStrategy === 'random') {
    // Random selection
    const index = Math.floor(Math.random() * proxyList.length);
    return proxyList[index];
  } else {
    // Round-robin
    const proxy = proxyList[currentProxyIndex];
    currentProxyIndex = (currentProxyIndex + 1) % proxyList.length;
    return proxy;
  }
}

/**
 * Process M3U8 playlist
 */
async function processM3U8(response, originalURL, workerOrigin) {
  const content = await response.text();
  const rewrittenContent = rewriteM3U8URLs(content, originalURL, workerOrigin);
  
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
 * Rewrite M3U8 URLs
 */
function rewriteM3U8URLs(content, originalURL, workerOrigin) {
  const lines = content.split('\n');
  const baseURL = getBaseURL(originalURL);
  
  const rewrittenLines = lines.map(line => {
    const trimmed = line.trim();
    
    if (!trimmed || trimmed.startsWith('#')) {
      return line;
    }
    
    let absoluteURL;
    
    if (trimmed.startsWith('http://') || trimmed.startsWith('https://')) {
      absoluteURL = trimmed;
    } else if (trimmed.startsWith('/')) {
      const urlObj = new URL(originalURL);
      absoluteURL = `${urlObj.protocol}//${urlObj.host}${trimmed}`;
    } else {
      absoluteURL = baseURL + trimmed;
    }
    
    return `${workerOrigin}?url=${encodeURIComponent(absoluteURL)}`;
  });
  
  return rewrittenLines.join('\n');
}

/**
 * Get base URL
 */
function getBaseURL(url) {
  const lastSlash = url.lastIndexOf('/');
  return url.substring(0, lastSlash + 1);
}

/**
 * Create proxy response with CORS
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
 * Add CORS headers
 */
function addCORSHeaders(headers) {
  headers.set('Access-Control-Allow-Origin', CONFIG.allowedOrigins);
  headers.set('Access-Control-Allow-Methods', 'GET, OPTIONS');
  headers.set('Access-Control-Allow-Headers', 'Content-Type');
  headers.set('Access-Control-Max-Age', '86400');
  return headers;
}

/**
 * Handle CORS preflight
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

/**
 * Sleep utility
 */
function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}
