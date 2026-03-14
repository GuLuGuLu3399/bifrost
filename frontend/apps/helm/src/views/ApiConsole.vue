<script setup lang="ts">
import { RouterLink } from "vue-router";
import { ref } from "vue";
import {
    batchGetPosts,
    changePassword,
    createAdminPost,
    createCategory,
    createComment,
    deleteAdminPost,
    deleteCategory,
    deleteComment,
    deleteTag,
    fetchAdminPostSource,
    fetchAuthStatus,
    fetchUserProfile,
    getAdminPost,
    getPublicPost,
    getPublicUser,
    getUploadTicket,
    listCategories,
    listComments,
    listDrafts,
    listPosts,
    listTags,
    searchPosts,
    searchSuggest,
    updateAdminPost,
    updateCategory,
    updateUserProfile,
    uploadImageFile,
} from "@/services/helmApi";
import { useSessionState } from "@/composables/useSessionState";

const loading = ref(false);
const result = ref("等待调用");
const { authenticated, refreshSession, logoutSession, lastCheckedAt } = useSessionState();

const userId = ref(2);
const slug = ref("hello-bifrost");
const postId = ref(1);
const commentId = ref(1);
const categoryId = ref(1);
const tagId = ref(1);
const authorId = ref(2);

const searchText = ref("bifrost");
const suggestText = ref("bi");

const oldPassword = ref("old_password");
const newPassword = ref("new_password");

const nickname = ref("helm-admin");
const bio = ref("managed by helm");

const keyword = ref("");
const status = ref<number | undefined>(undefined);

const tagNamesCsv = ref("backend,go");
const markdown = ref("# hello\n\nthis is markdown from helm");
const uploadFilePath = ref("C:/tmp/example.png");

const categoryName = ref("tech");
const categorySlug = ref("tech");
const categoryDesc = ref("tech category");

const commentContent = ref("looks good");

const ticketFilename = ref("avatar.webp");
const ticketUsage = ref<"avatar" | "cover" | "post_image">("avatar");

function splitTags(input: string): string[] {
    return input
        .split(",")
        .map((item) => item.trim())
        .filter((item) => item.length > 0);
}

function stringifyError(error: unknown): string {
    return error instanceof Error ? error.message : String(error);
}

async function runAction(action: () => Promise<unknown>): Promise<void> {
    loading.value = true;
    try {
        const data = await action();
        result.value = JSON.stringify(data ?? { ok: true }, null, 2);
    } catch (error) {
        result.value = stringifyError(error);
    } finally {
        loading.value = false;
    }
}
</script>

