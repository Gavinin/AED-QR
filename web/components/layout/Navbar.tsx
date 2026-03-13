"use client";

import { 
  Navbar, 
  NavbarBrand, 
  NavbarContent, 
  Avatar, 
  Dropdown, 
  DropdownTrigger, 
  DropdownMenu, 
  DropdownItem,
} from "@heroui/react";
import { useTranslation } from "react-i18next";
import { useTheme } from "next-themes";
import { FaMoon, FaSun, FaGlobe, FaSignOutAlt } from "react-icons/fa";
import { useRouter, usePathname } from "next/navigation";
import { useEffect, useState } from "react";
import "../../lib/i18n/client";
import LanguageSelector from "../shared/LanguageSelector";

// Use relative path to leverage Next.js rewrites
const API_URL = "/api/admin";

export default function AppNavbar() {
  const { t, i18n } = useTranslation();
  const router = useRouter();
  const pathname = usePathname();
  const { theme, setTheme } = useTheme();
  const [username, setUsername] = useState("");
  const [mounted, setMounted] = useState(false);
  const [hasToken, setHasToken] = useState(false);

  useEffect(() => {
    setMounted(true);
    // Initial check from local storage
    const token = localStorage.getItem("token");
    setHasToken(!!token);
    
    if (token) {
        const storedUsername = localStorage.getItem("username");
        if (storedUsername) {
          setUsername(storedUsername);
        }
    }
  }, [pathname]);

  const handleLogout = async () => {
    try {
        const token = localStorage.getItem("token");
        if (token) {
             await fetch(`${API_URL}/logout`, {
                method: "POST",
                headers: { Authorization: `Bearer ${token}` }
            });
        }
    } catch (error) {
        console.error("Logout failed", error);
    } finally {
        localStorage.removeItem("token");
        localStorage.removeItem("username");
        setUsername("");
        router.push("/login");
    }
  };

  const toggleTheme = () => {
    setTheme(theme === "dark" ? "light" : "dark");
  };

  if (!mounted) return null;

  // Only show navbar if user is authenticated (has token)
  if (!hasToken) return null;

  // Don't show navbar on login page
  if (typeof window !== 'undefined') {
    if (window.location.pathname === '/login') {
        return null;
    }
  }

  return (
    <Navbar isBordered height="4rem" maxWidth="full" className="h-navbar">
      <NavbarBrand>
        <p className="font-bold text-inherit">AED QR</p>
      </NavbarBrand>
      
      <NavbarContent as="div" justify="end">
        <Dropdown placement="bottom-end">
          <DropdownTrigger>
            <Avatar
              as="button"
              className="transition-transform"
              color="primary"
              name={username ? username.substring(0, 2).toUpperCase() : "?"}
              size="sm"
              isBordered
            />
          </DropdownTrigger>
          <DropdownMenu aria-label="Profile Actions" variant="flat">
            <DropdownItem key="profile" className="h-14 gap-2" textValue="Signed in as">
              <p className="font-semibold">{t("dashboard.welcome")}</p>
              <p className="font-semibold">{username || "Guest"}</p>
            </DropdownItem>
            
            <DropdownItem key="language" textValue="Language" isReadOnly className="p-0">
              <LanguageSelector 
                customTrigger={(onOpen) => (
                  <div 
                    className="flex w-full items-center gap-2 px-2 py-1.5 cursor-pointer hover:bg-default-100 rounded-small transition-colors"
                    onClick={onOpen}
                  >
                    <FaGlobe />
                    <span>{t("dashboard.language")}</span>
                  </div>
                )}
              />
            </DropdownItem>

            <DropdownItem key="theme" startContent={theme === "dark" ? <FaSun /> : <FaMoon />} onPress={toggleTheme}>
              {t("dashboard.dark_mode")}
            </DropdownItem>
            <DropdownItem key="logout" color="danger" startContent={<FaSignOutAlt />} onPress={handleLogout}>
              {t("dashboard.logout")}
            </DropdownItem>
          </DropdownMenu>
        </Dropdown>
      </NavbarContent>
    </Navbar>
  );
}
