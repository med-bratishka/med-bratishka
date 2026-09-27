import { useCallback, useEffect, useState } from 'react'
import { webPushApi } from '../api/index'
import { useAuth } from '../context/AuthContext'

function isSupported() {
  return 'serviceWorker' in navigator && 'PushManager' in window && 'Notification' in window
}

function urlBase64ToUint8Array(value) {
  const padding = '='.repeat((4 - (value.length % 4)) % 4)
  const base64 = (value + padding).replace(/-/g, '+').replace(/_/g, '/')
  const raw = window.atob(base64)
  return Uint8Array.from([...raw].map(char => char.charCodeAt(0)))
}

async function getRegistration() {
  await navigator.serviceWorker.register('/push-sw.js')
  return navigator.serviceWorker.ready
}

export function useWebPush() {
  const { user } = useAuth()
  const [status, setStatus] = useState('loading')
  const [error, setError] = useState('')

  const refresh = useCallback(async () => {
    if (!user?.id) return
    if (user.role !== 'doctor' && user.role !== 'patient') {
      setStatus('unavailable')
      return
    }
    if (!isSupported()) {
      setStatus('unsupported')
      return
    }
    if (Notification.permission === 'denied') {
      setStatus('denied')
      return
    }
    try {
      const config = await webPushApi.getConfig()
      if (!config.data?.enabled || !config.data?.public_key) {
        setStatus('unavailable')
        return
      }
      const registration = await getRegistration()
      const subscription = await registration.pushManager.getSubscription()
      if (!subscription) {
        setStatus('disabled')
        return
      }
      await webPushApi.subscribe(subscription.toJSON())
      setStatus('enabled')
      setError('')
    } catch {
      setStatus('disabled')
      setError('Не удалось проверить push-уведомления')
    }
  }, [user?.id, user?.role])

  useEffect(() => {
    refresh()
  }, [refresh])

  const enable = useCallback(async () => {
    if (!isSupported()) {
      setStatus('unsupported')
      return
    }
    setStatus('loading')
    setError('')
    try {
      const permission = await Notification.requestPermission()
      if (permission !== 'granted') {
        setStatus(permission === 'denied' ? 'denied' : 'disabled')
        return
      }
      const config = await webPushApi.getConfig()
      if (!config.data?.enabled || !config.data?.public_key) {
        setStatus('unavailable')
        return
      }
      const registration = await getRegistration()
      const current = await registration.pushManager.getSubscription()
      const subscription = current || await registration.pushManager.subscribe({
        userVisibleOnly: true,
        applicationServerKey: urlBase64ToUint8Array(config.data.public_key),
      })
      await webPushApi.subscribe(subscription.toJSON())
      setStatus('enabled')
    } catch {
      setStatus('disabled')
      setError('Не удалось включить push-уведомления')
    }
  }, [])

  const disable = useCallback(async () => {
    if (!isSupported()) return
    setStatus('loading')
    setError('')
    try {
      const registration = await getRegistration()
      const subscription = await registration.pushManager.getSubscription()
      if (subscription) {
        await webPushApi.unsubscribe(subscription.endpoint)
        await subscription.unsubscribe()
      }
      setStatus('disabled')
    } catch {
      setStatus('enabled')
      setError('Не удалось отключить push-уведомления')
    }
  }, [])

  return { status, error, enable, disable }
}
