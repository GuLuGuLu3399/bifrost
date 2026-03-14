import { useRuntimeConfig } from "#app";

/**
 * 图片工具函数
 */

/**
 * 获取 CDN 图片 URL
 * @param key 对象存储的 key
 * @returns 完整的图片 URL
 */
export const getImageUrl = (key?: string): string => {
  const config = useRuntimeConfig();
  if (!key) return "/placeholder.png";
  return `${config.public.cdnUrl}/${key}`;
};

/**
 * 日期格式化函数集合
 */

/**
 * 格式化日期为简单格式（yyyy-MM-dd）
 */
export const formatDate = (dateStr?: string): string => {
  if (!dateStr) return "";
  try {
    return new Date(dateStr).toLocaleDateString("zh-CN");
  } catch {
    return "";
  }
};

/**
 * 格式化日期为详细格式（yyyy-MM-dd HH:mm）
 */
export const formatDateDetailed = (dateStr?: string): string => {
  if (!dateStr) return "";
  try {
    return new Date(dateStr).toLocaleString("zh-CN", {
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
    });
  } catch {
    return "";
  }
};

/**
 * 相对时间格式（如"2小时前"）
 */
export const formatRelativeTime = (dateStr?: string): string => {
  if (!dateStr) return "";
  try {
    const now = new Date();
    const past = new Date(dateStr);
    const diff = now.getTime() - past.getTime();

    const seconds = Math.floor(diff / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);
    const days = Math.floor(hours / 24);

    if (seconds < 60) return "刚刚";
    if (minutes < 60) return `${minutes}分钟前`;
    if (hours < 24) return `${hours}小时前`;
    if (days < 7) return `${days}天前`;

    return formatDate(dateStr);
  } catch {
    return "";
  }
};

/**
 * 字符串工具函数
 */

/**
 * 截断字符串，超长部分用省略号表示
 */
export const truncate = (text: string, length: number = 100): string => {
  if (!text) return "";
  return text.length > length ? text.substring(0, length) + "..." : text;
};

/**
 * 移除 HTML 标签
 */
export const stripHtml = (html?: string): string => {
  if (!html) return "";
  return html.replace(/<[^>]*>/g, "");
};

/**
 * 计算阅读时间（基于字数，假设每分钟300字）
 */
export const calculateReadingTime = (content?: string): number => {
  if (!content) return 0;
  const text = stripHtml(content);
  const words = text.length;
  return Math.max(1, Math.ceil(words / 300));
};
