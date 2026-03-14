import type { PostsListResponse, PostDetailResponse } from "@bifrost/shared";
import { useApi } from "./useApi";

/**
 * 文章数据获取 Composable
 * 处理文章列表和详情的数据逻辑
 */

/**
 * 获取文章列表
 */
export const usePostList = (
  pageSize: number = 20,
  categoryId?: string,
  tagId?: string
) => {
  const query: Record<string, any> = {
    "page.page_size": pageSize,
  };

  if (categoryId) {
    query.category_id = categoryId;
  }
  if (tagId) {
    query.tag_id = tagId;
  }

  return useApi<PostsListResponse>("/v1/posts", {
    query,
  });
};

/**
 * 获取文章详情（通过 slug）
 */
export const usePostDetail = (slug: string) => {
  return useApi<PostDetailResponse>(`/v1/posts/${slug}`);
};

/**
 * 批量获取文章摘要（用于搜索结果聚合）
 */
export const usePostsBatch = (postIds: string[]) => {
  return useApi("/v1/posts:batch", {
    method: "POST",
    body: {
      post_ids: postIds,
    },
  });
};
