<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import {
    createCategory,
    deleteCategory,
    deleteTag,
    listCategories,
    listTags,
    updateCategory,
} from "@/services/helmApi";
import { extractCategories, extractTags, type IdLike } from "@/utils/taxonomy";

interface CategoryRow {
    id: IdLike;
    name: string;
    slug: string;
    description?: string;
}

interface TagRow {
    id: IdLike;
    name: string;
    post_count?: number;
}

const loading = ref(false);
const result = ref("等待加载");
const categories = ref<CategoryRow[]>([]);
const tags = ref<TagRow[]>([]);
const editingCategoryId = ref<IdLike | null>(null);
const name = ref("");
const slug = ref("");
const description = ref("");
const router = useRouter();
const route = useRoute();

function extractErrorText(error: unknown): string {
    return error instanceof Error ? error.message : String(error);
}

async function handleAuthError(error: unknown): Promise<boolean> {
    const message = extractErrorText(error).toLowerCase();
    const needLogin =
        message.includes("missing authorization token") ||
        message.includes("401") ||
        message.includes("unauthorized");

    if (!needLogin) {
        return false;
    }

    result.value = "登录状态失效，请重新登录后再创建分类。";
    await router.replace({
        path: "/auth",
        query: {
            mode: "login",
            redirect: route.fullPath,
        },
    });
    return true;
}

function applyCategoryForm(row?: CategoryRow): void {
    editingCategoryId.value = row?.id ?? null;
    name.value = row?.name ?? "";
    slug.value = row?.slug ?? "";
    description.value = row?.description ?? "";
}

async function loadTaxonomy(): Promise<void> {
    loading.value = true;
    try {
        const [categoryResp, tagResp] = await Promise.all([listCategories(), listTags()]);
        categories.value = extractCategories(categoryResp);
        tags.value = extractTags(tagResp);
        result.value = JSON.stringify({ categoryResp, tagResp }, null, 2);
    } catch (error) {
        if (await handleAuthError(error)) {
            return;
        }
        result.value = `加载失败: ${String(error)}`;
    } finally {
        loading.value = false;
    }
}

async function saveCategory(): Promise<void> {
    loading.value = true;
    try {
        const payload = {
            name: name.value,
            slug: slug.value,
            description: description.value || undefined,
        };

        const data = editingCategoryId.value
            ? await updateCategory(editingCategoryId.value, {
                categoryId: editingCategoryId.value,
                ...payload,
            })
            : await createCategory(payload);

        result.value = JSON.stringify(data, null, 2);
        applyCategoryForm();
        await loadTaxonomy();
    } catch (error) {
        if (await handleAuthError(error)) {
            return;
        }
        result.value = `保存失败: ${String(error)}`;
    } finally {
        loading.value = false;
    }
}

async function removeCategory(id: IdLike): Promise<void> {
    if (!window.confirm(`确认删除分类 #${id} ?`)) {
        return;
    }

    loading.value = true;
    try {
        const data = await deleteCategory(id);
        result.value = JSON.stringify(data, null, 2);
        await loadTaxonomy();
    } catch (error) {
        if (await handleAuthError(error)) {
            return;
        }
        result.value = `删除分类失败: ${String(error)}`;
    } finally {
        loading.value = false;
    }
}

async function removeTag(id: IdLike): Promise<void> {
    if (!window.confirm(`确认删除标签 #${id} ?`)) {
        return;
    }

    loading.value = true;
    try {
        const data = await deleteTag(id);
        result.value = JSON.stringify(data, null, 2);
        await loadTaxonomy();
    } catch (error) {
        if (await handleAuthError(error)) {
            return;
        }
        result.value = `删除标签失败: ${String(error)}`;
    } finally {
        loading.value = false;
    }
}

onMounted(() => {
    void loadTaxonomy();
});
</script>

