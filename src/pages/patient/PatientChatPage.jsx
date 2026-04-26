import { useState, useEffect, useRef } from 'react'
import { useLocation } from 'react-router-dom'
import { chatApi } from '../../api/index'

export default function PatientChatPage() {
  const location = useLocation()
  const { chatId, chat } = location.state || {}

  const [messages, setMessages] = useState([])
  const [text, setText] = useState('')
  const [loading, setLoading] = useState(false)
  const [sending, setSending] = useState(false)
  const bottomRef = useRef(null)

  useEffect(() => {
    if (!chatId) return
    setLoading(true)
    chatApi.getMessages(chatId)
      .then((res) => { const data = res.data; const list = data?.messages ?? data; setMessages(Array.isArray(list) ? list : []) })
      .catch(console.error)
      .finally(() => setLoading(false))
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
      setMessages((prev) => [...prev, msg])
    } catch {
      setText(draft)
    } finally {
      setSending(false)
    }
  }

  const handleKeyDown = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); sendMessage() }
  }

  const doctorName = chat?.doctor_name || chat?.title || (chatId ? `Врач #${chatId}` : null)

  return (
    <div className="flex flex-col h-screen">
      <div className="bg-white border-b border-gray-100 px-6 py-4">
        <p className="text-sm font-medium text-gray-800">
          {doctorName || 'Чат с врачом'}
        </p>
        <p className="text-xs text-gray-400">
          {doctorName ? 'Ваш лечащий врач' : 'Подключитесь к врачу чтобы начать общение'}
        </p>
      </div>

      <div className="flex-1 overflow-y-auto px-6 py-4 flex flex-col gap-3">
        {!chatId ? (
          <div className="flex-1 flex items-center justify-center">
            <div className="text-center">
              <div className="w-12 h-12 rounded-full bg-gray-100 flex items-center justify-center mx-auto mb-3">
                <svg width="20" height="20" viewBox="0 0 16 16" fill="none" stroke="#9ca3af" strokeWidth="1.5">
                  <path d="M14 10c0 .6-.5 1-1 1H4l-2 2V3c0-.6.5-1 1-1h10c.6 0 1 .4 1 1v7z"/>
                </svg>
              </div>
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
            const isPatient = msg.sender_role === 'patient' || msg.is_mine
            return (
              <div key={msg.id || i} className={`flex ${isPatient ? 'justify-end' : 'justify-start'}`}>
                <div className={`max-w-[70%] px-4 py-2.5 rounded-2xl text-sm ${
                  isPatient
                    ? 'bg-brand-500 text-white rounded-br-sm'
                    : 'bg-white border border-gray-100 text-gray-800 rounded-bl-sm'
                }`}>
                  {msg.text || msg.content}
                  <p className={`text-xs mt-1 ${isPatient ? 'text-blue-100' : 'text-gray-400'}`}>
                    {msg.created_at ? new Date(msg.created_at).toLocaleTimeString('ru', { hour: '2-digit', minute: '2-digit' }) : ''}
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
