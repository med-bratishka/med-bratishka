import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { patientApi, chatApi } from '../../api/index'

const AVATAR_COLORS = [
  'bg-gradient-to-br from-amber-400 to-amber-600 text-black',
  'bg-gradient-to-br from-violet-400 to-violet-600 text-white',
  'bg-gradient-to-br from-emerald-400 to-emerald-600 text-black',
  'bg-gradient-to-br from-rose-400 to-rose-600 text-white',
  'bg-gradient-to-br from-cyan-400 to-cyan-600 text-black',
]

function Avatar({ name, id }) {
  const color = AVATAR_COLORS[(id ?? 0) % AVATAR_COLORS.length]
  const initials = name ? name.split(' ').map((w) => w[0]).join('').slice(0, 2).toUpperCase() : '?'
  return (
      <div className={`w-12 h-12 rounded-xl flex items-center justify-center flex-shrink-0 text-sm font-bold shadow-lg ${color}`}>
        {initials}
      </div>
  )
}

function DoctorRow({ chat, onChat, onUnlink }) {
  const name = chat.other_name || chat.other_login || `Врач #${chat.doctor_id}`
  return (
      <div className="group relative bg-zinc-900/60 backdrop-blur-sm border border-zinc-800 rounded-xl p-4 hover:border-amber-500/50 hover:shadow-xl hover:shadow-amber-900/10 transition-all duration-300">
        <div className="flex items-center gap-4">
          <Avatar name={name} id={chat.doctor_id} />
          <div className="flex-1 min-w-0">
            <p className="text-sm font-semibold text-gray-200 truncate">{name}</p>
            <p className="text-xs text-zinc-500 mt-0.5">Нажмите для начала чата</p>
          </div>
          <div className="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-opacity duration-200">
            <button
                onClick={() => onChat(chat)}
                className="w-9 h-9 rounded-lg flex items-center justify-center text-zinc-500 hover:text-amber-400 hover:bg-amber-500/10 transition-colors"
                title="Открыть чат"
            >
              <svg width="18" height="18" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                <path d="M2 4.5C2 3.12 3.12 2 4.5 2h7C12.88 2 14 3.12 14 4.5v5c0 1.38-1.12 2.5-2.5 2.5H8l-3 2v-2H4.5C3.12 12 2 10.88 2 9.5v-5z"/>
              </svg>
            </button>
            <button
                onClick={() => onUnlink(chat)}
                className="w-9 h-9 rounded-lg flex items-center justify-center text-zinc-500 hover:text-red-400 hover:bg-red-500/10 transition-colors"
                title="Отвязать врача"
            >
              <svg width="18" height="18" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                <path d="M6.5 9.5l-2 2a2.121 2.121 0 000 3 2.121 2.121 0 003 0l2-2"/>
                <path d="M9.5 6.5l2-2a2.121 2.121 0 000-3 2.121 2.121 0 00-3 0l-2 2"/>
                <path d="M2 2l12 12"/>
              </svg>
            </button>
          </div>
          <svg className="text-zinc-700 flex-shrink-0 group-hover:text-amber-500/50 transition-colors" width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5"><path d="M6 4l4 4-4 4"/></svg>
        </div>
      </div>
  )
}

