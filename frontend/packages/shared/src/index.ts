export type JsonValue =
  | string
  | number
  | boolean
  | null
  | JsonValue[]
  | { [key: string]: JsonValue };

export type Branded<T, Brand extends string> = T & { __brand: Brand };

// ============================================================================
// API Types for Beacon & Mirror Services
// ============================================================================

// Author info
export interface AuthorInfo {
  id: string;
  nickname: string;
  avatar_key?: string;
}

// Category item
export interface CategoryItem {
  id: string;
  name: string;
  slug: string;
  description?: string;
  post_count?: number;
}

// Tag item
export interface TagItem {
  id: string;
  name: string;
  slug: string;
  post_count?: number;
}

// Post summary (for listings)
export interface PostSummary {
  id: string;
  title: string;
  slug: string;
  summary?: string;
  cover_image_key?: string;
  view_count?: number;
  like_count?: number;
  published_at?: string;
  updated_at?: string;
  author?: AuthorInfo;
  category?: CategoryItem;
  tags?: TagItem[];
}

// Post detail (for full article page)
export interface PostDetail extends PostSummary {
  html_body: string;
  toc_json?: JsonValue; // Table of contents
  prev_post_id?: string;
  prev_post_slug?: string;
  prev_post_title?: string;
  next_post_id?: string;
  next_post_slug?: string;
  next_post_title?: string;
}

// Page response
export interface PageResponse {
  next_page_token?: string;
  total_count?: number;
}

// API response for posts list
export interface PostsListResponse {
  posts: PostSummary[];
  page: PageResponse;
}

// API response for post detail
export interface PostDetailResponse {
  post: PostDetail;
}

// Comment item
export interface CommentItem {
  id: string;
  post_id: string;
  content: string;
  author?: AuthorInfo;
  created_at?: string;
  updated_at?: string;
  reply_count?: number;
  parent_id?: string;
}

// Search hit
export interface SearchHit {
  id: string;
  score: number;
  title: string;
  slug: string;
  highlight_title?: string;
  highlight_content?: string;
  published_at?: string;
}

// Search facets
export interface SearchFacets {
  categories?: Record<string, number>;
  tags?: Record<string, number>;
}

// Search response
export interface SearchResponse {
  hits: SearchHit[];
  total_hits: number;
  took_ms?: number;
  facets?: SearchFacets;
  page: PageResponse;
}
