import { useState, useEffect, useCallback } from 'react'
import { chatApi } from '../api/index'
import { useAuth } from '../context/AuthContext'

const getLastSeen = (chatId) => parseInt(localStorage.getItem(`seen_${chatId}`) || '0', 10)

export const markAsRead = (chatId, lastMsgId) => {
  if (lastMsgId) localStorage.setItem(`seen_${chatId}`, String(lastMsgId))
}

export function useChatsWithMeta(pollInterval = 8000) {
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
      await Promise.all(list.map(async (chat) => {
        try {
          const r = await chatApi.getMessages(chat.id)
          const msgs = r.data?.items ?? r.data?.messages ?? r.data ?? []
          if (!Array.isArray(msgs) || msgs.length === 0) {
            m[chat.id] = { unread: 0, lastMsg: '', lastTs: chat.updated_at ?? 0, lastMsgId: 0 }
            return
          }
          // Бэк может возвращать сообщения в любом порядке — сортируем
          const sorted = [...msgs].sort((a, b) => (a.created_at ?? a.id ?? 0) - (b.created_at ?? b.id ?? 0))
          const last = sorted[sorted.length - 1]

          // Первый заход или некорректное значение — помечаем всё прочитанным
          const stored = localStorage.getItem(`seen_${chat.id}`)
          if (stored === null || parseInt(stored) > 1e12) {
            localStorage.setItem(`seen_${chat.id}`, String(last.id))
          }

          const lastSeen = getLastSeen(chat.id)
          // Считаем только чужие непрочитанные
          const unread = sorted.filter(msg => msg.id > lastSeen && msg.sender_id !== user?.id).length

          m[chat.id] = {
            unread,
            lastMsg: last.content || '',
            lastTs: last.created_at ?? chat.updated_at ?? 0,
            lastMsgId: last.id,
          }
        } catch {
          m[chat.id] = { unread: 0, lastMsg: '', lastTs: chat.updated_at ?? 0, lastMsgId: 0 }
        }
      }))

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

  const markChatAsRead = useCallback((chatId) => {
    setMeta(prev => {
      const lastMsgId = prev[chatId]?.lastMsgId
      if (lastMsgId) markAsRead(chatId, lastMsgId)
      return { ...prev, [chatId]: { ...prev[chatId], unread: 0 } }
    })
  }, [])

  return { chats, meta, loading, reload: load, markChatAsRead }
}
