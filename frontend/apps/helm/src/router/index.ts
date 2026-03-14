import { createRouter, createWebHistory, RouteRecordRaw } from "vue-router";

const routes: RouteRecordRaw[] = [
  {
    path: "/",
    name: "Dashboard",
    component: () => import("@/views/Dashboard.vue"),
  },
  {
    path: "/auth/login",
    name: "Login",
    component: () => import("@/views/auth/Login.vue"),
  },
  {
    path: "/auth/register",
    name: "Register",
    component: () => import("@/views/auth/Register.vue"),
  },
];

export const router = createRouter({
  history: createWebHistory(),
  routes,
});
