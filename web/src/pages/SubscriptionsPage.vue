<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { api, type Subscription } from '../api'
import SubscriptionCard from '../components/SubscriptionCard.vue'

const items = ref<Subscription[]>([])
const message = ref('')

async function load() {
  try { items.value = await api.listSubscriptions() }
  catch (error) { message.value = String(error) }
}

onMounted(() => {
  load()
  window.addEventListener('subscriptions-changed', load)
})
onBeforeUnmount(() => window.removeEventListener('subscriptions-changed', load))
</script>

<template>
  <section>
    <h1>订阅</h1>
    <p v-if="message" class="message">{{ message }}</p>
    <p v-if="!items.length">暂无订阅</p>
    <SubscriptionCard
      v-for="item in items"
      :key="item.id"
      :item="item"
      @changed="load"
      @deleted="load"
      @message="message = $event"
    />
  </section>
</template>
