import { Sun, Moon, Monitor } from "lucide-react";
import { Button } from "./ui/button";

interface ThemeToggleProps {
  theme: "light" | "dark" | "system";
  onCycle: () => void;
}

export function ThemeToggle({ theme, onCycle }: ThemeToggleProps) {
  const Icon = theme === "light" ? Sun : theme === "dark" ? Moon : Monitor;
  const label = theme === "light" ? "Light mode" : theme === "dark" ? "Dark mode" : "System theme";

  return (
    <Button
      variant="outline"
      size="sm"
      className="h-8 w-8 p-0"
      title={label}
      onClick={onCycle}
    >
      <Icon className="h-3.5 w-3.5" />
    </Button>
  );
}
