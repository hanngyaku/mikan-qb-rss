<script setup lang="ts">
import { reactive, ref } from 'vue'
import { api, type Subscription, type UpdateSubscription } from '../api'

const props = defineProps<{ item: Subscription }>()
const emit = defineEmits<{ changed: []; deleted: []; message: [text: string] }>()
const editing = ref(false)
const showDelete = ref(false)
const form = reactive<UpdateSubscription>({
  rssUrl: props.item.rssUrl,
  regex: props.item.regex,
  excludeRegex: props.item.excludeRegex,
  saveDirName: props.item.saveDirName,
  season: props.item.season,
  enabled: props.item.enabled,
})

async function save() {
  if (!props.item.id) return
  try {
    await api.updateSubscription(props.item.id, form)
    editing.value = false
    emit('changed')
    emit('message', '订阅已保存并同步')
  } catch (error) { emit('message', String(error)) }
}

async function sync() {
  if (!props.item.id) return
  try {
    await api.syncSubscription(props.item.id)
    emit('message', '已同步到 qBittorrent')
  } catch (error) { emit('message', String(error)) }
}

async function remove() {
  if (!props.item.id) return
  try {
    await api.deleteSubscription(props.item.id)
    showDelete.value = false
    emit('deleted')
  } catch (error) { emit('message', String(error)) }
}
</script>

<template>
  <details class="card">
    <summary>
      <span class="subscription-name">{{ item.name }}</span>
    </summary>
    <form v-if="editing" @submit.prevent="save">
      <label>目录名称<input v-model.trim="form.saveDirName" required></label>
      <label>RSS 源<input v-model.trim="form.rssUrl" type="url" required></label>
      <label>正则表达式<input v-model="form.regex"></label>
      <label>排除正则<input v-model="form.excludeRegex"></label>
      <label>Season<input v-model.number="form.season" type="number" min="1" required></label>
      <label class="checkbox"><input v-model="form.enabled" type="checkbox">启用</label>
      <div class="actions"><button type="submit">保存并同步</button><button type="button" class="secondary" @click="editing = false">取消</button></div>
    </form>
    <template v-else>
      <dl>
        <dt>目录名称</dt><dd>{{ item.saveDirName }}</dd>
        <dt>RSS 源</dt><dd><a :href="item.rssUrl" target="_blank">{{ item.rssUrl }}</a></dd>
        <dt>正则表达式</dt><dd>{{ item.regex || '无' }}</dd>
        <dt>排除正则</dt><dd>{{ item.excludeRegex || '无' }}</dd>
        <dt>保存路径</dt><dd>{{ item.savePath }}</dd>
        <dt>Season</dt><dd>{{ item.season }}</dd>
        <dt>放送日期</dt><dd>{{ item.broadcastDay || '未知' }}</dd>
        <dt>放送开始</dt><dd>{{ item.broadcastStart || '未知' }}</dd>
        <dt>官方网站</dt><dd><a v-if="item.officialUrl" :href="item.officialUrl" target="_blank">{{ item.officialUrl }}</a><span v-else>无</span></dd>
        <dt>Bangumi</dt><dd><a v-if="item.bangumiUrl" :href="item.bangumiUrl" target="_blank">{{ item.bangumiUrl }}</a><span v-else>无</span></dd>
        <dt>概况介绍</dt><dd class="description">{{ item.description || '暂无介绍' }}</dd>
        <dt>状态</dt><dd>{{ item.enabled ? '启用' : '停用' }}</dd>
      </dl>
      <div class="actions">
        <button @click="editing = true">编辑</button>
        <button class="secondary" @click="sync">同步</button>
        <button class="danger" @click="showDelete = true">删除</button>
      </div>
    </template>
  </details>
  <div v-if="showDelete" class="modal-backdrop" @click.self="showDelete = false">
    <section class="modal-panel confirm-panel" role="alertdialog" aria-modal="true">
      <h2>删除订阅</h2>
      <p>确定删除“{{ item.name }}”吗？对应的 qBittorrent RSS 源和规则也会被删除。</p>
      <div class="actions">
        <button class="secondary" @click="showDelete = false">取消</button>
        <button class="danger" @click="remove">确认删除</button>
      </div>
    </section>
  </div>
</template>
