import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { patientApi, chatApi } from '../../api/index'

const AVATAR_COLORS = [
  'bg-blue-100 text-blue-600',
  'bg-violet-100 text-violet-600',
  'bg-emerald-100 text-emerald-600',
  'bg-amber-100 text-amber-600',
  'bg-rose-100 text-rose-600',
  'bg-cyan-100 text-cyan-600',
]

function Avatar({ name, id }) {
  const color = AVATAR_COLORS[(id ?? 0) % AVATAR_COLORS.length]
  const initials = name ? name.split(' ').map((w) => w[0]).join('').slice(0, 2).toUpperCase() : '?'
  return (
    <div className={`w-10 h-10 rounded-full flex items-center justify-center flex-shrink-0 text-sm font-semibold ${color}`}>
      {initials}
    </div>
  )
}

function DoctorRow({ chat, onChat }) {
  const name = chat.other_name || chat.other_login || `Врач #${chat.doctor_id}`
  return (
    <button
      onClick={() => onChat(chat)}
      className="w-full flex items-center gap-3 px-5 py-3.5 hover:bg-gray-50 transition-colors text-left"
    >
      <Avatar name={name} id={chat.doctor_id} />
      <div className="flex-1 min-w-0">
        <p className="text-sm font-medium text-gray-800 truncate">{name}</p>
        <p className="text-xs text-gray-400 truncate mt-0.5">Нет сообщений</p>
      </div>
      <svg className="text-gray-300 flex-shrink-0" width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5"><path d="M6 4l4 4-4 4"/></svg>
    </button>
  )
}

export default function PatientDashboard() {
  const [chats, setChats] = useState([])
  const [loading, setLoading] = useState(true)
  const [bindCode, setBindCode] = useState('')
  const [binding, setBinding] = useState(false)
  const [bindError, setBindError] = useState(null)
  const [bound, setBound] = useState(false)
  const navigate = useNavigate()

  const loadChats = () => {
    setLoading(true)
    return chatApi.getChats()
      .then((res) => {
        const list = res.data?.items ?? res.data?.chats ?? res.data
        setChats(Array.isArray(list) ? list : [])
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
      setBound(true)
      setBindCode('')
      await new Promise((r) => setTimeout(r, 300))
      await loadChats()
    } catch (err) {
      const raw = JSON.stringify(err?.response?.data ?? err?.message ?? String(err), null, 2)
      setBindError('[HTTP ' + (err?.response?.status ?? 'network') + '] ' + raw)
    } finally {
      setBinding(false)
    }
  }

  return (
    <div className="p-6 max-w-xl">
      <div className="mb-6">
        <h1 className="text-lg font-medium text-gray-800">Мой врач</h1>
        <p className="text-sm text-gray-400 mt-0.5">Привязанные врачи и переписка</p>
      </div>

      {chats.length === 0 && !loading && !bound && (
        <div className="card mb-4">
          <p className="text-sm font-medium text-gray-700 mb-1">Привязаться к врачу</p>
          <p className="text-xs text-gray-400 mb-3">Введите код, который вам дал врач</p>
          <div className="flex gap-2">
            <input
              className="input-field flex-1 font-mono tracking-widest uppercase"
              placeholder="DR-XXXXXX"
              value={bindCode}
              onChange={(e) => { setBindCode(e.target.value.toUpperCase()); setBindError(null) }}
              onKeyDown={(e) => e.key === 'Enter' && handleBind()}
              disabled={binding}
            />
            <button onClick={handleBind} disabled={binding || !bindCode.trim()} className="btn-primary disabled:opacity-50">
              {binding ? '...' : 'Привязать'}
            </button>
          </div>
          {bindError && <p className="text-xs text-red-500 mt-2">{bindError}</p>}
        </div>
      )}

      {loading ? (
        <div className="bg-white border border-gray-100 rounded-xl divide-y divide-gray-50 overflow-hidden">
          {[1, 2].map((i) => (
            <div key={i} className="flex items-center gap-3 px-5 py-3.5">
              <div className="w-10 h-10 rounded-full bg-gray-100 animate-pulse flex-shrink-0" />
              <div className="flex-1 space-y-1.5">
                <div className="h-3 bg-gray-100 rounded animate-pulse w-32" />
                <div className="h-2.5 bg-gray-100 rounded animate-pulse w-20" />
              </div>
            </div>
          ))}
        </div>
      ) : bound && chats.length === 0 ? (
        <div className="card flex flex-col items-center text-center py-10">
          <div className="w-11 h-11 rounded-full bg-green-50 flex items-center justify-center mb-3">
            <svg width="18" height="18" viewBox="0 0 16 16" fill="none" stroke="#22c55e" strokeWidth="1.5"><path d="M3 8l4 4 6-7"/></svg>
          </div>
          <p className="text-sm font-medium text-gray-700 mb-1">Привязка выполнена</p>
          <p className="text-xs text-gray-400 mb-3">Чат появится после того, как врач его откроет</p>
          <button onClick={loadChats} className="text-xs text-brand-400 hover:underline">Обновить</button>
        </div>
      ) : chats.length === 0 ? (
        <div className="card flex flex-col items-center text-center py-10">
          <div className="w-11 h-11 rounded-full bg-gray-100 flex items-center justify-center mb-3">
            <svg width="18" height="18" viewBox="0 0 16 16" fill="none" stroke="#9ca3af" strokeWidth="1.5">
              <circle cx="8" cy="5" r="3"/><path d="M2 14c0-3.31 2.69-5 6-5s6 1.69 6 5"/>
            </svg>
          </div>
          <p className="text-sm font-medium text-gray-600 mb-1">Врач не подключён</p>
          <p className="text-xs text-gray-400">Введите код выше, чтобы привязаться к врачу</p>
        </div>
      ) : (
        <div className="bg-white border border-gray-100 rounded-xl divide-y divide-gray-50 overflow-hidden">
          {chats.map((chat) => (
            <DoctorRow
              key={chat.id}
              chat={chat}
              onChat={(c) => navigate('/patient/chat', { state: { chatId: c.id, chat: c } })}
            />
          ))}
        </div>
      )}
    </div>
  )
}
