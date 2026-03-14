import { createRouter, createWebHistory, RouteRecordRaw } from "vue-router";
import { fetchAuthStatus } from "@/services/helmApi";

const routes: RouteRecordRaw[] = [
  {
    path: "/",
    name: "Home",
    component: () => import("@/views/Home.vue"),
    meta: { requiresAuth: true, layout: "shell" },
  },
  {
    path: "/dashboard",
    redirect: "/",
  },
  {
    path: "/posts",
    name: "Posts",
    component: () => import("@/views/Posts.vue"),
    meta: { requiresAuth: true, layout: "shell" },
  },
  {
    path: "/taxonomy",
    name: "Taxonomy",
    component: () => import("@/views/Taxonomy.vue"),
    meta: { requiresAuth: true, layout: "shell" },
  },
  {
    path: "/comments",
    name: "Comments",
    component: () => import("@/views/Comments.vue"),
    meta: { requiresAuth: true, layout: "shell" },
  },
  {
    path: "/lab/api",
    name: "ApiConsole",
    component: () => import("@/views/ApiConsole.vue"),
    meta: { requiresAuth: true, layout: "standalone" },
  },
  {
    path: "/profile",
    name: "Profile",
    component: () => import("@/views/Profile.vue"),
    meta: { requiresAuth: true, layout: "shell" },
  },
  {
    path: "/auth",
    name: "AuthCenter",
    component: () => import("@/views/AuthCenter.vue"),
    meta: { layout: "auth" },
  },
  {
    path: "/auth/login",
    redirect: (to) => ({
      path: "/auth",
      query: {
        ...to.query,
        mode: "login",
      },
    }),
  },
  {
    path: "/auth/register",
    redirect: (to) => ({
      path: "/auth",
      query: {
        ...to.query,
        mode: "register",
      },
    }),
  },
  {
    path: "/:pathMatch(.*)*",
    redirect: "/",
  },
];

export const router = createRouter({
  history: createWebHistory(),
  routes,
});

router.beforeEach(async (to) => {
  if (!to.meta.requiresAuth) {
    return true;
  }

  try {
    const ok = await fetchAuthStatus();
    if (ok) {
      return true;
    }
  } catch {
    // Fall through to login redirect.
  }

  return {
    path: "/auth",
    query: {
      mode: "login",
      redirect: to.fullPath,
    },
  };
});
