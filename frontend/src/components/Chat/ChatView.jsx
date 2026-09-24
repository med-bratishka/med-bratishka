import { useState, useEffect, useRef } from 'react'
import { chatApi } from '../../api/index'
import { useAuth } from '../../context/AuthContext'

const setLastSeen = (chatId, msgId) => localStorage.setItem(`seen_${chatId}`, String(msgId))

// Форматирует timestamp в строку времени
const toTime = (ts) => {
  if (!ts) return ''
  const ms = ts > 1e10 ? ts : ts * 1000
  return new Date(ms).toLocaleTimeString('ru', { hour: '2-digit', minute: '2-digit' })
}

// Форматирует дату для разделителя
const toDateLabel = (ts) => {
  if (!ts) return ''
  const ms = ts > 1e10 ? ts : ts * 1000
  const d = new Date(ms)
  const now = new Date()
  const yesterday = new Date(now)
  yesterday.setDate(now.getDate() - 1)
  if (d.toDateString() === now.toDateString()) return 'Сегодня'
  if (d.toDateString() === yesterday.toDateString()) return 'Вчера'
  return d.toLocaleDateString('ru', { day: 'numeric', month: 'long', year: d.getFullYear() !== now.getFullYear() ? 'numeric' : undefined })
}

// Группирует сообщения по дням
function groupByDate(messages) {
  const groups = []
  let lastLabel = null
  for (const msg of messages) {
    const label = toDateLabel(msg.created_at)
    if (label !== lastLabel) {
      groups.push({ type: 'divider', label, key: `d-${msg.id}` })
      lastLabel = label
    }
    groups.push({ type: 'msg', msg, key: msg.id })
  }
  return groups
}

