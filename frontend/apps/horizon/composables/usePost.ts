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
  tagId?: string,
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
    transform: (input: any) => normalizePostsListResponse(input),
  });
};

/**
 * 获取文章详情（通过 slug）
 */
export const usePostDetail = (slug: string) => {
  return useApi<PostDetailResponse>(`/v1/posts/${encodeURIComponent(slug)}`, {
    transform: (input: any) => normalizePostDetailResponse(input),
  });
};

/**
 * 批量获取文章摘要（用于搜索结果聚合）
 */
export const usePostsBatch = (postIds: string[]) => {
  return useApi<PostsListResponse>("/v1/posts:batch", {
    method: "POST",
    body: {
      post_ids: postIds,
    },
    transform: (input: any) => normalizePostsListResponse(input),
  });
};

function asRecord(data: unknown): Record<string, any> {
  return data && typeof data === "object" ? (data as Record<string, any>) : {};
}

function idToString(value: unknown): string {
  if (typeof value === "string") {
    return value;
  }
  if (typeof value === "number" && Number.isFinite(value)) {
    return String(value);
  }
  return "";
}

function normalizeAuthor(author: unknown): any {
  const obj = asRecord(author);
  return {
    id: idToString(obj.id),
    nickname: typeof obj.nickname === "string" ? obj.nickname : "",
    avatar_key:
      typeof obj.avatar_key === "string"
        ? obj.avatar_key
        : typeof obj.avatarKey === "string"
          ? obj.avatarKey
          : undefined,
  };
}

function normalizeCategory(category: unknown): any {
  const obj = asRecord(category);
  return {
    id: idToString(obj.id),
    name: typeof obj.name === "string" ? obj.name : "",
    slug: typeof obj.slug === "string" ? obj.slug : "",
    description:
      typeof obj.description === "string" ? obj.description : undefined,
    post_count:
      typeof obj.post_count === "number"
        ? obj.post_count
        : typeof obj.postCount === "number"
          ? obj.postCount
          : undefined,
  };
}

function normalizeTags(tags: unknown): any[] {
  if (!Array.isArray(tags)) {
    return [];
  }

  return tags.map((item) => {
    const obj = asRecord(item);
    return {
      id: idToString(obj.id),
      name: typeof obj.name === "string" ? obj.name : "",
      slug: typeof obj.slug === "string" ? obj.slug : "",
      post_count:
        typeof obj.post_count === "number"
          ? obj.post_count
          : typeof obj.postCount === "number"
            ? obj.postCount
            : undefined,
    };
  });
}

function normalizePostSummary(post: unknown): any {
  const obj = asRecord(post);
  return {
    id: idToString(obj.id),
    title: typeof obj.title === "string" ? obj.title : "",
    slug: typeof obj.slug === "string" ? obj.slug : "",
    summary: typeof obj.summary === "string" ? obj.summary : undefined,
    cover_image_key:
      typeof obj.cover_image_key === "string"
        ? obj.cover_image_key
        : typeof obj.coverImageKey === "string"
          ? obj.coverImageKey
          : undefined,
    view_count:
      typeof obj.view_count === "number"
        ? obj.view_count
        : typeof obj.viewCount === "number"
          ? obj.viewCount
          : undefined,
    like_count:
      typeof obj.like_count === "number"
        ? obj.like_count
        : typeof obj.likeCount === "number"
          ? obj.likeCount
          : undefined,
    comment_count:
      typeof obj.comment_count === "number"
        ? obj.comment_count
        : typeof obj.commentCount === "number"
          ? obj.commentCount
          : undefined,
    published_at: obj.published_at ?? obj.publishedAt,
    updated_at: obj.updated_at ?? obj.updatedAt,
    author: obj.author ? normalizeAuthor(obj.author) : undefined,
    category: obj.category ? normalizeCategory(obj.category) : undefined,
    tags: normalizeTags(obj.tags),
  };
}

function normalizePostsListResponse(data: unknown): PostsListResponse {
  const obj = asRecord(data);
  const posts = Array.isArray(obj.posts)
    ? obj.posts.map((item) => normalizePostSummary(item))
    : [];

  const pageObj = asRecord(obj.page);
  const page = {
    next_page_token:
      typeof pageObj.next_page_token === "string"
        ? pageObj.next_page_token
        : typeof pageObj.nextPageToken === "string"
          ? pageObj.nextPageToken
          : undefined,
    total_count:
      typeof pageObj.total_count === "number"
        ? pageObj.total_count
        : typeof pageObj.totalCount === "number"
          ? pageObj.totalCount
          : undefined,
  };

  return { posts, page };
}

function normalizePostDetailResponse(data: unknown): PostDetailResponse {
  const obj = asRecord(data);
  const postObj = obj.post ? asRecord(obj.post) : obj;
  const summary = normalizePostSummary(postObj);

  return {
    post: {
      ...summary,
      html_body:
        typeof postObj.html_body === "string"
          ? postObj.html_body
          : typeof postObj.htmlBody === "string"
            ? postObj.htmlBody
            : "",
      toc_json: postObj.toc_json ?? postObj.tocJson,
      prev_post_id: idToString(postObj.prev_post_id ?? postObj.prevPostId),
      prev_post_slug:
        typeof postObj.prev_post_slug === "string"
          ? postObj.prev_post_slug
          : typeof postObj.prevPostSlug === "string"
            ? postObj.prevPostSlug
            : undefined,
      prev_post_title:
        typeof postObj.prev_post_title === "string"
          ? postObj.prev_post_title
          : typeof postObj.prevPostTitle === "string"
            ? postObj.prevPostTitle
            : undefined,
      next_post_id: idToString(postObj.next_post_id ?? postObj.nextPostId),
      next_post_slug:
        typeof postObj.next_post_slug === "string"
          ? postObj.next_post_slug
          : typeof postObj.nextPostSlug === "string"
            ? postObj.nextPostSlug
            : undefined,
      next_post_title:
        typeof postObj.next_post_title === "string"
          ? postObj.next_post_title
          : typeof postObj.nextPostTitle === "string"
            ? postObj.nextPostTitle
            : undefined,
    },
  };
}
