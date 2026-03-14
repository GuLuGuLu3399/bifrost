import { invoke } from "@tauri-apps/api/core";

interface RawLoginResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

interface RawUploadResponse {
  object_key: string;
  upload_url?: string;
}

export interface LoginSession {
  accessToken: string;
  refreshToken: string;
  expiresInSeconds: number;
}

export interface UploadResult {
  objectKey: string;
  uploadUrl?: string;
}

export interface ListPostsQuery {
  pageSize?: number;
  pageToken?: string;
  categoryId?: number;
  tagId?: number;
  authorId?: number;
}

export interface RegisterPayload {
  username: string;
  email: string;
  password: string;
}

type HttpMethod = "GET" | "POST" | "PUT" | "DELETE";

interface RequestOptions {
  query?: Record<string, unknown>;
  body?: unknown;
  authRequired?: boolean;
}

export interface ChangePasswordPayload {
  oldPassword: string;
  newPassword: string;
}

export interface UpdateProfilePayload {
  nickname?: string;
  bio?: string;
  avatarKey?: string;
  theme?: string;
  language?: string;
}

export interface ListDraftsQuery {
  page?: number;
  pageSize?: number;
  keyword?: string;
  categoryId?: number | string;
  status?: number;
}

export interface CreatePostPayload {
  title: string;
  slug: string;
  rawMarkdown: string;
  categoryId?: number | string;
  tagNames?: string[];
  status?: number;
  coverImageKey?: string;
  resourceKey?: string;
}

export interface UpdatePostPayload {
  postId: number | string;
  title?: string;
  rawMarkdown?: string;
  status?: number;
  coverImageKey?: string;
  resourceKey?: string;
  categoryId?: number | string;
  tagNames?: string[];
}

export interface CreateCategoryPayload {
  name: string;
  slug: string;
  description?: string;
}

export interface UpdateCategoryPayload {
  categoryId: number | string;
  name?: string;
  slug?: string;
  description?: string;
}

export interface CreateCommentPayload {
  parentId?: number;
  content: string;
}

export interface GetUploadTicketPayload {
  filename: string;
  usage: "avatar" | "cover" | "post_image";
}

const isDev = import.meta.env.DEV;

function debugLog(action: string, payload: unknown): void {
  if (!isDev) {
    return;
  }

  console.log(`[helmApi] ${action}:`, payload);
}

async function gatewayRequest<T = unknown>(
  method: HttpMethod,
  path: string,
  options: RequestOptions = {},
): Promise<T> {
  const rawResponse = await invoke<T>("gateway_request_cmd", {
    method,
    path,
    query: options.query,
    body: options.body,
    authRequired: options.authRequired ?? true,
  });
  debugLog(`${method} ${path}`, rawResponse);
  return rawResponse;
}

export async function loginWithIdentifier(
  identifier: string,
  password: string,
): Promise<LoginSession> {
  const rawResponse = await invoke<RawLoginResponse>("login_cmd", {
    identifier,
    password,
  });

  debugLog("login response", rawResponse);

  return {
    accessToken: rawResponse.access_token,
    refreshToken: rawResponse.refresh_token,
    expiresInSeconds: rawResponse.expires_in,
  };
}

export async function registerUser(payload: RegisterPayload): Promise<void> {
  const rawResponse = await invoke<string>("register_cmd", { dto: payload });
  debugLog("register response", rawResponse);
}

export async function uploadImageFile(filePath: string): Promise<UploadResult> {
  const rawResponse = await invoke<RawUploadResponse>("upload_image_cmd", {
    filePath,
  });

  debugLog("upload response", rawResponse);

  return {
    objectKey: rawResponse.object_key,
    uploadUrl: rawResponse.upload_url,
  };
}

export async function fetchAuthStatus(): Promise<boolean> {
  const rawResponse = await invoke<boolean>("is_authenticated");
  debugLog("auth status", rawResponse);
  return rawResponse;
}

export async function logoutUser(): Promise<void> {
  await invoke("logout_cmd");
  debugLog("logout response", "ok");
}

export async function fetchUserProfile(): Promise<unknown> {
  return gatewayRequest("GET", "/v1/users/profile");
}

export async function listPosts(query: ListPostsQuery = {}): Promise<unknown> {
  return gatewayRequest("GET", "/v1/posts", {
    query: {
      "page.page_size": query.pageSize,
      "page.page_token": query.pageToken,
      category_id: query.categoryId,
      tag_id: query.tagId,
      author_id: query.authorId,
    },
  });
}

export async function changePassword(
  payload: ChangePasswordPayload,
): Promise<unknown> {
  return gatewayRequest("POST", "/v1/users/password", {
    body: {
      old_password: payload.oldPassword,
      new_password: payload.newPassword,
    },
  });
}

export async function updateUserProfile(
  payload: UpdateProfilePayload,
): Promise<unknown> {
  return gatewayRequest("PUT", "/v1/users/profile", {
    body: {
      nickname: payload.nickname,
      bio: payload.bio,
      avatar_key: payload.avatarKey,
      theme: payload.theme,
      language: payload.language,
    },
  });
}

export async function getPublicUser(userId: number): Promise<unknown> {
  return gatewayRequest("GET", `/v1/users/${userId}`, { authRequired: false });
}

