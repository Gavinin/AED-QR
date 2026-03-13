import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: 'export',
  // Static exports do not support rewrites. 
  // API calls must be handled by Nginx proxy or absolute URLs in client code.
};

export default nextConfig;
