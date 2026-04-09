export function formatAge(createdAt: string): string {
  if (!createdAt) return "?";
  let t: Date;
  if (createdAt.includes("T")) {
    t = new Date(createdAt);
  } else {
    t = new Date(createdAt.replace(" ", "T") + "Z");
  }
  if (isNaN(t.getTime())) return "?";
  const seconds = Math.floor((Date.now() - t.getTime()) / 1000);
  if (seconds < 60) return `${seconds}s ago`;
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  return `${Math.floor(hours / 24)}d ago`;
}

export function formatExpiry(expiresAt: string): string {
  if (!expiresAt) return "";
  let t: Date;
  if (expiresAt.includes("T")) {
    t = new Date(expiresAt);
  } else {
    t = new Date(expiresAt.replace(" ", "T") + "Z");
  }
  if (isNaN(t.getTime())) return "";
  const diff = t.getTime() - Date.now();
  if (diff <= 0) return "expired";
  const hours = Math.floor(diff / 3600000);
  if (hours < 1) return `in ${Math.floor(diff / 60000)}m`;
  if (hours < 24) return `in ${hours}h`;
  return `in ${Math.floor(hours / 24)}d`;
}
