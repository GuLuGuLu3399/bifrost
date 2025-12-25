<template>
  <main class="shell">
    <section class="card">
      <p class="eyebrow">Bifrost Helm</p>
      <h1>Desktop shell is ready</h1>
      <p class="muted">Connect shared packages and desktop features here.</p>
    </section>
  </main>
</template>

<style scoped>
.shell {
  min-height: 100vh;
  display: grid;
  place-items: center;
  background: radial-gradient(circle at 20% 20%, #1e293b, #0b1224 55%);
  color: #e2e8f0;
  font-family: "Inter", system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI",
    sans-serif;
}

.card {
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 16px;
  padding: 28px 32px;
  box-shadow: 0 25px 60px rgba(0, 0, 0, 0.35);
  max-width: 480px;
  text-align: center;
}

.eyebrow {
  letter-spacing: 0.12em;
  text-transform: uppercase;
  font-size: 12px;
  color: #94a3b8;
  margin-bottom: 8px;
}

h1 {
  font-size: 26px;
  margin: 0 0 8px 0;
}

.muted {
  color: #cbd5e1;
  margin: 0;
}
</style>
