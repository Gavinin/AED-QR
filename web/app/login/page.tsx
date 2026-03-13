"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Button, Card, CardHeader, CardBody, Image, Form, Dropdown, DropdownTrigger, DropdownMenu, DropdownItem } from "@heroui/react";
import { Input } from "@heroui/input";
import { useTranslation } from "react-i18next";
import { useTheme } from "next-themes";
import { FaMoon, FaSun, FaGlobe } from "react-icons/fa";
import "../../lib/i18n/client";

// Use environment variable for API URL
const BASE_API_URL = process.env.NEXT_PUBLIC_API_URL || "/api";
const API_URL = `${BASE_API_URL}/admin`;

export default function LoginPage() {
  const { t, i18n } = useTranslation();
  const router = useRouter();
  const { theme, setTheme } = useTheme();
  
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [captcha, setCaptcha] = useState("");
  const [captchaId, setCaptchaId] = useState("");
  const [captchaImage, setCaptchaImage] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  const fetchCaptcha = async () => {
    try {
      const response = await fetch(`${API_URL}/captch`, { cache: 'no-store' });
      if (!response.ok) throw new Error("Failed to fetch captcha");
      const data = await response.json();
      setCaptchaId(data.captcha_id);
      setCaptchaImage(data.captcha);
    } catch (err) {
      console.error("Failed to fetch captcha", err);
    }
  };

  useEffect(() => {
    // Check if already logged in
    const token = localStorage.getItem("token");
    if (token) {
        // Validate token
        fetch(`${API_URL}/check_token`, {
            headers: { Authorization: `Bearer ${token}` }
        })
        .then(res => {
            if (res.ok) {
                router.push("/");
            } else {
                throw new Error("Token invalid");
            }
        })
        .catch(() => {
            localStorage.removeItem("token");
            fetchCaptcha();
        });
    } else {
        fetchCaptcha();
    }
  }, [router]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError("");

    try {
      const response = await fetch(`${API_URL}/login`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          username,
          password,
          captcha_id: captchaId,
          captcha,
        }),
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || "Login failed");
      }

      const { token } = data;
      localStorage.setItem("token", token);
      localStorage.setItem("username", username);
      router.push("/");
    } catch (err: any) {
      console.error("Login failed", err);
      setError(err.message || t("login.error"));
      fetchCaptcha(); // Refresh captcha on failure
      setCaptcha("");
    } finally {
      setLoading(false);
    }
  };

  const toggleTheme = () => {
    setTheme(theme === "dark" ? "light" : "dark");
  };

  const changeLanguage = () => {
    const newLang = i18n.language === "en" ? "zh" : "en";
    i18n.changeLanguage(newLang);
  };

  if (!mounted) return null;

  return (
    <div className="flex flex-col items-center justify-center min-h-screen bg-background p-4 relative">
      
      {/* Top Right Controls */}
      <div className="absolute top-4 right-4 flex gap-2">
        <Button isIconOnly variant="ghost" onPress={changeLanguage} aria-label="Switch Language">
            <FaGlobe size={20} />
            <span className="sr-only">{i18n.language === "en" ? "EN" : "ZH"}</span>
        </Button>
        <Button isIconOnly variant="ghost" onPress={toggleTheme} aria-label="Toggle Theme">
            {theme === "dark" ? <FaSun size={20} /> : <FaMoon size={20} />}
        </Button>
      </div>

      <Card className="w-full max-w-md">
        <CardHeader className="flex justify-center pb-0">
          <h1 className="text-2xl font-bold">{t("login.title")}</h1>
        </CardHeader>
        <CardBody>
          <Form onSubmit={handleSubmit} className="flex flex-col gap-4">
            <Input
              label={t("login.username")}
              placeholder={t("login.username_required")}
              value={username}
              onValueChange={setUsername}
              isRequired
              autoComplete="username"
            />
            <Input
              label={t("login.password")}
              placeholder={t("login.password_required")}
              type="password"
              value={password}
              onValueChange={setPassword}
              isRequired
              autoComplete="current-password"
            />
            
            <div className="flex gap-2 items-end">
              <Input
                label={t("login.captcha")}
                placeholder={t("login.captcha_required")}
                value={captcha}
                onValueChange={setCaptcha}
                isRequired
                className="flex-grow"
              />
              {captchaImage && (
                <Image 
                  src={captchaImage} 
                  alt="Captcha" 
                  className="cursor-pointer bg-white"
                  height={56}
                  width={128}
                  radius="md"
                  onClick={fetchCaptcha}
                />
              )}
            </div>

            {error && <p className="text-danger text-sm text-center">{error}</p>}

            <Button 
              type="submit" 
              color="primary" 
              isLoading={loading}
              fullWidth
            >
              {loading ? t("login.loading") : t("login.submit")}
            </Button>
          </Form>
        </CardBody>
      </Card>
    </div>
  );
}