<template>
    <section class="mx-auto max-w-7xl space-y-6">
        <header class="rounded-[28px] bg-slate-950 px-6 py-7 text-white shadow-[0_20px_60px_rgba(15,23,42,0.24)]">
            <div class="flex flex-col gap-5 lg:flex-row lg:items-end lg:justify-between">
                <div>
                    <div class="text-xs uppercase tracking-[0.24em] text-slate-400">Helm Lab</div>
                    <h2 class="mt-3 text-3xl font-semibold tracking-tight">接口实验页</h2>
                    <p class="mt-2 max-w-2xl text-sm leading-6 text-slate-300">
                        这里承载联调、回归与快速验证，不再出现在主导航栏。所有请求仍然走 Rust 和 Tauri 通道。
                    </p>
                </div>
                <div class="flex flex-wrap gap-2 text-sm">
                    <RouterLink class="rounded-full border border-white/15 px-4 py-2 text-slate-100 hover:bg-white/10"
                        to="/">返回概览</RouterLink>
                    <RouterLink class="rounded-full border border-white/15 px-4 py-2 text-slate-100 hover:bg-white/10"
                        to="/posts">文章管理</RouterLink>
                    <RouterLink class="rounded-full border border-white/15 px-4 py-2 text-slate-100 hover:bg-white/10"
                        to="/auth">认证中心</RouterLink>
                </div>
            </div>
        </header>

        <section class="grid gap-4 xl:grid-cols-[1.15fr_0.85fr]">
            <div class="grid gap-4">
                <article class="rounded-[24px] border border-slate-200 bg-white p-5 shadow-sm">
                    <div class="flex items-center justify-between gap-3">
                        <div>
                            <h3 class="text-sm font-semibold text-slate-800">会话概况</h3>
                            <p class="mt-1 text-sm text-slate-500">当前状态：{{ authenticated ? "已登录" : "未登录" }}</p>
                        </div>
                        <div class="text-xs text-slate-400">
                            最近检查：{{ lastCheckedAt ? new Date(lastCheckedAt).toLocaleTimeString() : "等待首次检查" }}
                        </div>
                    </div>
                    <div class="mt-4 flex flex-wrap gap-2">
                        <button
                            class="rounded-xl bg-slate-900 px-3 py-2 text-sm font-medium text-white hover:bg-slate-800 disabled:opacity-50"
                            :disabled="loading" @click="refreshSession">刷新会话</button>
                        <button
                            class="rounded-xl bg-slate-100 px-3 py-2 text-sm font-medium text-slate-700 hover:bg-slate-200 disabled:opacity-50"
                            :disabled="loading" @click="logoutSession">退出登录</button>
                        <button
                            class="rounded-xl bg-steel-100 px-3 py-2 text-sm font-medium text-steel-800 hover:bg-steel-200 disabled:opacity-50"
                            :disabled="loading" @click="runAction(() => fetchAuthStatus())">读取认证状态</button>
                    </div>
                </article>

                <article class="rounded-[24px] border border-slate-200 bg-white p-5 shadow-sm">
                    <h3 class="text-sm font-semibold text-slate-800">基础参数</h3>
                    <div class="mt-4 grid gap-3 md:grid-cols-2 xl:grid-cols-3">
                        <label class="field">
                            <span>user_id</span>
                            <input v-model.number="userId" type="number" min="1" />
                        </label>
                        <label class="field">
                            <span>slug</span>
                            <input v-model="slug" type="text" />
                        </label>
                        <label class="field">
                            <span>post_id</span>
                            <input v-model.number="postId" type="number" min="1" />
                        </label>
                        <label class="field">
                            <span>comment_id</span>
                            <input v-model.number="commentId" type="number" min="1" />
                        </label>
                        <label class="field">
                            <span>category_id</span>
                            <input v-model.number="categoryId" type="number" min="1" />
                        </label>
                        <label class="field">
                            <span>tag_id</span>
                            <input v-model.number="tagId" type="number" min="1" />
                        </label>
                        <label class="field">
                            <span>author_id</span>
                            <input v-model.number="authorId" type="number" min="1" />
                        </label>
                        <label class="field">
                            <span>search q</span>
                            <input v-model="searchText" type="text" />
                        </label>
                        <label class="field">
                            <span>suggest prefix</span>
                            <input v-model="suggestText" type="text" />
                        </label>
                        <label class="field">
                            <span>drafts keyword</span>
                            <input v-model="keyword" type="text" />
                        </label>
                        <label class="field">
                            <span>drafts status</span>
                            <input v-model.number="status" type="number" min="0" max="3" />
                        </label>
                        <label class="field">
                            <span>tag_names csv</span>
                            <input v-model="tagNamesCsv" type="text" />
                        </label>
                        <label class="field md:col-span-2 xl:col-span-2">
                            <span>upload file path</span>
                            <input v-model="uploadFilePath" type="text" />
                        </label>
                        <label class="field">
                            <span>ticket filename</span>
                            <input v-model="ticketFilename" type="text" />
                        </label>
                        <label class="field">
                            <span>ticket usage</span>
                            <select v-model="ticketUsage">
                                <option value="avatar">avatar</option>
                                <option value="cover">cover</option>
                                <option value="post_image">post_image</option>
                            </select>
                        </label>
                    </div>
                </article>

                <article class="lab-card">
                    <div class="lab-head">
                        <div>
                            <h3>Auth / User</h3>
                            <p>会话检查、资料拉取与密码、档案更新。</p>
                        </div>
                    </div>
                    <div class="actions">
                        <button :disabled="loading" @click="runAction(() => fetchUserProfile())">GET
                            /v1/users/profile</button>
                        <button :disabled="loading"
                            @click="runAction(() => changePassword({ oldPassword, newPassword }))">POST
                            /v1/users/password</button>
                        <button :disabled="loading" @click="runAction(() => updateUserProfile({ nickname, bio }))">PUT
                            /v1/users/profile</button>
                        <button :disabled="loading" @click="runAction(() => getPublicUser(userId))">GET
                            /v1/users/{id}</button>
                    </div>
                    <div class="inline-grid">
                        <input v-model="oldPassword" type="password" placeholder="old password" />
                        <input v-model="newPassword" type="password" placeholder="new password" />
                        <input v-model="nickname" type="text" placeholder="nickname" />
                        <input v-model="bio" type="text" placeholder="bio" />
                    </div>
                </article>

                <article class="lab-card">
                    <div class="lab-head">
                        <div>
                            <h3>Public Post / Search</h3>
                            <p>公开内容查询、评论读取与搜索验证。</p>
                        </div>
                    </div>
                    <div class="actions">
                        <button :disabled="loading"
                            @click="runAction(() => listPosts({ pageSize: 20, categoryId, tagId, authorId }))">GET
                            /v1/posts</button>
                        <button :disabled="loading" @click="runAction(() => getPublicPost(slug))">GET
                            /v1/posts/{slug}</button>
                        <button :disabled="loading" @click="runAction(() => batchGetPosts([postId]))">POST
                            /v1/posts:batch</button>
                        <button :disabled="loading" @click="runAction(() => listComments(postId, { pageSize: 20 }))">GET
                            /v1/posts/{id}/comments</button>
                        <button :disabled="loading"
                            @click="runAction(() => searchPosts({ q: searchText, page: 1, pageSize: 20 }))">GET
                            /v1/search</button>
                        <button :disabled="loading"
                            @click="runAction(() => searchSuggest({ prefix: suggestText, limit: 5 }))">GET
                            /v1/search/suggest</button>
                    </div>
                </article>

                <article class="lab-card">
                    <div class="lab-head">
                        <div>
                            <h3>Admin Post</h3>
                            <p>草稿、创建、详情、更新、删除与源码查看。</p>
                        </div>
                    </div>
                    <div class="actions">
                        <button :disabled="loading"
                            @click="runAction(() => listDrafts({ page: 1, pageSize: 20, keyword, status }))">GET
                            /v1/drafts</button>
                        <button :disabled="loading"
                            @click="runAction(() => createAdminPost({ title: `helm-${Date.now()}`, slug: `helm-${Date.now()}`, rawMarkdown: markdown, categoryId, tagNames: splitTags(tagNamesCsv), status: 1 }))">POST
                            /v1/admin/posts</button>
                        <button :disabled="loading" @click="runAction(() => getAdminPost(postId))">GET
                            /v1/admin/posts/{id}</button>
                        <button :disabled="loading"
                            @click="runAction(() => updateAdminPost(postId, { postId, rawMarkdown: markdown, categoryId, tagNames: splitTags(tagNamesCsv) }))">PUT
                            /v1/admin/posts/{id}</button>
                        <button :disabled="loading" @click="runAction(() => deleteAdminPost(postId))">DELETE
                            /v1/admin/posts/{id}</button>
                        <button :disabled="loading" @click="runAction(() => fetchAdminPostSource(postId))">GET
                            /v1/admin/posts/{id}/source</button>
                    </div>
                    <textarea v-model="markdown" rows="6" placeholder="markdown content" />
                </article>

                <article class="lab-card">
                    <div class="lab-head">
                        <div>
                            <h3>Admin Meta / Storage</h3>
                            <p>分类、标签、评论与上传票据的管理接口。</p>
                        </div>
                    </div>
                    <div class="actions">
                        <button :disabled="loading" @click="runAction(() => listCategories())">GET
                            /v1/categories</button>
                        <button :disabled="loading" @click="runAction(() => listTags())">GET /v1/tags</button>
                        <button :disabled="loading"
                            @click="runAction(() => createCategory({ name: categoryName, slug: categorySlug, description: categoryDesc }))">POST
                            /v1/categories</button>
                        <button :disabled="loading"
                            @click="runAction(() => updateCategory(categoryId, { categoryId, name: categoryName, slug: categorySlug, description: categoryDesc }))">PUT
                            /v1/categories/{id}</button>
                        <button :disabled="loading" @click="runAction(() => deleteCategory(categoryId))">DELETE
                            /v1/categories/{id}</button>
                        <button :disabled="loading" @click="runAction(() => deleteTag(tagId))">DELETE
                            /v1/tags/{id}</button>
                        <button :disabled="loading"
                            @click="runAction(() => createComment(postId, { content: commentContent }))">POST
                            /v1/posts/{id}/comments</button>
                        <button :disabled="loading" @click="runAction(() => deleteComment(commentId))">DELETE
                            /v1/comments/{id}</button>
                        <button :disabled="loading"
                            @click="runAction(() => getUploadTicket({ filename: ticketFilename, usage: ticketUsage }))">POST
                            /v1/storage/upload_ticket</button>
                        <button :disabled="loading"
                            @click="runAction(() => uploadImageFile(uploadFilePath))">upload_image_cmd(filePath)</button>
                    </div>
                    <div class="inline-grid">
                        <input v-model="categoryName" type="text" placeholder="category name" />
                        <input v-model="categorySlug" type="text" placeholder="category slug" />
                        <input v-model="categoryDesc" type="text" placeholder="category desc" />
                        <input v-model="commentContent" type="text" placeholder="comment content" />
                    </div>
                </article>
            </div>

            <aside class="sticky top-6 h-fit rounded-[24px] border border-slate-200 bg-white p-5 shadow-sm">
                <div class="flex items-center justify-between gap-3">
                    <div>
                        <h3 class="text-sm font-semibold text-slate-800">响应面板</h3>
                        <p class="mt-1 text-sm text-slate-500">所有结果和错误统一输出在这里。</p>
                    </div>
                    <span class="rounded-full bg-slate-100 px-3 py-1 text-xs text-slate-500">{{ loading ? "请求中" : "空闲"
                    }}</span>
                </div>
                <pre
                    class="mt-4 min-h-[720px] overflow-auto rounded-2xl bg-slate-950 p-4 text-xs leading-6 text-emerald-200">{{ result }}</pre>
            </aside>
        </section>
    </section>
