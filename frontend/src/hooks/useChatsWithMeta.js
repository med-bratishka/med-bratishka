import { useState, useEffect, useCallback } from 'react'
import { chatApi } from '../api/index'
import { useAuth } from '../context/AuthContext'
import { useNotificationsSocket } from './useNotificationsSocket'

export function useChatsWithMeta(pollInterval = 30000) {
  const { user } = useAuth()
  const [chats, setChats] = useState([])
  const [meta, setMeta] = useState({})
  const [loading, setLoading] = useState(true)

  const load = useCallback(async () => {
    try {
      const res = await chatApi.getChats()
      const arr = res.data?.items ?? res.data?.chats ?? res.data
      const list = Array.isArray(arr) ? arr : []

      const m = {}
      list.forEach((chat) => {
        m[chat.id] = {
          unread: chat.unread_count ?? 0,
          lastMsg: chat.last_message || '',
          lastTs: chat.last_message_at ?? chat.updated_at ?? 0,
          lastMsgId: chat.last_message_id ?? 0,
          lastReadMessageId: chat.last_read_message_id ?? 0,
        }
      })

      list.sort((a, b) => (m[b.id]?.lastTs ?? 0) - (m[a.id]?.lastTs ?? 0))
      setChats(list)
      setMeta(m)
    } catch (e) {
      console.error(e)
    } finally {
      setLoading(false)
    }
  }, [user?.id])

  useEffect(() => { load() }, [load])

  useEffect(() => {
    const t = setInterval(load, pollInterval)
    return () => clearInterval(t)
  }, [load, pollInterval])

  useNotificationsSocket({
    onChatNotification: () => {
      load()
    },
  })

  const markChatAsRead = useCallback(async (chatId) => {
    const lastMsgId = meta[chatId]?.lastMsgId
    try {
      await chatApi.markRead(chatId, lastMsgId)
    } catch (e) {
      console.error(e)
    }
    setMeta(prev => {
      return { ...prev, [chatId]: { ...prev[chatId], unread: 0, lastReadMessageId: lastMsgId || prev[chatId]?.lastReadMessageId || 0 } }
    })
  }, [meta])

  return { chats, meta, loading, reload: load, markChatAsRead }
}
