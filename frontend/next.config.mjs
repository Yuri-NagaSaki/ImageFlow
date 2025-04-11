/** @type {import('next').NextConfig} */
import fs from 'node:fs';
import path from 'node:path';
import dotenv from 'dotenv';

// Load environment variables from parent directory's .env file
const parentEnvPath = path.resolve(process.cwd(), '../.env');
if (fs.existsSync(parentEnvPath)) {
  const parentEnv = dotenv.parse(fs.readFileSync(parentEnvPath));
  // Merge parent .env variables into process.env
  for (const [key, value] of Object.entries(parentEnv)) {
    process.env[key] = value;
  }
}

/** @type {boolean} */
const isStaticExport = !process.env.NEXT_PUBLIC_API_URL;

// Parse remote patterns to extract protocol if present
const parseRemotePatterns = (patterns) => {
  if (!patterns) return [{ protocol: 'http', hostname: 'default-domain.com' }];

  const patternList = patterns.split(',');
  return patternList.map(pattern => {
    pattern = pattern.trim();
    if (!pattern) return null; // 忽略空字符串

    try {
      // 尝试解析为 URL
      const url = new URL(pattern);
      return {
        protocol: url.protocol.replace(':', ''),
        hostname: url.hostname
      };
    } catch {
      // 如果解析失败，返回默认值
      return {
        protocol: 'http',
        hostname: pattern
      };
    }
  }).filter(Boolean); // 过滤掉无效的条目
};

const remotePatterns = parseRemotePatterns(process.env.NEXT_PUBLIC_REMOTE_PATTERNS);

const nextConfig = {
  reactStrictMode: true,
  swcMinify: true,
  output: isStaticExport ? 'export' : 'standalone',
  images: {
    unoptimized: isStaticExport,
    remotePatterns: remotePatterns
  },
  optimizeFonts: false,
  // We'll get the config from the API instead of environment variables
  env: {}
};

export default nextConfig;
