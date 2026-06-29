<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { api, type UpdateSettings } from '../api'

const form = reactive<UpdateSettings>({
  qbUrl: '', qbUsername: '', qbPassword: '', downloadRoot: '/downloads/anime',
  defaultCategory: 'MikanRSS', rssInterval: 30,
})
const passwordSet = ref(false)
const message = ref('')

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
    })
    passwordSet.value = data.passwordSet ?? false
  } catch (error) { message.value = String(error) }
})

async function save() {
  try {
    const data = await api.updateSettings(form)
    passwordSet.value = data.passwordSet ?? false
    form.qbPassword = ''
    message.value = '设置已保存'
  } catch (error) { message.value = String(error) }
}

async function test() {
  try {
    const settings = await api.updateSettings(form)
    passwordSet.value = settings.passwordSet ?? false
    form.qbPassword = ''
    const data = await api.testQB()
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
      <div class="actions"><button type="submit">保存设置</button><button type="button" class="secondary" @click="test">测试连接</button></div>
    </form>
    <p v-if="message" class="message">{{ message }}</p>
  </section>
</template>
