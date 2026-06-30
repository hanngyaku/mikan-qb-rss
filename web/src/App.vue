<script setup lang="ts">
import { ref } from 'vue'
import SubscriptionEditorModal from './components/SubscriptionEditorModal.vue'

const showAdd = ref(false)

function saved() {
  showAdd.value = false
  window.dispatchEvent(new Event('subscriptions-changed'))
}
</script>

<template>
  <header>
    <strong>Mikan qBittorrent RSS</strong>
    <nav>
      <button class="nav-add" title="添加 RSS" @click="showAdd = true">+</button>
      <RouterLink to="/">放送</RouterLink>
      <RouterLink to="/subscriptions">订阅</RouterLink>
      <RouterLink to="/settings">设置</RouterLink>
      <RouterLink to="/logs">日志</RouterLink>
    </nav>
  </header>
  <main><RouterView /></main>
  <SubscriptionEditorModal v-if="showAdd" @close="showAdd = false" @saved="saved" />
</template>
