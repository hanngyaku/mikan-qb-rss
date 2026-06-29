<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api } from '../api'

const count = ref(100)
const lines = ref<string[]>([])
const message = ref('')

async function load() {
  try {
    lines.value = (await api.getLogs(count.value)).lines ?? []
    message.value = ''
  } catch (error) { message.value = String(error) }
}

onMounted(load)
</script>

<template>
  <section>
    <div class="title">
      <h1>日志</h1>
      <form class="log-controls" @submit.prevent="load">
        <label>最新行数<input v-model.number="count" type="number" min="1" max="1000"></label>
        <button type="submit">刷新</button>
      </form>
    </div>
    <p v-if="message" class="message">{{ message }}</p>
    <pre class="logs">{{ lines.join('\n') }}</pre>
  </section>
</template>
