#!/usr/bin/env node

const baseURL = (process.env.WECHECKIN_PERF_BASE_URL || 'http://127.0.0.1:8083').replace(/\/+$/, '');
const strict = isEnabled(process.env.WECHECKIN_PERF_STRICT);
const repeat = positiveInt(process.env.WECHECKIN_PERF_REPEAT, 5);
const timeoutMs = positiveInt(process.env.WECHECKIN_PERF_TIMEOUT_MS, 10000);

const tokens = {
  admin: process.env.WECHECKIN_ADMIN_TOKEN || '',
  user: process.env.WECHECKIN_USER_TOKEN || '',
  dingtalk: process.env.WECHECKIN_DINGTALK_TOKEN || '',
};

const endpoints = [
  { name: '后台用户列表', method: 'GET', path: '/api/v2/admin/users?page=1&pageSize=20', token: 'admin', thresholdMs: 250 },
  { name: '后台管理员列表', method: 'GET', path: '/api/v2/admin/managers?page=1&pageSize=20', token: 'admin', thresholdMs: 250 },
  { name: '后台角色列表', method: 'GET', path: '/api/v2/admin/roles?page=1&pageSize=20', token: 'admin', thresholdMs: 250 },
  { name: '钉钉 H5 启动信息', method: 'GET', path: '/api/v2/dingtalk/h5/bootstrap', token: 'dingtalk', thresholdMs: 150 },
  { name: '钉钉 H5 工作台', method: 'GET', path: '/api/v2/dingtalk/h5/workbench', token: 'dingtalk', thresholdMs: 250 },
  { name: '钉钉 H5 绩效列表', method: 'GET', path: '/api/v2/dingtalk/h5/reviews?page=1&pageSize=20', token: 'dingtalk', thresholdMs: 300 },
  { name: '通知列表', method: 'GET', path: '/api/v2/news?page=1&pageSize=20', token: 'user', optionalToken: true, thresholdMs: 300 },
  { name: '打卡任务列表', method: 'GET', path: '/api/v2/enrollments?page=1&pageSize=20', token: 'user', optionalToken: true, thresholdMs: 300 },
  { name: '赛事活动列表', method: 'GET', path: '/api/v2/events?page=1&pageSize=20', token: 'user', optionalToken: true, thresholdMs: 300 },
  { name: '问卷列表', method: 'GET', path: '/api/v2/surveys?page=1&pageSize=20', token: 'user', optionalToken: true, thresholdMs: 300 },
  { name: '考试列表', method: 'GET', path: '/api/v2/exams?page=1&pageSize=20', token: 'user', optionalToken: true, thresholdMs: 300 },
];

const results = [];

for (const endpoint of endpoints) {
  const token = endpoint.token ? tokens[endpoint.token] : '';
  if (endpoint.token && !endpoint.optionalToken && !token) {
    results.push({ ...endpoint, status: 'SKIP', reason: `missing WECHECKIN_${endpoint.token.toUpperCase()}_TOKEN` });
    continue;
  }

  const samples = [];
  const statuses = [];
  let errorMessage = '';

  for (let i = 0; i < repeat; i += 1) {
    const result = await measure(endpoint, token);
    if (result.error) {
      errorMessage = result.error;
      break;
    }
    samples.push(result.durationMs);
    statuses.push(result.statusCode);
  }

  if (errorMessage) {
    results.push({ ...endpoint, status: 'ERROR', reason: errorMessage });
    continue;
  }

  results.push({
    ...endpoint,
    status: statuses.every((statusCode) => statusCode >= 200 && statusCode < 400) ? 'OK' : `HTTP ${[...new Set(statuses)].join('/')}`,
    samples,
    stats: summarize(samples),
  });
}

printReport(results);

const failed = results.filter((item) => item.status !== 'SKIP' && (item.status === 'ERROR' || item.status.startsWith('HTTP') || (item.stats?.p95 ?? 0) > item.thresholdMs));
if (failed.length > 0) {
  const prefix = strict ? '性能检查失败' : '性能检查告警';
  console.log(`\n${prefix}: ${failed.map((item) => item.name).join('、')}`);
  if (strict) {
    process.exitCode = 1;
  }
}

