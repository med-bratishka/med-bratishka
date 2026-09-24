import { useState, useEffect, useRef } from 'react'
import { chatApi } from '../../api/index'
import { useAuth } from '../../context/AuthContext'

const toTime = (ts) => {
  if (!ts) return ''
  const ms = ts > 1e10 ? ts : ts * 1000
  return new Date(ms).toLocaleTimeString('ru', { hour: '2-digit', minute: '2-digit' })
}

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

export function ChatView({ chatId, chat, otherName, otherSub, onBack }) {
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
            if (list.length > 0) {
              const sortedList = [...list].sort((a, b) => (a.created_at ?? a.id ?? 0) - (b.created_at ?? b.id ?? 0))
              chatApi.markRead(chatId, sortedList[sortedList.length - 1].id).catch(console.error)
            }
          }
        })
        .catch(console.error)
        .finally(() => { if (!silent) setReady(true) })
  }

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
        chatApi.markRead(chatId, msg.id).catch(console.error)
      } else {
        await loadMessages(true)
      }
      setTimeout(() => scrollToBottom('smooth'), 50)
    } catch { setText(draft) }
    finally { setSending(false) }
  }

  const sorted = [...messages].sort((a, b) => (a.created_at ?? 0) - (b.created_at ?? 0))
  const items = groupByDate(sorted)

  return (
      <div className="flex flex-col h-screen bg-gradient-to-br from-black via-zinc-950 to-zinc-900">
        {/* Шапка */}
        <div className="header-luxury px-6 py-4 flex-shrink-0 flex items-center gap-3">
          {onBack && (
              <button onClick={onBack} className="w-9 h-9 rounded-xl flex items-center justify-center text-zinc-400 hover:text-amber-400 hover:bg-amber-500/10 transition-all duration-200 flex-shrink-0">
                <svg width="18" height="18" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round"><path d="M10 4L6 8l4 4"/></svg>
              </button>
          )}
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-amber-400 to-amber-600 flex items-center justify-center shadow-lg shadow-amber-900/30">
            <span className="text-black font-bold text-sm">
              {otherName ? otherName.split(' ').map(w => w[0]).join('').slice(0, 2).toUpperCase() : '?'}
            </span>
            </div>
            <div>
              <p className="text-sm font-bold text-gray-200">{otherName || 'Чат'}</p>
              {otherSub && <p className="text-xs text-zinc-500 mt-0.5">{otherSub}</p>}
            </div>
          </div>
        </div>

        {/* Сообщения */}
        <div className="flex-1 overflow-y-auto px-6 py-4 flex flex-col">
          {!chatId ? (
              <div className="flex-1 flex items-center justify-center">
                <div className="text-center">
                  <p className="text-sm font-semibold text-gray-400 mb-1">Выберите диалог</p>
                  <p className="text-xs text-zinc-600">Перейдите в список чатов</p>
                </div>
              </div>
          ) : !ready ? (
              <div className="flex-1 flex items-center justify-center">
                <div className="flex gap-2">
                  {[0, 1, 2].map(i => (
                      <div key={i} className="w-2.5 h-2.5 rounded-full bg-gradient-to-br from-amber-400 to-amber-600 animate-bounce shadow-lg shadow-amber-900/30" style={{ animationDelay: `${i * 0.15}s` }} />
                  ))}
                </div>
              </div>
          ) : messages.length === 0 ? (
              <div className="flex-1 flex items-center justify-center">
                <div className="text-center">
                  <p className="text-sm text-zinc-500">Нет сообщений</p>
                  <p className="text-xs text-zinc-600 mt-1">Напишите первым!</p>
                </div>
              </div>
          ) : (
              <>
                <div className="flex-1" />
                {items.map((item) =>
                    item.type === 'divider' ? (
                        <div key={item.key} className="flex items-center gap-3 my-4">
                          <div className="flex-1 h-px bg-gradient-to-r from-transparent via-zinc-700 to-transparent" />
                          <span className="text-xs text-zinc-500 bg-zinc-900/80 px-3 py-1 rounded-full border border-zinc-800">{item.label}</span>
                          <div className="flex-1 h-px bg-gradient-to-r from-transparent via-zinc-700 to-transparent" />
                        </div>
                    ) : (
                        (() => {
                          const { msg } = item
                          const isMine = user?.id ? msg.sender_id === user.id : false
                          return (
                              <div key={item.key} className={`flex mb-3 ${isMine ? 'justify-end' : 'justify-start'}`}>
                                <div className={`max-w-[75%] px-4 py-2.5 rounded-2xl text-sm shadow-lg ${
                                    isMine
                                        ? 'bg-gradient-to-br from-amber-500 to-amber-600 text-black rounded-br-sm shadow-amber-900/20'
                                        : 'bg-zinc-900/80 border border-zinc-800 text-gray-200 rounded-bl-sm backdrop-blur-sm'
                                }`}>
                                  <p className="leading-relaxed">{msg.content}</p>
                                  <p className={`text-xs mt-1 text-right ${isMine ? 'text-black/60' : 'text-zinc-500'}`}>
                                    {toTime(msg.created_at)}
                                  </p>
                                </div>
                              </div>
                          )
                        })()
                    )
                )}
                <div ref={bottomRef} />
              </>
          )}
        </div>

        {/* Поле ввода */}
        <div className="border-t border-amber-600/20 bg-zinc-900/50 backdrop-blur-md px-6 py-4 flex gap-3 items-end flex-shrink-0">
        <textarea
            className="input-field flex-1 resize-none min-h-[44px] max-h-[120px] text-gray-200"
            rows={1}
            placeholder="Напишите сообщение..."
            value={text}
            onChange={(e) => setText(e.target.value)}
            onKeyDown={(e) => { if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); sendMessage() } }}
            disabled={!chatId || sending}
        />
          <button
              className="btn-primary px-6 py-2.5 flex-shrink-0 disabled:opacity-50 disabled:cursor-not-allowed shadow-lg shadow-amber-900/30"
              onClick={sendMessage}
              disabled={!chatId || !text.trim() || sending}
          >
            {sending ? (
                <svg className="animate-spin h-5 w-5" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none"/>
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/>
                </svg>
            ) : (
                'Отправить'
            )}
          </button>
        </div>
      </div>
  )
}
