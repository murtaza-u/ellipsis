NProgress.configure({ showSpinner: false })

document.addEventListener('htmx:beforeRequest', () => {
  NProgress.start()
})

document.addEventListener('htmx:afterRequest', () => {
  NProgress.done()
})
