"use client";

import { useEffect, useState } from "react";
import { useRouter, usePathname } from "next/navigation";

// Use environment variable for API URL
const BASE_API_URL = process.env.NEXT_PUBLIC_API_URL || "/api";
const API_URL = `${BASE_API_URL}/admin`;

export default function AuthGuard({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const pathname = usePathname();
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Skip auth check for login page and public AED pages
    if (pathname === "/login" || pathname.startsWith("/aed")) {
      setLoading(false);
      return;
    }

    const checkAuth = async () => {
      const token = localStorage.getItem("token");
      if (!token) {
        router.push("/login");
        return;
      }

      try {
        const response = await fetch(`${API_URL}/check_token`, {
          headers: { Authorization: `Bearer ${token}` },
        });
        
        if (!response.ok) {
            throw new Error("Token invalid");
        }

        setLoading(false);
      } catch (error) {
        console.error("Token invalid", error);
        localStorage.removeItem("token");
        router.push("/login");
      }
    };

    checkAuth();

    // Check token every minute
    const interval = setInterval(checkAuth, 60000);

    return () => clearInterval(interval);
  }, [router, pathname]);

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-primary"></div>
      </div>
    );
  }

  return <>{children}</>;
}
