<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from "vue";
import {
    createAdminPost,
    deleteAdminPost,
    fetchAdminPostSource,
    getAdminPost,
    listCategories,
    listDrafts,
    listTags,
    updateAdminPost,
} from "@/services/helmApi";
import {
    normalizePostStatusLabel,
    parseListDraftsResponse,
    type IdLike,
    type PostRowViewModel,
} from "@/utils/postAdapter";
import { extractCategories, extractTags } from "@/utils/taxonomy";

interface CategoryOption {
    id: IdLike;
    name: string;
}

interface TagOption {
    id: IdLike;
    name: string;
}

const postStatusOptions = [
    { value: 1, label: "草稿" },
    { value: 2, label: "已发布" },
    { value: 3, label: "已归档" },
];

const queryStatusOptions = [
    { value: undefined, label: "全部状态" },
    ...postStatusOptions,
];

const pageSize = ref(10);
const page = ref(1);
const keyword = ref("");
const categoryId = ref<IdLike | null>(null);
const queryStatus = ref<number | undefined>(undefined);
const loading = ref(false);
const result = ref("等待查询");
const posts = ref<PostRowViewModel[]>([]);
const categories = ref<CategoryOption[]>([]);
const tags = ref<TagOption[]>([]);
const totalCount = ref(0);
const selectedPostId = ref<IdLike | null>(null);
const editorOpen = ref(false);
const editingPostId = ref<IdLike | null>(null);
const createTitle = ref("");
const createSlug = ref("");
const createMarkdown = ref("# 新文章\n\n在这里开始编写内容");
const createStatus = ref<number>(1);
const createCategoryId = ref<IdLike | null>(null);
const createTagNames = ref("");
const editTitle = ref("");
const editSlug = ref("");
const editMarkdown = ref("");
const editStatus = ref<number | null>(null);
const editCategoryId = ref<IdLike | null>(null);
const editTagNames = ref("");

function asObj(data: unknown): Record<string, unknown> {
    if (!data || typeof data !== "object") {
        return {};
    }
    return data as Record<string, unknown>;
}

function normalizeCategories(data: unknown): void {
    categories.value = extractCategories(data).map((item) => ({
        id: item.id,
        name: item.name,
    }));
}

function normalizeTags(data: unknown): void {
    tags.value = extractTags(data).map((item) => ({
        id: item.id,
        name: item.name,
    }));
}

async function loadFilters(): Promise<void> {
    try {
        const [categoryData, tagData] = await Promise.all([listCategories(), listTags()]);
        normalizeCategories(categoryData);
        normalizeTags(tagData);
    } catch {
        categories.value = [];
        tags.value = [];
    }
}

function normalizeListResponse(data: unknown): void {
    const normalized = parseListDraftsResponse(data);
    posts.value = normalized.posts;
    totalCount.value = normalized.totalCount;
}

function formatTime(ts?: number | null): string {
    if (!ts || ts <= 0) {
        return "-";
    }
    return new Date(ts).toLocaleString();
}

function getPostStatusLabel(status?: number | string): string {
    return normalizePostStatusLabel(status);
}

function getTagNames(row: PostRowViewModel): string[] {
    return row.tagNames;
}

function slugify(input: string): string {
    return input
        .trim()
        .toLowerCase()
        .replace(/[^a-z0-9\u4e00-\u9fa5\s-]/g, "")
        .replace(/\s+/g, "-")
        .replace(/-+/g, "-");
}

function syncSlugFromTitle(): void {
    if (!createTitle.value.trim()) {
        return;
    }

    createSlug.value = slugify(createTitle.value);
}

async function queryPosts(): Promise<void> {
    loading.value = true;
    try {
        const data = await listDrafts({
            page: page.value,
            pageSize: pageSize.value,
            keyword: keyword.value || undefined,
            categoryId: categoryId.value ?? undefined,
            status: queryStatus.value,
        });
        normalizeListResponse(data);
        result.value = JSON.stringify(data, null, 2);
    } catch (error) {
        posts.value = [];
        totalCount.value = 0;
        result.value = `查询失败: ${String(error)}`;
    } finally {
        loading.value = false;
    }
}