export default function PatientDashboard() {
  const [chats, setChats] = useState([])
  const [loading, setLoading] = useState(true)
  const [bindCode, setBindCode] = useState('')
  const [binding, setBinding] = useState(false)
  const [bindError, setBindError] = useState(null)
  const navigate = useNavigate()

  const loadChats = () => {
    setLoading(true)
    return chatApi.getChats()
        .then((res) => {
          const list = res.data?.items ?? res.data?.chats ?? res.data
          const chatsArray = Array.isArray(list) ? list : []
          setChats(chatsArray)
        })
        .catch(console.error)
        .finally(() => setLoading(false))
  }

  useEffect(() => { loadChats() }, [])

  const handleBind = async () => {
    if (!bindCode.trim()) return
    setBinding(true)
    setBindError(null)
    try {
      const bindRes = await patientApi.bindDoctor(bindCode.trim().toUpperCase())
      const doctorId = bindRes.data?.doctor_id ?? bindRes.data?.doctorId ?? null
      if (doctorId) {
        try { await chatApi.createChat(doctorId) } catch (e) {
          if (e?.response?.status !== 409) console.warn(e)
        }
      }
      setBindCode('')
      await new Promise((r) => setTimeout(r, 300))
      await loadChats()
    } catch (err) {
      const msg = err?.response?.data?.message?.toLowerCase?.() || ''
      const status = err?.response?.status

      if (status === 404 || msg.includes('not found') || msg.includes('invalid')) {
        setBindError('Неверный код приглашения')
      } else if (status === 409 || msg.includes('already')) {
        setBindError('Вы уже привязаны к этому врачу')
      } else if (status === 400) {
        setBindError('Код истёк или использован')
      } else {
        setBindError('Не удалось привязаться')
      }
    } finally {
      setBinding(false)
    }
  }

  const handleUnlink = async (chat) => {
    const name = chat.other_name || chat.other_login || `Врач #${chat.doctor_id}`
    if (!window.confirm(`Отвязаться от ${name}?`)) return
    try {
      await chatApi.closeChat(chat.id)
      await patientApi.unlinkDoctor(chat.doctor_id)
      await loadChats()
    } catch (err) {
      alert(err?.response?.data?.message || 'Не удалось отвязать врача')
    }
  }

  return (
      <div className="min-h-screen relative">
        {/* Уникальный фон с градиентными пятнами */}
        <div className="fixed inset-0 bg-gradient-to-br from-black via-zinc-950 to-zinc-900">
          <div className="absolute top-0 left-1/4 w-96 h-96 bg-amber-600/10 rounded-full blur-3xl"></div>
          <div className="absolute bottom-1/4 right-1/4 w-96 h-96 bg-amber-800/10 rounded-full blur-3xl"></div>
          <div className="absolute top-1/2 left-1/2 w-64 h-64 bg-violet-600/5 rounded-full blur-3xl"></div>
        </div>

        {/* Контент */}
        <div className="relative z-10 p-8 max-w-4xl">
          {/* Заголовок */}
          <div className="mb-8">
            <h1 className="text-3xl font-bold text-gold mb-2">Мои врачи</h1>
            <p className="text-sm text-zinc-500">Управление привязанными врачами и переписка</p>
          </div>

          {/* Форма добавления врача */}
          {chats.length < 3 && (
              <div className="mb-6">
                <div className="luxury-card border border-amber-600/20 shadow-xl shadow-amber-900/10">
                  <div className="flex items-center gap-2 mb-4">
                    <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-amber-400 to-amber-600 flex items-center justify-center">
                      <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="black" strokeWidth="2" strokeLinecap="round">
                        <path d="M8 2v12M2 8h12"/>
                      </svg>
                    </div>
                    <div>
                      <p className="text-sm font-semibold text-gray-200">
                        {chats.length === 0 ? 'Привязаться к врачу' : 'Добавить ещё врача'}
                      </p>
                      <p className="text-xs text-zinc-500">
                        {chats.length === 0 ? 'Введите код приглашения от врача' : `Добавлено врачей: ${chats.length} из 3`}
                      </p>
                    </div>
                  </div>

                  <div className="flex gap-3">
                    <input
                        className="input-field flex-1 font-mono tracking-widest uppercase"
                        placeholder="DR-XXXXXX"
                        value={bindCode}
                        onChange={(e) => { setBindCode(e.target.value.toUpperCase()); setBindError(null) }}
                        onKeyDown={(e) => e.key === 'Enter' && handleBind()}
                        disabled={binding}
                    />
                    <button
                        onClick={handleBind}
                        disabled={binding || !bindCode.trim()}
                        className="btn-primary disabled:opacity-50 disabled:cursor-not-allowed px-6"
                    >
                      {binding ? (
                          <span className="flex items-center gap-2">
                      <svg className="animate-spin h-4 w-4" viewBox="0 0 24 24">
                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none"/>
                        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/>
                      </svg>
                      ...
                    </span>
                      ) : 'Добавить'}
                    </button>
                  </div>

                  {bindError && (
                      <p className="text-xs text-red-400 bg-red-950/30 border border-red-900/50 rounded-lg px-3 py-2 mt-3">{bindError}</p>
                  )}
                </div>
              </div>
          )}

          {/* Загрузка */}
          {loading ? (
              <div className="space-y-3">
                {[1, 2].map((i) => (
                    <div key={i} className="luxury-card animate-pulse">
                      <div className="flex items-center gap-4">
                        <div className="w-12 h-12 rounded-xl bg-zinc-800" />
                        <div className="flex-1 space-y-2">
                          <div className="h-4 bg-zinc-800 rounded w-32" />
                          <div className="h-3 bg-zinc-800 rounded w-24" />
                        </div>
                      </div>
                    </div>
                ))}
              </div>
          ) : chats.length === 0 ? (
              /* Пустое состояние */
              <div className="luxury-card flex flex-col items-center text-center py-16 border border-zinc-800">
                <div className="w-16 h-16 rounded-2xl bg-gradient-to-br from-zinc-800 to-zinc-900 flex items-center justify-center mb-4 border border-zinc-700">
                  <svg width="28" height="28" viewBox="0 0 16 16" fill="none" stroke="#713f12" strokeWidth="1.5">
                    <circle cx="8" cy="5" r="3"/>
                    <path d="M2 14c0-3.31 2.69-5 6-5s6 1.69 6 5"/>
                  </svg>
                </div>
                <p className="text-base font-semibold text-gray-300 mb-2">Врачи не подключены</p>
                <p className="text-sm text-zinc-500 max-w-sm">
                  Введите код приглашения выше, чтобы привязаться к вашему врачу и начать общение
                </p>
              </div>
          ) : (
              /* Список врачей */
              <div className="space-y-3">
                <div className="flex items-center justify-between mb-2">
                  <p className="text-sm text-zinc-500">
                    {chats.length} {chats.length === 1 ? 'врач' : chats.length < 5 ? 'врача' : 'врачей'}
                  </p>
                  {chats.length >= 3 && (
                      <span className="text-xs text-amber-500/80 bg-amber-950/30 px-3 py-1 rounded-full border border-amber-600/20">
                  Достигнут лимит
                </span>
                  )}
                </div>
                {chats.map((chat) => (
                    <DoctorRow
                        key={chat.id}
                        chat={chat}
                        onChat={(c) => navigate('/patient/chat', { state: { chatId: c.id, chat: c } })}
                        onUnlink={handleUnlink}
                    />
                ))}
              </div>
          )}
        </div>
      </div>
  )
}