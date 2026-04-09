import { useCallback, useState } from "react";

const API_KEY_STORAGE = "openberth_api_key";

export function authHeaders(apiKey: string): Record<string, string> {
  if (!apiKey) return {};
  return { "X-API-Key": apiKey };
}

export function useAuth() {
  const [apiKey] = useState(() => localStorage.getItem(API_KEY_STORAGE) || "");
  const [currentUserId, setCurrentUserId] = useState<string | null>(null);
  const [userRole, setUserRole] = useState<string | null>(null);
  const [userName, setUserName] = useState<string | null>(null);
  const [hasPassword, setHasPassword] = useState(false);
  const [serverVersion, setServerVersion] = useState("");

  const isAdmin = userRole === "admin";

  const handleLogin = useCallback(() => {
    window.location.href = "/login?redirect=/gallery/";
  }, []);

  const handleLogout = useCallback(async () => {
    localStorage.removeItem(API_KEY_STORAGE);
    await fetch("/logout", { method: "POST", credentials: "same-origin" });
    window.location.reload();
  }, []);

  return {
    apiKey,
    currentUserId,
    setCurrentUserId,
    userRole,
    setUserRole,
    userName,
    setUserName,
    hasPassword,
    setHasPassword,
    serverVersion,
    setServerVersion,
    isAdmin,
    handleLogin,
    handleLogout,
  };
}
