import { createRouter, createWebHistory } from 'vue-router'
import SubscriptionsPage from '../pages/SubscriptionsPage.vue'
import SettingsPage from '../pages/SettingsPage.vue'

export default createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: SubscriptionsPage },
    { path: '/settings', component: SettingsPage },
  ],
})
