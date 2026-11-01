import { ref } from 'vue'

const notification = ref('')
let toastTimer: ReturnType<typeof setTimeout> | null = null

export function useToast() {
  function showToast(msg: string, duration = 4000) {
    if (toastTimer) {
      clearTimeout(toastTimer)
      toastTimer = null
    }
    notification.value = msg
    if (msg) {
      toastTimer = setTimeout(() => {
        notification.value = ''
        toastTimer = null
      }, duration)
    }
  }

  function clearToast() {
    if (toastTimer) {
      clearTimeout(toastTimer)
      toastTimer = null
    }
    notification.value = ''
  }

  return {
    notification,
    showToast,
    clearToast,
  }
}