async function queryNextPage(): Promise<void> {
    page.value = page.value + 1;
    await queryPosts();
}

async function createPost(): Promise<void> {
    loading.value = true;
    try {
        const data = await createAdminPost({
            title: createTitle.value,
            slug: createSlug.value || slugify(createTitle.value),
            rawMarkdown: createMarkdown.value,
            categoryId: createCategoryId.value ?? undefined,
            tagNames: splitTags(createTagNames.value),
            status: createStatus.value,
        });
        result.value = JSON.stringify(data, null, 2);
        selectedPostId.value = null;
        createTitle.value = "";
        createSlug.value = "";
        createMarkdown.value = "# 新文章\n\n在这里开始编写内容";
        createStatus.value = 1;
        createCategoryId.value = null;
        createTagNames.value = "";
    } catch (error) {
        result.value = `创建失败: ${String(error)}`;
    } finally {
        loading.value = false;
    }
}

function resetFilters(): void {
    page.value = 1;
    keyword.value = "";
    categoryId.value = null;
    queryStatus.value = undefined;
    posts.value = [];
    totalCount.value = 0;
    result.value = "等待查询";
}

function splitTags(csv: string): string[] {
    return csv
        .split(",")
        .map((x) => x.trim())
        .filter((x) => x.length > 0);
}

function toggleTagValue(currentValue: string, tagName: string): string {
    const current = splitTags(currentValue);
    const exists = current.includes(tagName);
    const next = exists
        ? current.filter((item) => item !== tagName)
        : [...current, tagName];
    return next.join(", ");
}

function toggleCreateTag(tagName: string): void {
    createTagNames.value = toggleTagValue(createTagNames.value, tagName);
}

function toggleEditTag(tagName: string): void {
    editTagNames.value = toggleTagValue(editTagNames.value, tagName);
}

function hasTag(currentValue: string, tagName: string): boolean {
    return splitTags(currentValue).includes(tagName);
}

function normalizeAdminPost(data: unknown): Record<string, unknown> {
    const obj = asObj(data);
    if (obj.post && typeof obj.post === "object") {
        return obj.post as Record<string, unknown>;
    }
    return obj;
}

function parseStatusValue(value: unknown): number | null {
    if (typeof value === "number") {
        return value;
    }
    if (typeof value === "string") {
        const normalized = value.trim().toUpperCase();
        if (normalized === "POST_STATUS_DRAFT" || normalized === "DRAFT") {
            return 1;
        }
        if (normalized === "POST_STATUS_PUBLISHED" || normalized === "PUBLISHED") {
            return 2;
        }
        if (normalized === "POST_STATUS_ARCHIVED" || normalized === "ARCHIVED") {
            return 3;
        }
        const parsed = Number(value);
        return Number.isFinite(parsed) ? parsed : null;
    }
    return null;
}

function parseCategoryId(value: unknown): IdLike | null {
    if (typeof value === "number" && Number.isFinite(value)) {
        return value;
    }
    if (typeof value === "string") {
        const normalized = value.trim();
        return normalized.length > 0 ? normalized : null;
    }
    return null;
}

async function openEditor(postId: IdLike): Promise<void> {
    loading.value = true;
    selectedPostId.value = postId;
    editingPostId.value = postId;
    try {
        const data = await getAdminPost(postId);
        const post = normalizeAdminPost(data);

        editTitle.value = typeof post.title === "string" ? post.title : "";
        editSlug.value = typeof post.slug === "string" ? post.slug : "";
        editMarkdown.value =
            typeof post.raw_markdown === "string"
                ? post.raw_markdown
                : typeof post.rawMarkdown === "string"
                    ? (post.rawMarkdown as string)
                    : "";

        editStatus.value = parseStatusValue(post.status);
        editCategoryId.value =
            parseCategoryId(post.category_id) ??
            parseCategoryId(post.categoryId) ??
            parseCategoryId(asObj(post.category).id);

        const tagNames = Array.isArray(post.tag_names)
            ? (post.tag_names as unknown[])
            : Array.isArray(post.tagNames)
                ? (post.tagNames as unknown[])
                : [];
        editTagNames.value = tagNames
            .map((x) => (typeof x === "string" ? x : ""))
            .filter((x) => x.length > 0)
            .join(", ");

        editorOpen.value = true;
        result.value = JSON.stringify(data, null, 2);
    } catch (error) {
        result.value = `打开编辑器失败: ${String(error)}`;
    } finally {
        loading.value = false;
    }
}

