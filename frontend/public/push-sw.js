self.addEventListener('push', (event) => {
  let data = {}
  try {
    data = event.data ? event.data.json() : {}
  } catch {
    data = { title: 'MedBratishka', body: 'Новое уведомление' }
  }

  event.waitUntil(self.registration.showNotification(data.title || 'MedBratishka', {
    body: data.body || 'Новое уведомление',
    tag: data.tag || 'medbratishka',
    data: { url: data.url || '/' },
    renotify: true,
  }))
})

self.addEventListener('notificationclick', (event) => {
  event.notification.close()
  const target = new URL(event.notification.data?.url || '/', self.location.origin).href

  event.waitUntil(
    clients.matchAll({ type: 'window', includeUncontrolled: true }).then((windows) => {
      for (const client of windows) {
        if ('navigate' in client) client.navigate(target)
        if ('focus' in client) return client.focus()
      }
      return clients.openWindow ? clients.openWindow(target) : undefined
    })
  )
})
