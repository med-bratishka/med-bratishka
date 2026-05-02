import { useState, useEffect, useRef } from 'react'
import { useLocation } from 'react-router-dom'
import { chatApi } from '../../api/index'
import { useAuth } from '../../context/AuthContext'

export default function PatientChatPage() {
  const location = useLocation()
  const { chatId, chat } = location.state || {}
  const { user } = useAuth()

  const [messages, setMessages] = useState([])
  const [text, setText] = useState('')
  const [loading, setLoading] = useState(false)
  const [sending, setSending] = useState(false)
  const bottomRef = useRef(null)

  const loadMessages = (silent = false) => {
    if (!chatId) return
    if (!silent) setLoading(true)
    return chatApi.getMessages(chatId)
      .then((res) => {
        const list = res.data?.items ?? res.data?.messages ?? res.data
        if (Array.isArray(list)) setMessages(list)
      })
      .catch(console.error)
      .finally(() => { if (!silent) setLoading(false) })
  }

  // Первичная загрузка
  useEffect(() => { loadMessages() }, [chatId])

  // Polling каждые 4 секунды
  useEffect(() => {
    if (!chatId) return
    const interval = setInterval(() => loadMessages(true), 4000)
    return () => clearInterval(interval)
  }, [chatId])

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
        setMessages((prev) => [...prev, msg])
      } else {
        await loadMessages(true)
      }
    } catch {
      setText(draft)
    } finally {
      setSending(false)
    }
  }

  const handleKeyDown = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); sendMessage() }
  }

  const doctorName = chat?.other_name || chat?.other_login || (chat?.doctor_id ? `Врач #${chat.doctor_id}` : null)

  return (
    <div className="flex flex-col h-screen">
      <div className="bg-white border-b border-gray-100 px-6 py-4">
        <p className="text-sm font-medium text-gray-800">{doctorName || 'Чат с врачом'}</p>
        <p className="text-xs text-gray-400">{doctorName ? 'Ваш лечащий врач' : 'Подключитесь к врачу чтобы начать общение'}</p>
      </div>

      <div className="flex-1 overflow-y-auto px-6 py-4 flex flex-col gap-3">
        {!chatId ? (
          <div className="flex-1 flex items-center justify-center">
            <div className="text-center">
              <p className="text-sm font-medium text-gray-600 mb-1">Сообщений пока нет</p>
              <p className="text-xs text-gray-400">Выберите чат на главной странице</p>
            </div>
          </div>
        ) : loading ? (
          <p className="text-sm text-gray-400 text-center mt-8">Загрузка...</p>
        ) : messages.length === 0 ? (
          <p className="text-sm text-gray-400 text-center mt-8">Нет сообщений. Напишите первым!</p>
        ) : (
          messages.map((msg, i) => {
            const isMine = user?.id
              ? msg.sender_id === user.id
              : false
            return (
              <div key={msg.id || i} className={`flex ${isMine ? 'justify-end' : 'justify-start'}`}>
                <div className={`max-w-[70%] px-4 py-2.5 rounded-2xl text-sm ${
                  isMine
                    ? 'bg-brand-400 text-white rounded-br-sm'
                    : 'bg-white border border-gray-100 text-gray-800 rounded-bl-sm'
                }`}>
                  {msg.content}
                  <p className={`text-xs mt-1 ${isMine ? 'text-blue-100' : 'text-gray-400'}`}>
                    {msg.created_at
                      ? new Date(msg.created_at).toLocaleTimeString('ru', { hour: '2-digit', minute: '2-digit' })
                      : ''}
                  </p>
                </div>
              </div>
            )
          })
        )}
        <div ref={bottomRef} />
      </div>

      <div className="bg-white border-t border-gray-100 px-6 py-4 flex gap-3 items-end">
        <textarea
          className="input-field flex-1 resize-none min-h-[40px] max-h-[120px]"
          rows={1}
          placeholder="Напишите сообщение..."
          value={text}
          onChange={(e) => setText(e.target.value)}
          onKeyDown={handleKeyDown}
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