</template>

<style scoped>
.field {
    display: grid;
    gap: 6px;
    font-size: 12px;
    font-weight: 600;
    color: rgb(71 85 105);
}

.lab-card {
    border: 1px solid rgb(226 232 240);
    border-radius: 24px;
    background: white;
    padding: 20px;
    box-shadow: 0 10px 30px rgba(15, 23, 42, 0.04);
}

.lab-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    margin-bottom: 14px;
}

.lab-head h3 {
    margin: 0;
    font-size: 14px;
    font-weight: 700;
    color: rgb(15 23 42);
}

.lab-head p {
    margin: 4px 0 0;
    font-size: 13px;
    color: rgb(100 116 139);
}

.actions {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-bottom: 14px;
}

.inline-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 10px;
}

label,
input,
select,
textarea,
button {
    font: inherit;
}

input,
select,
textarea {
    width: 100%;
    padding: 10px 12px;
    border: 1px solid rgb(203 213 225);
    border-radius: 14px;
    background: rgb(248 250 252);
}

button {
    padding: 10px 12px;
    border-radius: 14px;
    background: rgb(241 245 249);
    color: rgb(51 65 85);
}

pre {
    white-space: pre-wrap;
    word-break: break-word;
    margin: 0;
}

@media (max-width: 960px) {
    .inline-grid {
        grid-template-columns: 1fr;
    }
}
</style>
