<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { api, type UpdateSettings } from '../api'

const form = reactive<UpdateSettings>({
  qbUrl: '', qbUsername: '', qbPassword: '', downloadRoot: '/downloads/anime',
  defaultCategory: 'MikanRSS', rssInterval: 30,
  defaultExcludeRegex: '',
})
const passwordSet = ref(false)
const message = ref('')
const rssProcessingEnabled = ref(false)
const rssAutoDownloadingEnabled = ref(false)
const qbRSSLoaded = ref(false)

async function loadQBRSS() {
  const data = await api.getQBRSSSettings()
  rssProcessingEnabled.value = data.processingEnabled ?? false
  rssAutoDownloadingEnabled.value = data.autoDownloadingEnabled ?? false
  form.rssInterval = data.refreshInterval ?? form.rssInterval
  qbRSSLoaded.value = true
}

onMounted(async () => {
  try {
    const data = await api.getSettings()
    Object.assign(form, {
      qbUrl: data.qbUrl,
      qbUsername: data.qbUsername,
      qbPassword: '',
      downloadRoot: data.downloadRoot,
      defaultCategory: data.defaultCategory,
      rssInterval: data.rssInterval,
      defaultExcludeRegex: data.defaultExcludeRegex,
    })
    passwordSet.value = data.passwordSet ?? false
    try { await loadQBRSS() } catch { /* qB 配置可在测试连接后读取 */ }
  } catch (error) { message.value = String(error) }
})

async function save() {
  try {
    const refreshInterval = form.rssInterval
    const data = await api.updateSettings(form)
    passwordSet.value = data.passwordSet ?? false
    form.qbPassword = ''
    if (!qbRSSLoaded.value) await loadQBRSS()
    form.rssInterval = refreshInterval
    await api.updateQBRSSSettings({
      processingEnabled: rssProcessingEnabled.value,
      autoDownloadingEnabled: rssAutoDownloadingEnabled.value,
      refreshInterval,
    })
    message.value = '设置已保存'
  } catch (error) { message.value = String(error) }
}

async function test() {
  try {
    const settings = await api.updateSettings(form)
    passwordSet.value = settings.passwordSet ?? false
    form.qbPassword = ''
    const data = await api.testQB()
    await loadQBRSS()
    message.value = `连接成功：qBittorrent ${data.version} / Web API ${data.webApiVersion}`
  } catch (error) { message.value = String(error) }
}
</script>

<template>
  <section class="panel">
    <h1>设置</h1>
    <form @submit.prevent="save">
      <label>qBittorrent 地址<input v-model.trim="form.qbUrl" type="url" required placeholder="http://qbittorrent:8080"></label>
      <label>用户名<input v-model="form.qbUsername" required></label>
      <label>密码<input v-model="form.qbPassword" type="password" :placeholder="passwordSet ? '已设置，留空则不修改' : '请输入密码'"></label>
      <label>下载根目录<input v-model.trim="form.downloadRoot" required></label>
      <label>默认分类<input v-model.trim="form.defaultCategory" required></label>
      <label>RSS 刷新间隔（分钟）<input v-model.number="form.rssInterval" type="number" min="1" required></label>
      <fieldset class="rss-settings">
        <legend>qBittorrent RSS 状态</legend>
        <label class="checkbox"><input v-model="rssProcessingEnabled" type="checkbox" :disabled="!qbRSSLoaded">获取 RSS 订阅</label>
        <label class="checkbox"><input v-model="rssAutoDownloadingEnabled" type="checkbox" :disabled="!qbRSSLoaded">RSS Torrent 自动下载</label>
        <small v-if="!qbRSSLoaded">保存连接信息并点击“测试连接”后读取状态</small>
      </fieldset>
      <label>默认排除正则<input v-model="form.defaultExcludeRegex" placeholder="例如：720|\d+-\d+"></label>
      <div class="actions"><button type="submit">保存设置</button><button type="button" class="secondary" @click="test">测试连接</button></div>
    </form>
    <p v-if="message" class="message">{{ message }}</p>
  </section>
</template>