function closeEditor(): void {
    if (loading.value) {
        return;
    }
    editorOpen.value = false;
}

function onEscClose(event: KeyboardEvent): void {
    if (event.key === "Escape" && editorOpen.value) {
        closeEditor();
    }
}

onMounted(() => {
    window.addEventListener("keydown", onEscClose);
    void loadFilters();
});

onBeforeUnmount(() => {
    window.removeEventListener("keydown", onEscClose);
});

async function saveEditor(): Promise<void> {
    if (!editingPostId.value) {
        return;
    }

    loading.value = true;
    try {
        const data = await updateAdminPost(editingPostId.value, {
            postId: editingPostId.value,
            title: editTitle.value || undefined,
            rawMarkdown: editMarkdown.value || undefined,
            status: editStatus.value ?? undefined,
            categoryId: editCategoryId.value ?? undefined,
            tagNames: splitTags(editTagNames.value),
        });
        result.value = JSON.stringify(data, null, 2);

        const idx = posts.value.findIndex((x) => x.id === editingPostId.value);
        if (idx >= 0) {
            posts.value[idx] = {
                ...posts.value[idx],
                title: editTitle.value || posts.value[idx].title,
                slug: editSlug.value || posts.value[idx].slug,
            };
        }

        editorOpen.value = false;
    } catch (error) {
        result.value = `保存失败: ${String(error)}`;
    } finally {
        loading.value = false;
    }
}

async function viewAdminDetail(postId: IdLike): Promise<void> {
    loading.value = true;
    selectedPostId.value = postId;
    try {
        const data = await getAdminPost(postId);
        result.value = JSON.stringify(data, null, 2);
    } catch (error) {
        result.value = `读取详情失败: ${String(error)}`;
    } finally {
        loading.value = false;
    }
}

async function viewAdminSource(postId: IdLike): Promise<void> {
    loading.value = true;
    selectedPostId.value = postId;
    try {
        const data = await fetchAdminPostSource(postId);
        result.value = JSON.stringify(data, null, 2);
    } catch (error) {
        result.value = `读取源码失败: ${String(error)}`;
    } finally {
        loading.value = false;
    }
}

async function removePost(postId: IdLike): Promise<void> {
    const ok = window.confirm(`确认删除文章 #${postId} ?`);
    if (!ok) {
        return;
    }

    loading.value = true;
    selectedPostId.value = postId;
    try {
        const data = await deleteAdminPost(postId);
        result.value = JSON.stringify(data, null, 2);
        posts.value = posts.value.filter((post) => post.id !== postId);
        totalCount.value = Math.max(0, totalCount.value - 1);
    } catch (error) {
        result.value = `删除失败: ${String(error)}`;
    } finally {
        loading.value = false;
    }
}
</script>

