import { Package } from "lucide-react";
import { AppCard } from "./AppCard";
import type { GalleryItem } from "../types";

interface AppGridProps {
  items: GalleryItem[];
  emptyMessage: string;
  onNavigate: (id: string) => void;
}

export function AppGrid({ items, emptyMessage, onNavigate }: AppGridProps) {
  if (items.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-muted-foreground">
        <Package className="h-10 w-10 mb-3 opacity-40" />
        <p className="text-sm">{emptyMessage}</p>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-2 sm:grid sm:grid-cols-2 sm:gap-3 lg:grid-cols-3">
      {items.map((item) => (
        <AppCard
          key={item.id}
          item={item}
          onNavigate={onNavigate}
        />
      ))}
    </div>
  );
}
