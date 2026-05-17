import { useEffect, useRef } from 'react'
import { useAuth } from '../context/AuthContext'

const TOPIC_CHAT_NOTIFICATIONS = 'chat_notifications'

function getAccessToken() {
  const stored = localStorage.getItem('medcare_user')
  if (!stored) return ''
  try {
    return JSON.parse(stored)?.token || ''
  } catch {
    return ''
  }
}

function buildSocketUrl(token) {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const params = new URLSearchParams({ token })
  return `${protocol}//${window.location.host}/ws/notifications?${params.toString()}`
}

export function useNotificationsSocket({ onChatNotification } = {}) {
  const { user } = useAuth()
  const onChatNotificationRef = useRef(onChatNotification)

  useEffect(() => {
    onChatNotificationRef.current = onChatNotification
  }, [onChatNotification])

  useEffect(() => {
    const token = getAccessToken()
    if (!user?.id || !token) return undefined

    let socket
    let reconnectTimer
    let reconnectAttempt = 0
    let closedByHook = false

    const connect = () => {
      socket = new WebSocket(buildSocketUrl(token))

      socket.onopen = () => {
        reconnectAttempt = 0
        socket.send(JSON.stringify({ type: 'subscribe', topic: TOPIC_CHAT_NOTIFICATIONS }))
      }

      socket.onmessage = (event) => {
        let msg
        try {
          msg = JSON.parse(event.data)
        } catch {
          return
        }

        if (msg.type === 'notification' && msg.topic === TOPIC_CHAT_NOTIFICATIONS) {
          onChatNotificationRef.current?.(msg.data)
        }
      }

      socket.onclose = () => {
        if (closedByHook) return
        const delay = Math.min(30000, 1000 * (2 ** reconnectAttempt))
        reconnectAttempt += 1
        reconnectTimer = window.setTimeout(connect, delay)
      }
    }

    connect()

    return () => {
      closedByHook = true
      if (reconnectTimer) window.clearTimeout(reconnectTimer)
      if (socket && socket.readyState !== WebSocket.CLOSED) {
        socket.close()
      }
    }
  }, [user?.id])
}
