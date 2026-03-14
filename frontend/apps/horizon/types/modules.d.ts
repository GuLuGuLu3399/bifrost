declare module "~/utils" {
  export function getImageUrl(key?: string): string;
  export function formatDate(dateStr?: string): string;
  export function formatDateDetailed(dateStr?: string): string;
  export function formatRelativeTime(dateStr?: string): string;
  export function truncate(text: string, length?: number): string;
  export function stripHtml(html?: string): string;
  export function calculateReadingTime(content?: string): number;
}

declare module "~/composables/useApi" {
  export function useApi<T = any>(url: string, options?: any): any;
}

declare module "~/composables/usePost" {
  export function usePostList(
    pageSize?: number,
    categoryId?: string,
    tagId?: string
  ): any;
  export function usePostDetail(slug: string): any;
  export function usePostsBatch(postIds: string[]): any;
}

declare module "~/composables/useSearch" {
  export function useSearch(): any;
}

declare module "#app" {
  import type { Ref } from "vue";

  export function useRoute(): any;
  export function useRouter(): any;
  export function useRuntimeConfig(): any;
  export function useCookie<T = string>(name: string, options?: any): Ref<T>;
  export function useFetch<T = any>(url: string, options?: any): any;
  export function navigateTo(
    to: string | any,
    options?: any
  ): Promise<void> | void;
  export function definePageMeta(meta: any): void;
  export function useHead(head: any): void;
  export function useSeoMeta(meta: any): void;
  export function createError(error: {
    statusCode: number;
    statusMessage: string;
  }): Error;
}
