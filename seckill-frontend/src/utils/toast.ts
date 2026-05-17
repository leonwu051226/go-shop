export type ToastType = 'info' | 'success' | 'warning' | 'error'

export const toast = (message: string, type: ToastType = 'info') => {
  window.dispatchEvent(
    new CustomEvent('app-toast', {
      detail: { message, type },
    }),
  )
}
