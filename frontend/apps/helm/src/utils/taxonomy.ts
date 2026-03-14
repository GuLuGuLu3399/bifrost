export type IdLike = number | string;

export interface TaxonomyCategoryRow {
  id: IdLike;
  name: string;
  slug: string;
  description?: string;
}

export interface TaxonomyTagRow {
  id: IdLike;
  name: string;
  post_count?: number;
}

function asRecord(data: unknown): Record<string, unknown> {
  return data && typeof data === "object"
    ? (data as Record<string, unknown>)
    : {};
}

function pickArray(data: unknown, keys: string[]): unknown[] {
  const obj = asRecord(data);
  for (const key of keys) {
    const value = obj[key];
    if (Array.isArray(value)) {
      return value;
    }
  }
  return [];
}

function toId(value: unknown): IdLike {
  if (typeof value === "number" && Number.isFinite(value) && value > 0) {
    return value;
  }
  if (typeof value === "string") {
    const normalized = value.trim();
    if (!/^\d+$/.test(normalized)) {
      return "";
    }

    const parsed = Number(normalized);
    if (Number.isFinite(parsed) && parsed <= Number.MAX_SAFE_INTEGER) {
      return parsed;
    }
    return normalized;
  }
  return "";
}

function isPositiveId(id: IdLike): boolean {
  if (typeof id === "number") {
    return Number.isFinite(id) && id > 0;
  }

  if (!/^\d+$/.test(id)) {
    return false;
  }

  try {
    return BigInt(id) > 0n;
  } catch {
    return false;
  }
}

export function extractCategories(data: unknown): TaxonomyCategoryRow[] {
  const items = pickArray(data, ["categories", "items"]);
  return items
    .map((item) => asRecord(item))
    .map((item) => {
      const id = toId(item.id ?? item.category_id);
      const name =
        typeof item.name === "string" ? item.name : `分类 ${String(id || "")}`;
      const slug =
        typeof item.slug === "string" && item.slug.trim().length > 0
          ? item.slug
          : String(item.slug ?? "");
      const description =
        typeof item.description === "string" ? item.description : undefined;
      return { id, name, slug, description };
    })
    .filter((item) => isPositiveId(item.id));
}

export function extractTags(data: unknown): TaxonomyTagRow[] {
  const items = pickArray(data, ["tags", "items"]);
  return items
    .map((item) => asRecord(item))
    .map((item) => {
      const id = toId(item.id ?? item.tag_id);
      const name =
        typeof item.name === "string" ? item.name : `标签 ${String(id || "")}`;
      const postCountRaw = item.post_count;
      const post_count =
        typeof postCountRaw === "number" ? postCountRaw : undefined;
      return { id, name, post_count };
    })
    .filter((item) => isPositiveId(item.id));
}