export function ChatView({ chatId, chat, otherName, otherSub }) {
  const { user } = useAuth()
  const [messages, setMessages] = useState([])
  const [ready, setReady] = useState(false)
  const [text, setText] = useState('')
  const [sending, setSending] = useState(false)
  const bottomRef = useRef(null)
  const isFirstLoad = useRef(true)

  const scrollToBottom = (behavior = 'instant') => {
    bottomRef.current?.scrollIntoView({ behavior })
  }

  const loadMessages = (silent = false) => {
    if (!chatId) return Promise.resolve()
    return chatApi.getMessages(chatId)
      .then((res) => {
        const list = res.data?.items ?? res.data?.messages ?? res.data
        if (Array.isArray(list)) {
          setMessages(list)
          if (list.length > 0) setLastSeen(chatId, list[list.length - 1].id)
        }
      })
      .catch(console.error)
      .finally(() => { if (!silent) setReady(true) })
  }

  // При первой загрузке — скроллим вниз мгновенно
  useEffect(() => {
    if (ready && isFirstLoad.current) {
      isFirstLoad.current = false
      scrollToBottom('instant')
    }
  }, [ready, messages])

  useEffect(() => {
    setReady(false)
    setMessages([])
    isFirstLoad.current = true
    loadMessages(false)
  }, [chatId])

  // Polling — скроллим вниз только если уже были внизу
  useEffect(() => {
    if (!chatId) return
    const interval = setInterval(() => {
      const el = bottomRef.current?.parentElement
      const isAtBottom = !el || el.scrollHeight - el.scrollTop - el.clientHeight < 100
      loadMessages(true).then(() => {
        if (isAtBottom) scrollToBottom('smooth')
      })
    }, 4000)
    return () => clearInterval(interval)
  }, [chatId])

  const sendMessage = async () => {
    if (!text.trim() || !chatId || sending) return
    const draft = text.trim()
    setText('')
    setSending(true)
    try {
      const res = await chatApi.sendMessage(chatId, draft)
      const msg = res.data?.message || res.data
      if (msg?.id) {
        setMessages(prev => [...prev, msg])
        setLastSeen(chatId, msg.id)
      } else {
        await loadMessages(true)
      }
      // При отправке своего — всегда скроллим вниз
      setTimeout(() => scrollToBottom('smooth'), 50)
    } catch { setText(draft) }
    finally { setSending(false) }
  }

  const sorted = [...messages].sort((a, b) => (a.created_at ?? 0) - (b.created_at ?? 0))
  const items = groupByDate(sorted)

  return (
    <div className="flex flex-col h-full">
      {/* Шапка */}
      <div className="bg-white border-b border-gray-100 px-6 py-4 flex-shrink-0">
        <p className="text-sm font-medium text-gray-800">{otherName || 'Чат'}</p>
        {otherSub && <p className="text-xs text-gray-400 mt-0.5">{otherSub}</p>}
      </div>

      {/* Сообщения */}
      <div className="flex-1 overflow-y-auto px-6 py-4 flex flex-col">
        {!chatId ? (
          <div className="flex-1 flex items-center justify-center">
            <div className="text-center">
              <p className="text-sm font-medium text-gray-600 mb-1">Выберите диалог</p>
              <p className="text-xs text-gray-400">Перейдите в список чатов</p>
            </div>
          </div>
        ) : !ready ? (
          <div className="flex-1 flex items-center justify-center">
            <div className="flex gap-1">
              {[0, 1, 2].map(i => (
                <div key={i} className="w-2 h-2 rounded-full bg-gray-300 animate-bounce" style={{ animationDelay: `${i * 0.15}s` }} />
              ))}
            </div>
          </div>
        ) : messages.length === 0 ? (
          <div className="flex-1 flex items-center justify-center">
            <p className="text-sm text-gray-400">Нет сообщений. Напишите первым!</p>
          </div>
        ) : (
          <>
            <div className="flex-1" /> {/* толкает сообщения вниз */}
            {items.map((item) =>
              item.type === 'divider' ? (
                // Разделитель даты
                <div key={item.key} className="flex items-center gap-3 my-3">
                  <div className="flex-1 h-px bg-gray-100" />
                  <span className="text-xs text-gray-400 bg-white px-2 flex-shrink-0">{item.label}</span>
                  <div className="flex-1 h-px bg-gray-100" />
                </div>
              ) : (
                // Сообщение
                (() => {
                  const { msg } = item
                  const isMine = user?.id ? msg.sender_id === user.id : false
                  return (
                    <div key={item.key} className={`flex mb-1 ${isMine ? 'justify-end' : 'justify-start'}`}>
                      <div className={`max-w-[70%] px-4 py-2.5 rounded-2xl text-sm ${
                        isMine
                          ? 'bg-brand-400 text-white rounded-br-sm'
                          : 'bg-white border border-gray-100 text-gray-800 rounded-bl-sm shadow-sm'
                      }`}>
                        <p className="leading-relaxed">{msg.content}</p>
                        <p className={`text-xs mt-1 text-right ${isMine ? 'text-blue-100' : 'text-gray-400'}`}>
                          {toTime(msg.created_at)}
                        </p>
                      </div>
                    </div>
                  )
                })()
              )
            )}
            {/* Якорь — скролл идёт сюда */}
            <div ref={bottomRef} />
          </>
        )}
      </div>

      {/* Поле ввода */}
      <div className="bg-white border-t border-gray-100 px-6 py-4 flex gap-3 items-end flex-shrink-0">
        <textarea
          className="input-field flex-1 resize-none min-h-[40px] max-h-[120px]"
          rows={1}
          placeholder="Напишите сообщение..."
          value={text}
          onChange={(e) => setText(e.target.value)}
          onKeyDown={(e) => { if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); sendMessage() } }}
          disabled={!chatId || sending}
        />
        <button
          className="btn-primary px-4 py-2.5 flex-shrink-0 disabled:opacity-50"
          onClick={sendMessage}
          disabled={!chatId || !text.trim() || sending}
        >
          Отправить
        </button>
      </div>
    </div>
  )
}
