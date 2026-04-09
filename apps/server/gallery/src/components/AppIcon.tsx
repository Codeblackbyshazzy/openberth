import { getFrameworkColor, getAppInitials } from "../lib/colors";

interface AppIconProps {
  title: string;
  framework: string;
  size?: "sm" | "md" | "lg";
}

const sizeClasses = {
  sm: { container: "h-8 w-8", text: "text-xs" },
  md: { container: "h-10 w-10", text: "text-sm" },
  lg: { container: "h-14 w-14", text: "text-lg" },
};

export function AppIcon({ title, framework, size = "md" }: AppIconProps) {
  const color = getFrameworkColor(framework, title);
  const initials = getAppInitials(title);
  const s = sizeClasses[size];

  return (
    <div className={`flex ${s.container} shrink-0 items-center justify-center rounded-lg ${color.bg}`}>
      <span className={`${s.text} font-bold ${color.text}`}>{initials}</span>
    </div>
  );
}