async function measure(endpoint, token) {
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), timeoutMs);
  const headers = {
    Accept: 'application/json',
    'X-Client-Platform': 'performance-check',
  };
  if (token) {
    headers.Authorization = normalizeAuthToken(token);
  }

  const startedAt = performance.now();
  try {
    const response = await fetch(`${baseURL}${endpoint.path}`, {
      method: endpoint.method,
      headers,
      signal: controller.signal,
    });
    const bodyText = await response.text();
    const businessError = validateBusinessResponse(bodyText);
    if (businessError) {
      return {
        error: businessError,
      };
    }
    return {
      statusCode: response.status,
      durationMs: performance.now() - startedAt,
    };
  } catch (error) {
    return {
      error: error?.name === 'AbortError' ? `timeout ${timeoutMs}ms` : String(error?.message || error),
    };
  } finally {
    clearTimeout(timer);
  }
}

function normalizeAuthToken(token) {
  return String(token || '').trim().replace(/^Bearer\s+/i, '');
}

function validateBusinessResponse(bodyText) {
  if (!bodyText) {
    return '';
  }
  let payload;
  try {
    payload = JSON.parse(bodyText);
  } catch {
    return 'invalid JSON response';
  }
  if (!payload || typeof payload !== 'object' || !Object.prototype.hasOwnProperty.call(payload, 'code')) {
    return '';
  }
  const code = Number(payload.code);
  if (Number.isFinite(code) && code !== 0) {
    const message = payload.msg ? `: ${payload.msg}` : '';
    return `business code ${payload.code}${message}`;
  }
  return '';
}

function summarize(samples) {
  const sorted = [...samples].sort((a, b) => a - b);
  const sum = sorted.reduce((total, value) => total + value, 0);
  const p95Index = Math.max(0, Math.ceil(sorted.length * 0.95) - 1);
  return {
    min: sorted[0],
    avg: sum / sorted.length,
    p95: sorted[p95Index],
    max: sorted[sorted.length - 1],
  };
}

function printReport(items) {
  console.log(`WeCheckin performance baseline`);
  console.log(`baseURL=${baseURL} repeat=${repeat} timeout=${timeoutMs}ms strict=${strict ? 'on' : 'off'}\n`);
  console.log([
    pad('接口', 18),
    pad('状态', 10),
    pad('阈值', 8),
    pad('min', 8),
    pad('avg', 8),
    pad('p95', 8),
    pad('max', 8),
    '路径',
  ].join('  '));

  for (const item of items) {
    if (item.status === 'SKIP' || item.status === 'ERROR') {
      console.log([
        pad(item.name, 18),
        pad(item.status, 10),
        pad(`${item.thresholdMs}ms`, 8),
        pad('-', 8),
        pad('-', 8),
        pad('-', 8),
        pad('-', 8),
        `${item.path} (${item.reason})`,
      ].join('  '));
      continue;
    }

    const marker = item.stats.p95 > item.thresholdMs ? 'WARN' : item.status;
    console.log([
      pad(item.name, 18),
      pad(marker, 10),
      pad(`${item.thresholdMs}ms`, 8),
      pad(formatMs(item.stats.min), 8),
      pad(formatMs(item.stats.avg), 8),
      pad(formatMs(item.stats.p95), 8),
      pad(formatMs(item.stats.max), 8),
      item.path,
    ].join('  '));
  }
}

function formatMs(value) {
  return `${value.toFixed(1)}ms`;
}

function pad(value, width) {
  const text = String(value);
  const visualLength = [...text].reduce((total, char) => total + (char.charCodeAt(0) > 255 ? 2 : 1), 0);
  return text + ' '.repeat(Math.max(0, width - visualLength));
}

function positiveInt(value, fallback) {
  const parsed = Number.parseInt(value || '', 10);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : fallback;
}

function isEnabled(value) {
  return ['1', 'true', 'yes', 'on'].includes(String(value || '').toLowerCase());
}
