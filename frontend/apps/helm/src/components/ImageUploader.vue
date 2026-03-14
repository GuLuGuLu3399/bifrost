<script setup lang="ts">
import { ref } from "vue";
import { invoke } from "@tauri-apps/api/core";
import { open } from "@tauri-apps/plugin-dialog";

interface LoginResponse {
  token: string;
  user_id?: string;
  nickname?: string;
}

interface UploadResponse {
  key: string;
  url?: string;
}

const username = ref("");
const password = ref("");
const isLoggedIn = ref(false);
const isLoading = ref(false);
const error = ref("");
const selectedFile = ref("");
const uploadResult = ref<UploadResponse | null>(null);

// Login function
async function handleLogin() {
  if (!username.value || !password.value) {
    error.value = "Please enter username and password";
    return;
  }

  isLoading.value = true;
  error.value = "";

  try {
    const response = await invoke<LoginResponse>("login_cmd", {
      username: username.value,
      password: password.value,
    });

    console.log("Login successful:", response);
    isLoggedIn.value = true;
    error.value = "";
  } catch (e) {
    error.value = `Login failed: ${e}`;
    console.error("Login error:", e);
  } finally {
    isLoading.value = false;
  }
}

// Select file function
async function selectImage() {
  try {
    const selected = await open({
      multiple: false,
      filters: [
        {
          name: "Image",
          extensions: ["png", "jpg", "jpeg", "webp", "gif"],
        },
      ],
    });

    if (selected && typeof selected === "string") {
      selectedFile.value = selected;
      error.value = "";
    }
  } catch (e) {
    error.value = `Failed to select file: ${e}`;
  }
}

// Upload image function
async function handleUpload() {
  if (!selectedFile.value) {
    error.value = "Please select an image first";
    return;
  }

  if (!isLoggedIn.value) {
    error.value = "Please login first";
    return;
  }

  isLoading.value = true;
  error.value = "";
  uploadResult.value = null;

  try {
    const response = await invoke<UploadResponse>("upload_image_cmd", {
      filePath: selectedFile.value,
    });

    console.log("Upload successful:", response);
    uploadResult.value = response;
    selectedFile.value = "";
  } catch (e) {
    error.value = `Upload failed: ${e}`;
    console.error("Upload error:", e);
  } finally {
    isLoading.value = false;
  }
}

// Check auth status on mount
async function checkAuth() {
  try {
    isLoggedIn.value = await invoke<boolean>("is_authenticated");
  } catch (e) {
    console.error("Check auth error:", e);
  }
}

// Logout function
async function handleLogout() {
  try {
    await invoke("logout_cmd");
    isLoggedIn.value = false;
    username.value = "";
    password.value = "";
  } catch (e) {
    error.value = `Logout failed: ${e}`;
  }
}

// Check auth on component mount
checkAuth();
</script>

<template>
  <div class="p-8 max-w-2xl mx-auto">
    <h1 class="text-3xl font-bold mb-8">Helm Admin Dashboard</h1>

    <!-- Error Display -->
    <div
      v-if="error"
      class="mb-4 p-4 bg-red-50 border border-red-200 rounded-lg text-red-700"
    >
      {{ error }}
    </div>

    <!-- Success Display -->
    <div
      v-if="uploadResult"
      class="mb-4 p-4 bg-green-50 border border-green-200 rounded-lg text-green-700"
    >
      <p class="font-semibold">Upload Successful!</p>
      <p class="text-sm mt-2">Key: {{ uploadResult.key }}</p>
      <p v-if="uploadResult.url" class="text-sm">URL: {{ uploadResult.url }}</p>
    </div>

    <!-- Login Section -->
    <div v-if="!isLoggedIn" class="mb-8 p-6 bg-gray-50 rounded-lg">
      <h2 class="text-xl font-semibold mb-4">Login</h2>
      <div class="space-y-4">
        <div>
          <label class="block text-sm font-medium mb-1">Username</label>
          <input
            v-model="username"
            type="text"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
            :disabled="isLoading"
          />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">Password</label>
          <input
            v-model="password"
            type="password"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
            :disabled="isLoading"
            @keyup.enter="handleLogin"
          />
        </div>
        <button
          @click="handleLogin"
          :disabled="isLoading"
          class="w-full px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:bg-gray-400 transition-colors"
        >
          {{ isLoading ? "Logging in..." : "Login" }}
        </button>
      </div>
    </div>

    <!-- Upload Section -->
    <div v-else class="space-y-6">
      <div class="flex items-center justify-between">
        <h2 class="text-xl font-semibold">Upload Image</h2>
        <button
          @click="handleLogout"
          class="px-4 py-2 text-sm text-red-600 hover:bg-red-50 rounded-lg transition-colors"
        >
          Logout
        </button>
      </div>

      <div class="p-6 bg-gray-50 rounded-lg space-y-4">
        <div>
          <button
            @click="selectImage"
            :disabled="isLoading"
            class="px-4 py-2 bg-gray-600 text-white rounded-lg hover:bg-gray-700 disabled:bg-gray-400 transition-colors"
          >
            Select Image
          </button>
          <p v-if="selectedFile" class="text-sm text-gray-600 mt-2">
            Selected: {{ selectedFile.split(/[\\/]/).pop() }}
          </p>
        </div>

        <button
          @click="handleUpload"
          :disabled="isLoading || !selectedFile"
          class="w-full px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:bg-gray-400 transition-colors"
        >
          {{ isLoading ? "Processing & Uploading..." : "Upload (WebP)" }}
        </button>

        <div class="text-sm text-gray-600 space-y-1">
          <p>• Images will be automatically resized to max 1920px</p>
          <p>• Converted to WebP format with 75% quality</p>
          <p>• Uploaded to Gjallar storage service</p>
        </div>
      </div>
    </div>
  </div>
</template>
