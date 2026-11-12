import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api, errText, type Account } from '../lib/api'

export const useSession = defineStore('session', () => {
  const initialized = ref(false)
  const unlocked = ref(false)
  const accounts = ref<Account[]>([])
  const busy = ref(false)
  const error = ref('')

  async function init() {
    try {
      initialized.value = await api.isInitialized()
    } catch (e) {
      error.value = errText(e)
    }
  }

  async function unlock(pw: string) {
    busy.value = true
    error.value = ''
    try {
      await api.unlock(pw)
      unlocked.value = true
      initialized.value = true
      await reloadAccounts()
    } catch (e) {
      error.value = errText(e)
      throw e
    } finally {
      busy.value = false
    }
  }

  async function lock() {
    await api.lock()
    unlocked.value = false
    accounts.value = []
    error.value = ''
  }

  async function reloadAccounts() {
    accounts.value = await api.listAccounts()
  }

  return { initialized, unlocked, accounts, busy, error, init, unlock, lock, reloadAccounts }
})
