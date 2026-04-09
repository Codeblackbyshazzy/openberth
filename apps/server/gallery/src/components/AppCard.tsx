import { ExternalLink } from "lucide-react";
import { Card, CardContent } from "./ui/card";
import { AppIcon } from "./AppIcon";
import { StatusDot } from "./StatusDot";
import type { GalleryItem } from "../types";

interface AppCardProps {
  item: GalleryItem;
  onNavigate: (id: string) => void;
}

export function AppCard({ item, onNavigate }: AppCardProps) {
  const notRunning = item.status !== "running";

  return (
    <Card
      className={`cursor-pointer transition-colors hover:bg-muted/50${notRunning ? " opacity-60" : ""}`}
      onClick={() => onNavigate(item.id)}
    >
      <CardContent className="flex items-start gap-2.5 py-2.5 px-3">
        <AppIcon title={item.title || item.name} framework={item.framework} size="sm" />
        <div className="min-w-0 flex-1">
          <h3 className="font-semibold text-sm truncate">{item.title || item.name}</h3>
          <div className="mt-0.5">
            <StatusDot status={item.status} />
          </div>
          {item.description && (
            <p className="text-xs text-muted-foreground line-clamp-1 mt-0.5">{item.description}</p>
          )}
        </div>
        <a
          href={item.url}
          target="_blank"
          rel="noopener noreferrer"
          className="inline-flex items-center gap-1 text-xs text-primary hover:underline shrink-0 mt-0.5"
          onClick={(e) => e.stopPropagation()}
        >
          Visit
          <ExternalLink className="h-3 w-3" />
        </a>
      </CardContent>
    </Card>
  );
}
