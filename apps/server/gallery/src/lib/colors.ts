interface FrameworkColor {
  bg: string;
  text: string;
}

const frameworkColors: Record<string, FrameworkColor> = {
  // Node.js / JavaScript frameworks
  nextjs:   { bg: "bg-emerald-100 dark:bg-emerald-900/30", text: "text-emerald-700 dark:text-emerald-300" },
  vite:     { bg: "bg-violet-100 dark:bg-violet-900/30",   text: "text-violet-700 dark:text-violet-300" },
  react:    { bg: "bg-sky-100 dark:bg-sky-900/30",         text: "text-sky-700 dark:text-sky-300" },
  vue:      { bg: "bg-green-100 dark:bg-green-900/30",     text: "text-green-700 dark:text-green-300" },
  svelte:   { bg: "bg-orange-100 dark:bg-orange-900/30",   text: "text-orange-700 dark:text-orange-300" },
  node:     { bg: "bg-lime-100 dark:bg-lime-900/30",       text: "text-lime-700 dark:text-lime-300" },
  express:  { bg: "bg-lime-100 dark:bg-lime-900/30",       text: "text-lime-700 dark:text-lime-300" },

  // Python
  python:   { bg: "bg-blue-100 dark:bg-blue-900/30",       text: "text-blue-700 dark:text-blue-300" },
  flask:    { bg: "bg-blue-100 dark:bg-blue-900/30",       text: "text-blue-700 dark:text-blue-300" },
  django:   { bg: "bg-teal-100 dark:bg-teal-900/30",       text: "text-teal-700 dark:text-teal-300" },
  fastapi:  { bg: "bg-cyan-100 dark:bg-cyan-900/30",       text: "text-cyan-700 dark:text-cyan-300" },
  streamlit:{ bg: "bg-red-100 dark:bg-red-900/30",         text: "text-red-700 dark:text-red-300" },

  // Go
  go:       { bg: "bg-cyan-100 dark:bg-cyan-900/30",       text: "text-cyan-700 dark:text-cyan-300" },

  // Static
  static:   { bg: "bg-gray-100 dark:bg-gray-800/50",       text: "text-gray-600 dark:text-gray-400" },
  html:     { bg: "bg-amber-100 dark:bg-amber-900/30",     text: "text-amber-700 dark:text-amber-300" },
  markdown: { bg: "bg-stone-100 dark:bg-stone-800/50",     text: "text-stone-600 dark:text-stone-400" },
  notebook: { bg: "bg-orange-100 dark:bg-orange-900/30",   text: "text-orange-700 dark:text-orange-300" },
};

const fallbackPalettes: FrameworkColor[] = [
  { bg: "bg-rose-100 dark:bg-rose-900/30",     text: "text-rose-700 dark:text-rose-300" },
  { bg: "bg-indigo-100 dark:bg-indigo-900/30", text: "text-indigo-700 dark:text-indigo-300" },
  { bg: "bg-fuchsia-100 dark:bg-fuchsia-900/30", text: "text-fuchsia-700 dark:text-fuchsia-300" },
  { bg: "bg-pink-100 dark:bg-pink-900/30",     text: "text-pink-700 dark:text-pink-300" },
  { bg: "bg-yellow-100 dark:bg-yellow-900/30", text: "text-yellow-700 dark:text-yellow-300" },
];

function simpleHash(str: string): number {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    hash = ((hash << 5) - hash + str.charCodeAt(i)) | 0;
  }
  return Math.abs(hash);
}

export function getFrameworkColor(framework: string, appName: string): FrameworkColor {
  const key = framework.toLowerCase().replace(/[^a-z]/g, "");
  if (frameworkColors[key]) return frameworkColors[key];
  return fallbackPalettes[simpleHash(appName) % fallbackPalettes.length];
}

export function getAppInitials(title: string): string {
  const words = title.trim().split(/\s+/);
  if (words.length >= 2) {
    return (words[0][0] + words[1][0]).toUpperCase();
  }
  return title.slice(0, 2).toUpperCase();
}
