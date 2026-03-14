<script setup lang="ts">
import { ref } from "vue";
import { createComment, deleteComment, listComments } from "@/services/helmApi";

interface CommentRow {
    id: number;
    parent_id?: number;
    content?: string;
    created_at?: number;
    status?: number;
    author?: { nickname?: string };
}

const commentStatusMap: Record<number, string> = {
    1: "待审核",
    2: "已通过",
    3: "垃圾评论",
};

const postId = ref(1);
const pageSize = ref(20);
const pageToken = ref("");
const loading = ref(false);
const result = ref("等待查询");
const comments = ref<CommentRow[]>([]);
const commentId = ref<number | null>(null);
const commentContent = ref("");
const nextPageToken = ref("");

function asRecord(data: unknown): Record<string, unknown> {
    return data && typeof data === "object" ? (data as Record<string, unknown>) : {};
}

function formatTime(ts?: number): string {
    if (!ts || ts <= 0) {
        return "-";
    }
    const ms = ts > 1_000_000_000_000 ? ts : ts * 1000;
    return new Date(ms).toLocaleString();
}

function getCommentStatusLabel(status?: number): string {
    if (!status) {
        return "未知状态";
    }

    return commentStatusMap[status] ?? `状态 ${status}`;
}

function normalizeComments(data: unknown): void {
    const obj = asRecord(data);
    const items = obj.comments;
    comments.value = Array.isArray(items) ? (items as CommentRow[]) : [];
    const page = asRecord(obj.page);
    nextPageToken.value =
        typeof page.next_page_token === "string"
            ? page.next_page_token
            : typeof page.nextPageToken === "string"
                ? page.nextPageToken
                : "";
}

async function loadComments(): Promise<void> {
    loading.value = true;
    try {
        const data = await listComments(postId.value, {
            pageSize: pageSize.value,
            pageToken: pageToken.value || undefined,
        });
        normalizeComments(data);
        result.value = JSON.stringify(data, null, 2);
    } catch (error) {
        comments.value = [];
        nextPageToken.value = "";
        result.value = `读取评论失败: ${String(error)}`;
    } finally {
        loading.value = false;
    }
}

async function loadNextPage(): Promise<void> {
    if (!nextPageToken.value) {
        return;
    }
    pageToken.value = nextPageToken.value;
    await loadComments();
}

async function submitComment(): Promise<void> {
    loading.value = true;
    try {
        const data = await createComment(postId.value, {
            content: commentContent.value,
            parentId: undefined,
        });
        result.value = JSON.stringify(data, null, 2);
        commentContent.value = "";
        pageToken.value = "";
        await loadComments();
    } catch (error) {
        result.value = `创建评论失败: ${String(error)}`;
    } finally {
        loading.value = false;
    }
}

async function removeComment(id: number): Promise<void> {
    if (!window.confirm(`确认删除评论 #${id} ?`)) {
        return;
    }
    loading.value = true;
    try {
        const data = await deleteComment(id);
        result.value = JSON.stringify(data, null, 2);
        comments.value = comments.value.filter((row) => row.id !== id);
    } catch (error) {
        result.value = `删除评论失败: ${String(error)}`;
    } finally {
        loading.value = false;
    }
}
</script>

<template>
    <section class="mx-auto max-w-6xl space-y-4">
        <header class="panel p-5">
            <h2 class="text-xl font-semibold text-slate-900">评论管理</h2>
            <p class="mt-2 text-sm text-slate-600">按文章读取评论，支持新增测试评论和删除已有评论。</p>
        </header>

        <section class="panel p-5">
            <div class="grid gap-3 md:grid-cols-3">
                <label>
                    <span class="mb-1 block text-xs font-medium text-slate-600">post_id</span>
                    <input v-model.number="postId" class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm"
                        type="number" min="1" />
                </label>
                <label>
                    <span class="mb-1 block text-xs font-medium text-slate-600">page_size</span>
                    <input v-model.number="pageSize" class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm"
                        type="number" min="1" max="100" />
                </label>
                <label>
                    <span class="mb-1 block text-xs font-medium text-slate-600">page_token</span>
                    <input v-model="pageToken" class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm"
                        type="text" />
                </label>
            </div>
            <div class="mt-4 flex gap-2">
                <button
                    class="rounded-lg bg-steel-600 px-4 py-2 text-sm font-medium text-white hover:bg-steel-700 disabled:opacity-50"
                    :disabled="loading" @click="loadComments">读取评论</button>
                <button
                    class="rounded-lg bg-slate-100 px-4 py-2 text-sm font-medium text-slate-700 hover:bg-slate-200 disabled:opacity-50"
                    :disabled="loading || !nextPageToken" @click="loadNextPage">下一页</button>
            </div>
        </section>

        <section class="panel p-5">
            <div class="mb-3 flex items-center justify-between">
                <h3 class="text-sm font-semibold text-slate-700">新增评论</h3>
                <span v-if="commentId" class="text-xs text-slate-500">最近选择评论: #{{ commentId }}</span>
            </div>
            <div class="grid gap-3 md:grid-cols-[1fr_auto]">
                <input v-model="commentContent" class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm"
                    type="text" placeholder="输入评论内容" />
                <button
                    class="rounded-lg bg-slate-900 px-4 py-2 text-sm font-medium text-white hover:bg-slate-800 disabled:opacity-50"
                    :disabled="loading || !commentContent.trim()" @click="submitComment">创建评论</button>
            </div>
        </section>

        <section class="panel p-5">
            <div class="mb-3 flex items-center justify-between">
                <h3 class="text-sm font-semibold text-slate-700">评论列表</h3>
                <span class="text-xs text-slate-500">{{ comments.length }} 条</span>
            </div>
            <div class="space-y-3">
                <article v-for="row in comments" :key="row.id" class="rounded-xl border border-slate-200 p-4">
                    <div class="flex items-start justify-between gap-3">
                        <div>
                            <div class="text-sm font-semibold text-slate-900">#{{ row.id }} {{ row.author?.nickname ||
                                "匿名用户" }}</div>
                            <div class="mt-1 text-xs text-slate-500">parent: {{ row.parent_id || 0 }} / {{
                                formatTime(row.created_at) }}</div>
                            <div class="mt-1 text-xs text-slate-500">{{ getCommentStatusLabel(row.status) }}</div>
                            <p class="mt-2 text-sm text-slate-700">{{ row.content || "无内容" }}</p>
                        </div>
                        <button
                            class="rounded bg-rose-100 px-2 py-1 text-xs text-rose-700 hover:bg-rose-200 disabled:opacity-50"
                            :disabled="loading" @click="commentId = row.id; removeComment(row.id)">删除</button>
                    </div>
                </article>
                <div v-if="comments.length === 0"
                    class="rounded-xl border border-dashed border-slate-200 p-6 text-center text-sm text-slate-500">
                    暂无评论数据，先按文章读取。</div>
            </div>
        </section>

        <section class="panel p-5">
            <h3 class="text-sm font-semibold text-slate-700">响应结果</h3>
            <pre
                class="mt-3 max-h-[360px] overflow-auto rounded-lg bg-slate-950 p-4 text-xs text-emerald-200">{{ result }}</pre>
        </section>
    </section>
</template>