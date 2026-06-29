import { createRouter, createWebHistory } from 'vue-router'
import SubscriptionsPage from '../pages/SubscriptionsPage.vue'
import SettingsPage from '../pages/SettingsPage.vue'
import LogsPage from '../pages/LogsPage.vue'

export default createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: SubscriptionsPage },
    { path: '/settings', component: SettingsPage },
    { path: '/logs', component: LogsPage },
  ],
})
