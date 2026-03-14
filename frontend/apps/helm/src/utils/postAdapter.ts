export type IdLike = string | number;

export interface PostRowViewModel {
  id: IdLike;
  title: string;
  slug: string;
  status?: number | string;
  author?: { nickname?: string } | null;
  category?: { name?: string } | null;
  publishedAt?: number | null;
  tagNames: string[];
}

interface ListDraftsViewModel {
  posts: PostRowViewModel[];
  totalCount: number;
}

function asRecord(data: unknown): Record<string, unknown> {
  return data && typeof data === "object"
    ? (data as Record<string, unknown>)
    : {};
}

function toId(value: unknown): IdLike {
  if (typeof value === "string" || typeof value === "number") {
    return value;
  }
  return "";
}

function normalizeTagNames(data: unknown): string[] {
  if (Array.isArray(data)) {
    return data
      .map((item) => {
        if (typeof item === "string") {
          return item;
        }
        const obj = asRecord(item);
        return typeof obj.name === "string" ? obj.name : "";
      })
      .filter((item) => item.length > 0);
  }
  return [];
}

function timestampToMs(value: unknown): number | null {
  if (typeof value === "number") {
    if (!Number.isFinite(value) || value <= 0) {
      return null;
    }
    return value > 1_000_000_000_000 ? value : value * 1000;
  }

  if (typeof value === "string") {
    const parsed = Number(value);
    if (Number.isFinite(parsed) && parsed > 0) {
      return parsed > 1_000_000_000_000 ? parsed : parsed * 1000;
    }
    return null;
  }

  const obj = asRecord(value);
  const seconds = obj.seconds;
  if (typeof seconds === "number" && Number.isFinite(seconds) && seconds > 0) {
    const nanos =
      typeof obj.nanos === "number" && Number.isFinite(obj.nanos)
        ? obj.nanos
        : 0;
    return seconds * 1000 + Math.floor(nanos / 1_000_000);
  }

  return null;
}

export function parseListDraftsResponse(data: unknown): ListDraftsViewModel {
  const obj = asRecord(data);
  const rawPosts = Array.isArray(obj.posts) ? obj.posts : [];

  const posts = rawPosts.map((item) => {
    const post = asRecord(item);
    const id = toId(post.id);
    const publishedAt = timestampToMs(post.published_at ?? post.publishedAt);

    const directTagNames = post.tag_names ?? post.tagNames;
    const nestedTags = post.tags;
    const tagNames = normalizeTagNames(directTagNames);
    const finalTagNames =
      tagNames.length > 0 ? tagNames : normalizeTagNames(nestedTags);

    return {
      id,
      title: typeof post.title === "string" ? post.title : "",
      slug: typeof post.slug === "string" ? post.slug : "",
      status:
        typeof post.status === "string" || typeof post.status === "number"
          ? post.status
          : undefined,
      author: (post.author as { nickname?: string } | null | undefined) ?? null,
      category: (post.category as { name?: string } | null | undefined) ?? null,
      publishedAt,
      tagNames: finalTagNames,
    } satisfies PostRowViewModel;
  });

  const totalCountRaw = obj.total_count ?? obj.totalCount;
  const totalCount =
    typeof totalCountRaw === "number" ? totalCountRaw : posts.length;

  return { posts, totalCount };
}

export function normalizePostStatusLabel(status?: number | string): string {
  if (typeof status === "string") {
    const normalized = status.trim().toUpperCase();
    if (normalized === "POST_STATUS_DRAFT" || normalized === "DRAFT") {
      return "草稿";
    }
    if (normalized === "POST_STATUS_PUBLISHED" || normalized === "PUBLISHED") {
      return "已发布";
    }
    if (normalized === "POST_STATUS_ARCHIVED" || normalized === "ARCHIVED") {
      return "已归档";
    }

    const parsed = Number(status);
    if (Number.isFinite(parsed)) {
      return normalizePostStatusLabel(parsed);
    }
    return "未知状态";
  }

  if (status === 1) {
    return "草稿";
  }
  if (status === 2) {
    return "已发布";
  }
  if (status === 3) {
    return "已归档";
  }

  return "未知状态";
}
