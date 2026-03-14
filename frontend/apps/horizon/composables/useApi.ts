import type { UseFetchOptions } from "nuxt/app";
import { useRuntimeConfig, useCookie, useFetch, navigateTo } from "#app";

/**
 * 统一的 API 请求 Composable
 * 自动注入 baseURL、认证 Token 和错误处理
 */
export const useApi = <T = any>(
  url: string,
  options: UseFetchOptions<T> = {}
) => {
  const config = useRuntimeConfig();
  const authToken = useCookie<string>("auth_token");

  // 构建请求头
  const headers: Record<string, string> = {};

  // 合并用户传入的请求头
  if (options.headers) {
    const userHeaders = options.headers;
    if (typeof userHeaders === "object") {
      for (const [key, value] of Object.entries(userHeaders)) {
        if (typeof value === "string") {
          headers[key] = value;
        }
      }
    }
  }

  if (authToken.value) {
    headers["Authorization"] = `Bearer ${authToken.value}`;
  }

  return useFetch<T>(url, {
    baseURL: config.public.apiBase,
    ...options,
    headers,
    onResponseError({ response }: { response: any }) {
      // 401 未授权，跳转登录
      if (response.status === 401) {
        authToken.value = "";
        navigateTo("/auth/login");
      }
      // 其他错误可在此处统一处理，如 toast 提示
    },
  });
};
