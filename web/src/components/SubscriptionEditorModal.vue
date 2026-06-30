<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { api, type Subscription } from '../api'

const props = defineProps<{ item?: Subscription }>()
const emit = defineEmits<{ close: []; saved: [item: Subscription] }>()
const message = ref('')
const form = reactive({
  rssUrl: props.item?.rssUrl ?? '',
  regex: props.item?.regex ?? '',
  excludeRegex: props.item?.excludeRegex ?? '',
  directoryName: props.item?.saveDirName ?? '',
  season: props.item?.season ?? 1,
  enabled: props.item?.enabled ?? true,
})

onMounted(async () => {
  if (props.item) return
  try {
    const settings = await api.getSettings()
    form.excludeRegex = settings.latestExcludeRegex || settings.defaultExcludeRegex || ''
  } catch (error) { message.value = String(error) }
})

async function save() {
  try {
    const item = props.item?.id
      ? await api.updateSubscription(props.item.id, {
          rssUrl: form.rssUrl,
          regex: form.regex,
          excludeRegex: form.excludeRegex,
          saveDirName: form.directoryName,
          season: form.season,
          enabled: form.enabled,
        })
      : await api.createSubscription({
          rssUrl: form.rssUrl,
          regex: form.regex,
          excludeRegex: form.excludeRegex,
          customDirName: form.directoryName,
          season: form.season,
        })
    emit('saved', item)
  } catch (error) { message.value = String(error) }
}
</script>

<template>
  <div class="modal-backdrop" @click.self="emit('close')">
    <section class="modal-panel" role="dialog" aria-modal="true">
      <div class="title"><h2>{{ item ? '编辑订阅' : '添加 RSS' }}</h2><button class="modal-close" @click="emit('close')">×</button></div>
      <form @submit.prevent="save">
        <label>RSS URL<input v-model.trim="form.rssUrl" type="url" required></label>
        <label>正则表达式（可选）<input v-model="form.regex"></label>
        <label>排除正则（可选）<input v-model="form.excludeRegex" placeholder="例如：720|\d+-\d+"></label>
        <label>{{ item ? '目录名称' : '自定义目录名（可选）' }}<input v-model.trim="form.directoryName" :required="!!item"></label>
        <label>Season<input v-model.number="form.season" type="number" min="1" required></label>
        <label v-if="item" class="checkbox"><input v-model="form.enabled" type="checkbox">启用</label>
        <p v-if="message" class="message">{{ message }}</p>
        <div class="actions"><button type="submit">保存</button><button type="button" class="secondary" @click="emit('close')">取消</button></div>
      </form>
    </section>
  </div>
</template>
