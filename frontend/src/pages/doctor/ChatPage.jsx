import { useState, useEffect, useRef } from 'react'
import { useLocation } from 'react-router-dom'
import { chatApi } from '../../api/index'
import { useAuth } from '../../context/AuthContext'

const setLastSeen = (chatId, msgId) => localStorage.setItem(`seen_${chatId}`, String(msgId))

export default function ChatPage() {
  const location = useLocation()
  const { chatId, chat } = location.state || {}
  const { user } = useAuth()

  const [messages, setMessages] = useState([])
  const [text, setText] = useState('')
  const [loading, setLoading] = useState(true)
  const [sending, setSending] = useState(false)
  const bottomRef = useRef(null)

  const loadMessages = (silent = false) => {
    if (!chatId) return
    if (!silent) setLoading(true)
    return chatApi.getMessages(chatId)
      .then((res) => {
        const list = res.data?.items ?? res.data?.messages ?? res.data
        if (Array.isArray(list)) {
          setMessages(list)
          // Помечаем все как прочитанные
          if (list.length > 0) setLastSeen(chatId, list[list.length - 1].id)
        }
      })
      .catch(console.error)
      .finally(() => { if (!silent) setLoading(false) })
  }

  useEffect(() => { loadMessages() }, [chatId])

  // Polling каждые 4 секунды
  useEffect(() => {
    if (!chatId) return
    const interval = setInterval(() => loadMessages(true), 4000)
    return () => clearInterval(interval)
  }, [chatId])

  // Скролл вниз при новых сообщениях
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

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
    } catch {
      setText(draft)
    } finally {
      setSending(false)
    }
  }

  const patientName = chat?.other_name || chat?.other_login || (chat?.patient_id ? `Пациент #${chat.patient_id}` : null)

  const formatTime = (ts) => {
    if (!ts) return ''
    // created_at может быть в миллисекундах или секундах
    const ms = ts > 1e10 ? ts : ts * 1000
    return new Date(ms).toLocaleTimeString('ru', { hour: '2-digit', minute: '2-digit' })
  }

  return (
    <div className="flex flex-col h-screen">
      <div className="bg-white border-b border-gray-100 px-6 py-4 flex-shrink-0">
        <p className="text-sm font-medium text-gray-800">{patientName || 'Сообщения'}</p>
        <p className="text-xs text-gray-400">{patientName ? 'Чат с пациентом' : 'Выберите пациента из списка'}</p>
      </div>

      {/* Сообщения — flex-col, новые внизу */}
      <div className="flex-1 overflow-y-auto px-6 py-4 flex flex-col-reverse gap-2">
        {!chatId ? (
          <div className="flex-1 flex items-center justify-center h-full">
            <div className="text-center">
              <p className="text-sm font-medium text-gray-600 mb-1">Чат пока пуст</p>
              <p className="text-xs text-gray-400">Выберите пациента на главной странице</p>
            </div>
          </div>
        ) : loading ? (
          <div className="flex-1 flex items-center justify-center">
            <p className="text-sm text-gray-400">Загрузка...</p>
          </div>
        ) : messages.length === 0 ? (
          <div className="flex-1 flex items-center justify-center">
            <p className="text-sm text-gray-400">Нет сообщений. Напишите первым!</p>
          </div>
        ) : (
          <>
            {/* Пустой div растягивает пространство вверх, прижимая сообщения вниз */}
            <div className="flex-1" />
            {messages.map((msg, i) => {
              const isMine = user?.id ? msg.sender_id === user.id : false
              return (
                <div key={msg.id || i} className={`flex ${isMine ? 'justify-end' : 'justify-start'}`}>
                  <div className={`max-w-[70%] px-4 py-2.5 rounded-2xl text-sm ${
                    isMine
                      ? 'bg-brand-400 text-white rounded-br-sm'
                      : 'bg-white border border-gray-100 text-gray-800 rounded-bl-sm shadow-sm'
                  }`}>
                    <p>{msg.content}</p>
                    <p className={`text-xs mt-1 ${isMine ? 'text-blue-100' : 'text-gray-400'}`}>
                      {formatTime(msg.created_at)}
                    </p>
                  </div>
                </div>
              )
            })}
          </>
        )}
        <div ref={bottomRef} />
      </div>

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