<template>
    <section class="mx-auto max-w-6xl space-y-4">
        <header class="panel p-5">
            <h2 class="text-xl font-semibold text-slate-900">文章管理</h2>
            <p class="mt-2 text-sm text-slate-600">补齐文章创建、管理草稿列表查询、详情、源码、编辑与删除等后台能力。</p>
        </header>

        <section class="panel p-5">
            <div class="mb-4 flex items-center justify-between">
                <div>
                    <h3 class="text-sm font-semibold text-slate-700">新建文章</h3>
                    <p class="mt-1 text-xs text-slate-500">这里直接调用后台创建接口，不再只停留在实验页。</p>
                </div>
                <button
                    class="rounded-lg bg-slate-100 px-3 py-2 text-sm font-medium text-slate-700 hover:bg-slate-200 disabled:opacity-50"
                    :disabled="loading || !createTitle.trim()" @click="syncSlugFromTitle">
                    由标题生成 slug
                </button>
            </div>

            <div class="grid gap-3 md:grid-cols-2">
                <label>
                    <span class="mb-1 block text-xs font-medium text-slate-600">title</span>
                    <input v-model="createTitle" class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm"
                        type="text" placeholder="输入文章标题" />
                </label>
                <label>
                    <span class="mb-1 block text-xs font-medium text-slate-600">slug</span>
                    <input v-model="createSlug" class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm"
                        type="text" placeholder="my-first-post" />
                </label>
                <label>
                    <span class="mb-1 block text-xs font-medium text-slate-600">status</span>
                    <select v-model.number="createStatus"
                        class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm">
                        <option v-for="item in postStatusOptions" :key="item.value" :value="item.value">{{ item.label }}
                        </option>
                    </select>
                </label>
                <label>
                    <span class="mb-1 block text-xs font-medium text-slate-600">category_id</span>
                    <select v-model="createCategoryId"
                        class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm">
                        <option :value="null">未选择分类</option>
                        <option v-for="item in categories" :key="item.id" :value="item.id">{{ item.name }}</option>
                    </select>
                </label>
                <label class="md:col-span-2">
                    <span class="mb-1 block text-xs font-medium text-slate-600">tag_names</span>
                    <input v-model="createTagNames" class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm"
                        type="text" placeholder="go, rust, bifrost" />
                    <div class="mt-2 flex flex-wrap gap-2">
                        <button v-for="item in tags" :key="item.id" type="button"
                            class="rounded-full border px-3 py-1 text-xs transition-colors" :class="hasTag(createTagNames, item.name)
                                ? 'border-steel-600 bg-steel-100 text-steel-700'
                                : 'border-slate-200 bg-white text-slate-600 hover:border-slate-300 hover:bg-slate-50'"
                            @click="toggleCreateTag(item.name)">
                            {{ item.name }}
                        </button>
                    </div>
                </label>
                <label class="md:col-span-2">
                    <span class="mb-1 block text-xs font-medium text-slate-600">raw_markdown</span>
                    <textarea v-model="createMarkdown"
                        class="min-h-[220px] w-full rounded-lg border border-slate-300 px-3 py-2 text-sm"
                        placeholder="请输入 markdown" />
                </label>
            </div>

            <div class="mt-4 flex gap-2">
                <button
                    class="rounded-lg bg-steel-600 px-4 py-2 text-sm font-medium text-white hover:bg-steel-700 disabled:opacity-50"
                    :disabled="loading || !createTitle.trim() || !createMarkdown.trim()" @click="createPost">
                    创建文章
                </button>
            </div>
        </section>

        <section class="panel p-5">
            <div class="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
                <label>
                    <span class="mb-1 block text-xs font-medium text-slate-600">page_size</span>
                    <input v-model.number="pageSize" class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm"
                        type="number" min="1" max="100" />
                </label>
                <label>
                    <span class="mb-1 block text-xs font-medium text-slate-600">page</span>
                    <input v-model.number="page" class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm"
                        type="number" min="1" />
                </label>
                <label>
                    <span class="mb-1 block text-xs font-medium text-slate-600">category_id</span>
                    <select v-model="categoryId" class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm">
                        <option :value="null">全部分类</option>
                        <option v-for="item in categories" :key="item.id" :value="item.id">{{ item.name }}</option>
                    </select>
                </label>
                <label>
                    <span class="mb-1 block text-xs font-medium text-slate-600">status</span>
                    <select v-model="queryStatus" class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm">
                        <option v-for="item in queryStatusOptions" :key="item.label" :value="item.value">{{ item.label
                        }}</option>
                    </select>
                </label>
                <label class="md:col-span-2 xl:col-span-1">
                    <span class="mb-1 block text-xs font-medium text-slate-600">keyword</span>
                    <input v-model="keyword" class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm"
                        type="text" placeholder="标题或正文关键词" />
                </label>
            </div>

            <div class="mt-4 flex gap-2">
                <button
                    class="rounded-lg bg-steel-600 px-4 py-2 text-sm font-medium text-white hover:bg-steel-700 disabled:opacity-50"
                    :disabled="loading" @click="queryPosts">查询文章</button>
                <button
                    class="rounded-lg bg-slate-100 px-4 py-2 text-sm font-medium text-slate-700 hover:bg-slate-200 disabled:opacity-50"
                    :disabled="loading" @click="queryNextPage">下一页</button>
                <button
                    class="rounded-lg bg-slate-100 px-4 py-2 text-sm font-medium text-slate-700 hover:bg-slate-200 disabled:opacity-50"
                    :disabled="loading" @click="resetFilters">重置</button>
            </div>
        </section>

        <section class="panel p-5">
            <div class="mb-3 flex items-center justify-between">
                <h3 class="text-sm font-semibold text-slate-700">文章结果</h3>
                <div class="text-xs text-slate-500">
                    total: {{ totalCount }}
                    <span class="ml-2 rounded bg-steel-100 px-2 py-0.5 text-steel-700">page {{ page }}</span>
                </div>
            </div>

            <div class="overflow-x-auto rounded-lg border border-slate-200">
                <table class="min-w-full border-collapse text-sm">
                    <thead class="bg-slate-100 text-left text-xs uppercase tracking-wide text-slate-600">
                        <tr>
                            <th class="px-3 py-2">ID</th>
                            <th class="px-3 py-2">标题</th>
                            <th class="px-3 py-2">Slug</th>
                            <th class="px-3 py-2">作者</th>
                            <th class="px-3 py-2">分类</th>
                            <th class="px-3 py-2">状态</th>
                            <th class="px-3 py-2">标签</th>
                            <th class="px-3 py-2">发布时间</th>
                            <th class="px-3 py-2">操作</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-if="posts.length === 0">
                            <td colspan="9" class="px-3 py-6 text-center text-slate-500">暂无数据，先执行查询。</td>
                        </tr>
                        <tr v-for="row in posts" :key="row.id" class="border-t border-slate-200">
                            <td class="px-3 py-2 text-slate-700">{{ row.id }}</td>
                            <td class="px-3 py-2 font-medium text-slate-900">{{ row.title }}</td>
                            <td class="px-3 py-2 font-mono text-xs text-slate-600">{{ row.slug }}</td>
                            <td class="px-3 py-2 text-slate-700">{{ row.author?.nickname || "-" }}</td>
                            <td class="px-3 py-2 text-slate-700">{{ row.category?.name || "-" }}</td>
                            <td class="px-3 py-2 text-slate-700">
                                <span class="rounded-full bg-slate-100 px-2 py-1 text-xs">{{
                                    getPostStatusLabel(row.status) }}</span>
                            </td>
                            <td class="px-3 py-2 text-slate-700">
                                <div class="flex flex-wrap gap-1">
                                    <span v-for="tagName in getTagNames(row)" :key="tagName"
                                        class="rounded-full bg-steel-50 px-2 py-1 text-xs text-steel-700">{{ tagName
                                        }}</span>
                                    <span v-if="getTagNames(row).length === 0">-</span>
                                </div>
                            </td>
                            <td class="px-3 py-2 text-slate-700">{{ formatTime(row.publishedAt) }}</td>
                            <td class="px-3 py-2">
                                <div class="flex gap-1">
                                    <button
                                        class="rounded bg-slate-100 px-2 py-1 text-xs text-slate-700 hover:bg-slate-200 disabled:opacity-50"
                                        :disabled="loading" @click="viewAdminDetail(row.id)">
                                        详情
                                    </button>
                                    <button
                                        class="rounded bg-steel-100 px-2 py-1 text-xs text-steel-700 hover:bg-steel-200 disabled:opacity-50"
                                        :disabled="loading" @click="viewAdminSource(row.id)">
                                        源码
                                    </button>
                                    <button
                                        class="rounded bg-amber-100 px-2 py-1 text-xs text-amber-700 hover:bg-amber-200 disabled:opacity-50"
                                        :disabled="loading" @click="openEditor(row.id)">
                                        编辑
                                    </button>
                                    <button
                                        class="rounded bg-rose-100 px-2 py-1 text-xs text-rose-700 hover:bg-rose-200 disabled:opacity-50"
                                        :disabled="loading" @click="removePost(row.id)">
                                        删除
                                    </button>
                                </div>
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </section>

        <Teleport to="body">
            <div v-if="editorOpen" class="fixed inset-0 z-50 flex">
                <button class="h-full flex-1 border-none bg-slate-950/35" aria-label="close drawer" :disabled="loading"
                    @click="closeEditor" />

                <section
                    class="h-full w-full max-w-2xl overflow-auto border-l border-slate-200 bg-white p-5 shadow-2xl">
                    <div class="mb-4 flex items-center justify-between">
                        <h3 class="text-sm font-semibold text-slate-700">编辑文章 #{{ editingPostId }}</h3>
                        <button
                            class="rounded bg-slate-100 px-2 py-1 text-xs text-slate-700 hover:bg-slate-200 disabled:opacity-50"
                            :disabled="loading" @click="closeEditor">
                            关闭
                        </button>
                    </div>

                    <div class="grid gap-3 md:grid-cols-2">
                        <label>
                            <span class="mb-1 block text-xs font-medium text-slate-600">title</span>
                            <input v-model="editTitle"
                                class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm" type="text"
                                placeholder="文章标题" />
                        </label>

                        <label>
                            <span class="mb-1 block text-xs font-medium text-slate-600">slug (只读展示)</span>
                            <input v-model="editSlug"
                                class="w-full rounded-lg border border-slate-200 bg-slate-100 px-3 py-2 text-sm text-slate-500"
                                type="text" readonly />
                        </label>

                        <label>
                            <span class="mb-1 block text-xs font-medium text-slate-600">status</span>
                            <select v-model.number="editStatus"
                                class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm">
                                <option v-for="item in postStatusOptions" :key="item.value" :value="item.value">{{
                                    item.label }}</option>
                            </select>
                        </label>

                        <label>
                            <span class="mb-1 block text-xs font-medium text-slate-600">category_id</span>
                            <select v-model="editCategoryId"
                                class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm">
                                <option :value="null">未选择分类</option>
                                <option v-for="item in categories" :key="item.id" :value="item.id">{{ item.name }}
                                </option>
                            </select>
                        </label>

                        <label class="md:col-span-2">
                            <span class="mb-1 block text-xs font-medium text-slate-600">tag_names (comma
                                separated)</span>
                            <input v-model="editTagNames"
                                class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm" type="text"
                                placeholder="go, rust, bifrost" />
                            <div class="mt-2 flex flex-wrap gap-2">
                                <button v-for="item in tags" :key="item.id" type="button"
                                    class="rounded-full border px-3 py-1 text-xs transition-colors"
                                    :class="hasTag(editTagNames, item.name)
                                        ? 'border-steel-600 bg-steel-100 text-steel-700'
                                        : 'border-slate-200 bg-white text-slate-600 hover:border-slate-300 hover:bg-slate-50'" @click="toggleEditTag(item.name)">
                                    {{ item.name }}
                                </button>
                            </div>
                        </label>

                        <label class="md:col-span-2">
                            <span class="mb-1 block text-xs font-medium text-slate-600">raw_markdown</span>
                            <textarea v-model="editMarkdown"
                                class="min-h-[260px] w-full rounded-lg border border-slate-300 px-3 py-2 text-sm"
                                placeholder="请输入 markdown" />
                        </label>
                    </div>

                    <div class="mt-4 flex gap-2">
                        <button
                            class="rounded-lg bg-steel-600 px-4 py-2 text-sm font-medium text-white hover:bg-steel-700 disabled:opacity-50"
                            :disabled="loading" @click="saveEditor">
                            保存更新
                        </button>
                        <button
                            class="rounded-lg bg-slate-100 px-4 py-2 text-sm font-medium text-slate-700 hover:bg-slate-200 disabled:opacity-50"
                            :disabled="loading" @click="closeEditor">
                            取消
                        </button>
                    </div>
                </section>
            </div>
        </Teleport>

        <section class="panel p-5">
            <h3 class="text-sm font-semibold text-slate-700">响应结果</h3>
            <p v-if="selectedPostId" class="mt-2 text-xs text-slate-500">当前操作文章 ID: {{ selectedPostId }}</p>
            <pre
                class="mt-3 max-h-[420px] overflow-auto rounded-lg bg-slate-950 p-4 text-xs text-emerald-200">{{ result }}</pre>
        </section>
    </section>
</template>
