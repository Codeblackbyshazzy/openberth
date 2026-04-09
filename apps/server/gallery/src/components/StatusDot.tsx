interface StatusDotProps {
  status: string;
}

const statusConfig: Record<string, { color: string; pulse?: boolean; label: string }> = {
  running:  { color: "bg-green-500", label: "Running" },
  building: { color: "bg-blue-500", pulse: true, label: "Building" },
  updating: { color: "bg-blue-500", pulse: true, label: "Updating" },
  failed:   { color: "bg-red-500", label: "Failed" },
  stopped:  { color: "bg-gray-400 dark:bg-gray-600", label: "Stopped" },
};

export function StatusDot({ status }: StatusDotProps) {
  const config = statusConfig[status] || { color: "bg-gray-400", label: status };

  return (
    <span className="flex items-center gap-1.5" title={config.label}>
      <span className={`inline-block h-2 w-2 rounded-full ${config.color} ${config.pulse ? "animate-pulse" : ""}`} />
      <span className="text-[10px] text-muted-foreground">{config.label}</span>
    </span>
  );
}
