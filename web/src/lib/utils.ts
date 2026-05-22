import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatBytes(n: number): string {
  if (!n || n < 0) return '0';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let i = 0;
  let v = n;
  while (v >= 1024 && i < units.length - 1) {
    v /= 1024;
    i++;
  }
  return `${v.toFixed(2)} ${units[i]}`;
}

export function errorMessage(e: unknown, fallback = '操作失败'): string {
  const r = (e as { response?: { data?: { error?: string } } })?.response?.data?.error;
  return r ?? fallback;
}
