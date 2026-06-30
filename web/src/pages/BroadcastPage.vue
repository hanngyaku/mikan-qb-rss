<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { api, type Subscription } from '../api'

const days = [
  { value: '星期一', label: '一' }, { value: '星期二', label: '二' },
  { value: '星期三', label: '三' }, { value: '星期四', label: '四' },
  { value: '星期五', label: '五' }, { value: '星期六', label: '六' },
  { value: '星期日', label: '日' },
]
const items = ref<Subscription[]>([])
const message = ref('')
const today = new Date().getDay()
const todayValue = days[today === 0 ? 6 : today - 1].value
const unknown = computed(() => items.value.filter(item => !days.some(day => day.value === item.broadcastDay)))

onMounted(async () => {
  try { items.value = await api.listSubscriptions() }
  catch (error) { message.value = String(error) }
})

function startDrag(event: DragEvent, item: Subscription) {
  if (item.id) event.dataTransfer?.setData('text/plain', String(item.id))
}

async function drop(event: DragEvent, day: string) {
  const id = Number(event.dataTransfer?.getData('text/plain'))
  if (!id) return
  try {
    const updated = await api.updateBroadcastDay(id, day)
    const index = items.value.findIndex(item => item.id === id)
    if (index >= 0) items.value[index] = updated
    message.value = `已保存到${day}`
  } catch (error) { message.value = String(error) }
}
</script>

<template>
  <section>
    <div class="title"><div><h1>放送表</h1><p class="muted">按放送日期排列，可拖动调整</p></div></div>
    <p v-if="message" class="message">{{ message }}</p>
    <div class="broadcast-board">
      <section
        v-for="day in days"
        :key="day.value"
        class="broadcast-column"
        :class="{ today: day.value === todayValue }"
        @dragover.prevent
        @drop="drop($event, day.value)"
      >
        <h2>{{ day.label }} <span v-if="day.value === todayValue">今天</span></h2>
        <article
          v-for="item in items.filter(value => value.broadcastDay === day.value)"
          :key="item.id"
          class="broadcast-card"
          draggable="true"
          @dragstart="startDrag($event, item)"
        >
          <img v-if="item.posterUrl" :src="item.posterUrl" :alt="item.name">
          <div class="broadcast-card-name">{{ item.name }}</div>
        </article>
      </section>
    </div>
    <section v-if="unknown.length" class="unknown-broadcast" @dragover.prevent>
      <h2>未知 <small>拖拽到上方设置放送日</small></h2>
      <div class="unknown-list">
        <article v-for="item in unknown" :key="item.id" class="broadcast-card" draggable="true" @dragstart="startDrag($event, item)">
          <img v-if="item.posterUrl" :src="item.posterUrl" :alt="item.name">
          <div class="broadcast-card-name">{{ item.name }}</div>
        </article>
      </div>
    </section>
  </section>
</template>
