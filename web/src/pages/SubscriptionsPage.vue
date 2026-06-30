<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { api, type CreateSubscription, type Subscription } from '../api'
import SubscriptionCard from '../components/SubscriptionCard.vue'

const items = ref<Subscription[]>([])
const showAdd = ref(false)
const message = ref('')
const form = reactive<CreateSubscription>({ rssUrl: '', regex: '', excludeRegex: '', customDirName: '', season: 1 })

async function load() {
  try { items.value = await api.listSubscriptions() }
  catch (error) { message.value = String(error) }
}

async function add() {
  try {
    const excludeRegex = form.excludeRegex
    await api.createSubscription(form)
    Object.assign(form, { rssUrl: '', regex: '', excludeRegex, customDirName: '', season: 1 })
    showAdd.value = false
    await load()
  } catch (error) { message.value = String(error) }
}

onMounted(async () => {
  await load()
  try {
    const settings = await api.getSettings()
    form.excludeRegex = settings.latestExcludeRegex || settings.defaultExcludeRegex || ''
  } catch (error) { message.value = String(error) }
})
</script>

<template>
  <section>
    <div class="title"><h1>订阅</h1><button @click="showAdd = !showAdd">添加 RSS</button></div>
    <form v-if="showAdd" class="panel" @submit.prevent="add">
      <label>RSS URL<input v-model.trim="form.rssUrl" type="url" required></label>
      <label>正则表达式（可选）<input v-model="form.regex"></label>
      <label>排除正则（可选）<input v-model="form.excludeRegex" placeholder="例如：720|\d+-\d+"></label>
      <label>自定义目录名（可选）<input v-model="form.customDirName"></label>
      <label>Season<input v-model.number="form.season" type="number" min="1" required></label>
      <button type="submit">保存</button>
    </form>
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
