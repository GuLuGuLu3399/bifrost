import { computed, onBeforeUnmount, onMounted, ref } from "vue";
import { fetchAuthStatus, logoutUser } from "@/services/helmApi";

const authenticated = ref(false);
const loading = ref(false);
const initialized = ref(false);
const lastCheckedAt = ref<number | null>(null);

let refreshTimer: number | null = null;
let subscribers = 0;

function setAuthenticated(nextValue: boolean): void {
  authenticated.value = nextValue;
  initialized.value = true;
  lastCheckedAt.value = Date.now();
}

async function refreshSession(): Promise<boolean> {
  loading.value = true;
  try {
    const nextValue = await fetchAuthStatus();
    setAuthenticated(nextValue);
    return nextValue;
  } catch {
    setAuthenticated(false);
    return false;
  } finally {
    loading.value = false;
  }
}

async function logoutSession(): Promise<void> {
  loading.value = true;
  try {
    await logoutUser();
  } finally {
    loading.value = false;
    setAuthenticated(false);
  }
}

function markAuthenticated(nextValue: boolean): void {
  setAuthenticated(nextValue);
}

function handleWindowFocus(): void {
  void refreshSession();
}

function startAutoRefresh(): void {
  if (refreshTimer !== null) {
    return;
  }

  refreshTimer = window.setInterval(() => {
    void refreshSession();
  }, 30000);

  window.addEventListener("focus", handleWindowFocus);
  document.addEventListener("visibilitychange", handleWindowFocus);
}

function stopAutoRefresh(): void {
  if (refreshTimer !== null) {
    window.clearInterval(refreshTimer);
    refreshTimer = null;
  }

  window.removeEventListener("focus", handleWindowFocus);
  document.removeEventListener("visibilitychange", handleWindowFocus);
}

export function useSessionState() {
  onMounted(() => {
    subscribers += 1;

    if (subscribers === 1) {
      startAutoRefresh();
    }

    if (!initialized.value) {
      void refreshSession();
    }
  });

  onBeforeUnmount(() => {
    subscribers = Math.max(0, subscribers - 1);
    if (subscribers === 0) {
      stopAutoRefresh();
    }
  });

  return {
    authenticated: computed(() => authenticated.value),
    loading: computed(() => loading.value),
    initialized: computed(() => initialized.value),
    lastCheckedAt: computed(() => lastCheckedAt.value),
    refreshSession,
    logoutSession,
    markAuthenticated,
  };
}