export async function getPublicPost(slug: string): Promise<unknown> {
  return gatewayRequest("GET", `/v1/posts/${encodeURIComponent(slug)}`, {
    authRequired: false,
  });
}

export async function batchGetPosts(postIds: number[]): Promise<unknown> {
  return gatewayRequest("POST", "/v1/posts:batch", {
    authRequired: false,
    body: { post_ids: postIds },
  });
}

export async function listComments(
  postId: number,
  query: { pageSize?: number; pageToken?: string; rootId?: number } = {},
): Promise<unknown> {
  return gatewayRequest("GET", `/v1/posts/${postId}/comments`, {
    authRequired: false,
    query: {
      "page.page_size": query.pageSize,
      "page.page_token": query.pageToken,
      root_id: query.rootId,
    },
  });
}

export async function createComment(
  postId: number,
  payload: CreateCommentPayload,
): Promise<unknown> {
  return gatewayRequest("POST", `/v1/posts/${postId}/comments`, {
    body: {
      post_id: postId,
      parent_id: payload.parentId ?? 0,
      content: payload.content,
    },
  });
}

export async function deleteComment(commentId: number): Promise<unknown> {
  return gatewayRequest("DELETE", `/v1/comments/${commentId}`);
}

export async function listCategories(): Promise<unknown> {
  return gatewayRequest("GET", "/v1/categories");
}

export async function createCategory(
  payload: CreateCategoryPayload,
): Promise<unknown> {
  return gatewayRequest("POST", "/v1/categories", {
    body: {
      name: payload.name,
      slug: payload.slug,
      description: payload.description,
    },
  });
}

export async function updateCategory(
  categoryId: number | string,
  payload: UpdateCategoryPayload,
): Promise<unknown> {
  return gatewayRequest("PUT", `/v1/categories/${categoryId}`, {
    body: {
      category_id: payload.categoryId,
      name: payload.name,
      slug: payload.slug,
      description: payload.description,
    },
  });
}

export async function deleteCategory(
  categoryId: number | string,
): Promise<unknown> {
  return gatewayRequest("DELETE", `/v1/categories/${categoryId}`);
}

export async function listTags(): Promise<unknown> {
  return gatewayRequest("GET", "/v1/tags");
}

export async function deleteTag(tagId: number | string): Promise<unknown> {
  return gatewayRequest("DELETE", `/v1/tags/${tagId}`);
}

export async function listDrafts(
  query: ListDraftsQuery = {},
): Promise<unknown> {
  return gatewayRequest("GET", "/v1/drafts", {
    query: {
      page: query.page,
      page_size: query.pageSize,
      keyword: query.keyword,
      category_id: query.categoryId,
      status: query.status,
    },
  });
}

export async function createAdminPost(
  payload: CreatePostPayload,
): Promise<unknown> {
  return gatewayRequest("POST", "/v1/admin/posts", {
    body: {
      title: payload.title,
      slug: payload.slug,
      raw_markdown: payload.rawMarkdown,
      category_id: payload.categoryId,
      tag_names: payload.tagNames,
      status: payload.status,
      cover_image_key: payload.coverImageKey,
      resource_key: payload.resourceKey,
    },
  });
}

export async function getAdminPost(postId: number | string): Promise<unknown> {
  return gatewayRequest("GET", `/v1/admin/posts/${postId}`);
}

export async function updateAdminPost(
  postId: number | string,
  payload: UpdatePostPayload,
): Promise<unknown> {
  return gatewayRequest("PUT", `/v1/admin/posts/${postId}`, {
    body: {
      post_id: payload.postId,
      title: payload.title,
      raw_markdown: payload.rawMarkdown,
      status: payload.status,
      cover_image_key: payload.coverImageKey,
      resource_key: payload.resourceKey,
      category_id: payload.categoryId,
      tag_names: payload.tagNames,
    },
  });
}

export async function deleteAdminPost(
  postId: number | string,
): Promise<unknown> {
  return gatewayRequest("DELETE", `/v1/admin/posts/${postId}`);
}

export async function fetchAdminPostSource(
  postId: number | string,
): Promise<unknown> {
  return gatewayRequest("GET", `/v1/admin/posts/${postId}/source`);
}

export async function searchPosts(query: {
  q?: string;
  query?: string;
  page?: number;
  pageSize?: number;
  categoryId?: number;
  tagId?: number;
  authorId?: number;
}): Promise<unknown> {
  return gatewayRequest("GET", "/v1/search", {
    authRequired: false,
    query: {
      q: query.q,
      query: query.query,
      page: query.page,
      page_size: query.pageSize,
      category_id: query.categoryId,
      tag_id: query.tagId,
      author_id: query.authorId,
    },
  });
}

export async function searchSuggest(
  query: { prefix?: string; q?: string; limit?: number } = {},
): Promise<unknown> {
  return gatewayRequest("GET", "/v1/search/suggest", {
    authRequired: false,
    query: {
      prefix: query.prefix,
      q: query.q,
      limit: query.limit,
    },
  });
}

export async function getUploadTicket(
  payload: GetUploadTicketPayload,
): Promise<unknown> {
  return gatewayRequest("POST", "/v1/storage/upload_ticket", {
    body: {
      filename: payload.filename,
      usage: payload.usage,
    },
  });
}