<template>
    <section class="mx-auto max-w-6xl space-y-4">
        <header class="panel p-5">
            <h2 class="text-xl font-semibold text-slate-900">分类与标签</h2>
            <p class="mt-2 text-sm text-slate-600">分类使用正式 CRUD 页面，标签先提供列表与删除能力。</p>
        </header>

        <section class="panel p-5">
            <div class="mb-4 flex items-center justify-between">
                <h3 class="text-sm font-semibold text-slate-700">
                    {{ editingCategoryId ? `编辑分类 #${editingCategoryId}` : "新建分类" }}
                </h3>
                <button
                    class="rounded-lg bg-slate-100 px-3 py-2 text-sm font-medium text-slate-700 hover:bg-slate-200 disabled:opacity-50"
                    :disabled="loading" @click="applyCategoryForm()">清空表单</button>
            </div>

            <div class="grid gap-3 md:grid-cols-2">
                <label>
                    <span class="mb-1 block text-xs font-medium text-slate-600">name</span>
                    <input v-model="name" class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm"
                        type="text" />
                </label>
                <label>
                    <span class="mb-1 block text-xs font-medium text-slate-600">slug</span>
                    <input v-model="slug" class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm"
                        type="text" />
                </label>
                <label class="md:col-span-2">
                    <span class="mb-1 block text-xs font-medium text-slate-600">description</span>
                    <textarea v-model="description" class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm"
                        rows="4" />
                </label>
            </div>

            <div class="mt-4 flex gap-2">
                <button
                    class="rounded-lg bg-steel-600 px-4 py-2 text-sm font-medium text-white hover:bg-steel-700 disabled:opacity-50"
                    :disabled="loading" @click="saveCategory">{{ editingCategoryId ? "保存分类" : "创建分类" }}</button>
                <button
                    class="rounded-lg bg-slate-100 px-4 py-2 text-sm font-medium text-slate-700 hover:bg-slate-200 disabled:opacity-50"
                    :disabled="loading" @click="loadTaxonomy">刷新列表</button>
            </div>
        </section>

        <div class="grid gap-4 lg:grid-cols-2">
            <section class="panel p-5">
                <div class="mb-3 flex items-center justify-between">
                    <h3 class="text-sm font-semibold text-slate-700">分类列表</h3>
                    <span class="text-xs text-slate-500">{{ categories.length }} 项</span>
                </div>
                <div class="space-y-3">
                    <article v-for="row in categories" :key="row.id" class="rounded-xl border border-slate-200 p-4">
                        <div class="flex items-start justify-between gap-3">
                            <div>
                                <h4 class="text-sm font-semibold text-slate-900">{{ row.name }}</h4>
                                <div class="mt-1 text-xs text-slate-500">#{{ row.id }} / {{ row.slug }}</div>
                                <p class="mt-2 text-sm text-slate-600">{{ row.description || "无描述" }}</p>
                            </div>
                            <div class="flex gap-2">
                                <button class="rounded bg-amber-100 px-2 py-1 text-xs text-amber-700 hover:bg-amber-200"
                                    :disabled="loading" @click="applyCategoryForm(row)">编辑</button>
                                <button class="rounded bg-rose-100 px-2 py-1 text-xs text-rose-700 hover:bg-rose-200"
                                    :disabled="loading" @click="removeCategory(row.id)">删除</button>
                            </div>
                        </div>
                    </article>
                    <div v-if="categories.length === 0"
                        class="rounded-xl border border-dashed border-slate-200 p-6 text-center text-sm text-slate-500">
                        暂无分类数据</div>
                </div>
            </section>

            <section class="panel p-5">
                <div class="mb-3 flex items-center justify-between">
                    <h3 class="text-sm font-semibold text-slate-700">标签列表</h3>
                    <span class="text-xs text-slate-500">{{ tags.length }} 项</span>
                </div>
                <div class="space-y-3">
                    <article v-for="row in tags" :key="row.id"
                        class="flex items-center justify-between rounded-xl border border-slate-200 p-4">
                        <div>
                            <h4 class="text-sm font-semibold text-slate-900">{{ row.name }}</h4>
                            <div class="mt-1 text-xs text-slate-500">#{{ row.id }}<span
                                    v-if="typeof row.post_count === 'number'"> / {{ row.post_count }} 篇文章</span></div>
                        </div>
                        <button class="rounded bg-rose-100 px-2 py-1 text-xs text-rose-700 hover:bg-rose-200"
                            :disabled="loading" @click="removeTag(row.id)">删除</button>
                    </article>
                    <div v-if="tags.length === 0"
                        class="rounded-xl border border-dashed border-slate-200 p-6 text-center text-sm text-slate-500">
                        暂无标签数据</div>
                </div>
            </section>
        </div>

        <section class="panel p-5">
            <h3 class="text-sm font-semibold text-slate-700">响应结果</h3>
            <pre
                class="mt-3 max-h-[360px] overflow-auto rounded-lg bg-slate-950 p-4 text-xs text-emerald-200">{{ result }}</pre>
        </section>
    </section>
</template>