import { useCallback, useEffect, useState } from "react";
import type { GalleryItem, GalleryResponse } from "../types";

interface UseGalleryOptions {
  apiKey: string;
  onAuthData: (data: GalleryResponse) => void;
}

export function useGallery({ apiKey, onAuthData }: UseGalleryOptions) {
  const [items, setItems] = useState<GalleryItem[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchGallery = useCallback(async () => {
    const headers: Record<string, string> = {};
    if (apiKey) headers["X-API-Key"] = apiKey;
    try {
      const r = await fetch("/api/gallery", { headers, credentials: "same-origin" });
      if (r.status === 401) {
        if (!import.meta.env.DEV) {
          window.location.href = "/login?redirect=/gallery/";
        }
        throw new Error("unauthorized");
      }
      const data: GalleryResponse = await r.json();
      setItems(data.deployments || []);
      onAuthData(data);
    } catch {
      if (import.meta.env.DEV) {
        const { mockGalleryResponse } = await import("../lib/mockData");
        setItems(mockGalleryResponse.deployments);
        onAuthData(mockGalleryResponse);
      }
    } finally {
      setLoading(false);
    }
  }, [apiKey, onAuthData]);

  useEffect(() => {
    fetchGallery();
  }, [fetchGallery]);

  return { items, loading, fetchGallery };
}
