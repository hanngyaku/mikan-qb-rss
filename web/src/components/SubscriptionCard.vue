<script setup lang="ts">
import type { Subscription } from '../api'
defineProps<{ item: Subscription }>()
</script>

<template>
  <details class="card">
    <summary>{{ item.name }}</summary>
    <dl>
      <dt>目录名称</dt><dd>{{ item.saveDirName }}</dd>
      <dt>RSS 源</dt><dd><a :href="item.rssUrl" target="_blank">{{ item.rssUrl }}</a></dd>
      <dt>正则表达式</dt><dd>{{ item.regex || '无' }}</dd>
      <dt>保存路径</dt><dd>{{ item.savePath }}</dd>
      <dt>状态</dt><dd>{{ item.enabled ? '启用' : '停用' }}</dd>
    </dl>
  </details>
</template>
