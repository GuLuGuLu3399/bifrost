<script setup lang="ts">
import { ref } from "vue";
import { useRouter } from "vue-router";
import { invoke } from "@tauri-apps/api/core";
import {
  Button,
  Input,
  Card,
  CardHeader,
  CardTitle,
  CardContent,
  CardFooter,
} from "@bifrost/ui";

const router = useRouter();
const loading = ref(false);
const errorMsg = ref("");

const form = ref({
  username: "",
  email: "",
  password: "",
  confirmPassword: "",
});

const handleRegister = async () => {
  errorMsg.value = "";
  if (form.value.password !== form.value.confirmPassword) {
    errorMsg.value = "ERROR: Passwords do not match";
    return;
  }
  if (!form.value.username || !form.value.email || !form.value.password) {
    errorMsg.value = "ERROR: Missing required fields";
    return;
  }
  try {
    loading.value = true;
    await invoke("register_cmd", {
      dto: {
        username: form.value.username,
        email: form.value.email,
        password: form.value.password,
      },
    });
    alert("User Created. Redirecting to Login...");
    router.push("/auth/login");
  } catch (e: any) {
    errorMsg.value = `SYSTEM_HALT: ${e}`;
  } finally {
    loading.value = false;
  }
};
</script>

<template>
  <div class="flex h-full w-full items-center justify-center p-4">
    <Card
      class="w-[400px] border-primary/50 shadow-none bg-background/95 backdrop-blur"
    >
      <CardHeader
        class="pb-2 text-center border-b"
        style="border-color: rgba(255, 255, 255, 0.2)"
      >
        <CardTitle class="text-xl font-mono tracking-tighter text-primary">
          > NEW_USER_REGISTRATION_
        </CardTitle>
      </CardHeader>
      <CardContent class="pt-6 space-y-4">
        <div class="space-y-2">
          <label
            class="text-xs font-mono text-muted-foreground"
            style="margin-left: 4px"
            >IDENTITY (USERNAME)</label
          >
          <Input
            v-model="form.username"
            placeholder="operator_01"
            class="font-mono"
          />
        </div>
        <div class="space-y-2">
          <label
            class="text-xs font-mono text-muted-foreground"
            style="margin-left: 4px"
            >COMM_LINK (EMAIL)</label
          >
          <Input
            v-model="form.email"
            type="email"
            placeholder="op@bifrost.com"
            class="font-mono"
          />
        </div>
        <div class="grid md-cols-2" style="gap: 16px">
          <div class="space-y-2">
            <label
              class="text-xs font-mono text-muted-foreground"
              style="margin-left: 4px"
              >SECRET</label
            >
            <Input v-model="form.password" type="password" class="font-mono" />
          </div>
          <div class="space-y-2">
            <label
              class="text-xs font-mono text-muted-foreground"
              style="margin-left: 4px"
              >CONFIRM</label
            >
            <Input
              v-model="form.confirmPassword"
              type="password"
              class="font-mono"
            />
          </div>
        </div>
        <div
          v-if="errorMsg"
          class="p-2 border bg-destructive/10 text-destructive text-xs font-mono"
          style="border-color: rgba(255, 0, 0, 0.4)"
        >
          {{ errorMsg }}
        </div>
      </CardContent>
      <CardFooter
        class="flex justify-between border-t"
        style="border-color: rgba(255, 255, 255, 0.2); padding-top: 16px"
      >
        <Button
          variant="ghost"
          @click="router.push('/auth/login')"
          class="text-xs font-mono text-muted-foreground"
        >
          <span>&lt; BACK_TO_LOGIN</span>
        </Button>
        <Button
          :disabled="loading"
          @click="handleRegister"
          class="font-mono font-bold tracking-wide"
        >
          {{ loading ? "PROCESSING..." : "INIT_SEQUENCE" }}
        </Button>
      </CardFooter>
    </Card>
  </div>
</template>
