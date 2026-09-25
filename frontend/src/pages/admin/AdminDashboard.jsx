import { useState, useEffect } from 'react'
import { catalogApi, adminApi } from '../../api/index'

const ROLE_LABELS = { admin: 'Администратор', doctor: 'Врач', patient: 'Пациент' }
const ROLE_COLORS = {
  admin: 'bg-purple-100 text-purple-700',
  doctor: 'bg-blue-100 text-blue-700',
  patient: 'bg-green-100 text-green-700',
}

function Avatar({ name, id }) {
  const colors = [
    'bg-blue-100 text-blue-600', 'bg-violet-100 text-violet-600',
    'bg-emerald-100 text-emerald-600', 'bg-amber-100 text-amber-600', 'bg-rose-100 text-rose-600',
  ]
  const color = colors[(id ?? 0) % colors.length]
  const initials = name ? name.split(' ').map(w => w[0]).join('').slice(0, 2).toUpperCase() : '??'
  return (
    <div className={`w-10 h-10 rounded-full flex items-center justify-center flex-shrink-0 text-sm font-semibold ${color}`}>
      {initials}
    </div>
  )
}

function StatCard({ title, value, icon }) {
  return (
    <div className="bg-white border border-gray-100 rounded-xl p-5">
      <div className="flex items-center justify-between mb-3">
        <p className="text-xs text-gray-400">{title}</p>
        <div className="w-8 h-8 rounded-lg bg-brand-50 flex items-center justify-center">{icon}</div>
      </div>
      <p className="text-2xl font-semibold text-gray-800">{value}</p>
    </div>
  )
}

function DoctorTable({ doctors, loading }) {
  const [search, setSearch] = useState('')
  const filtered = doctors.filter(d => {
    const name = [d.first_name, d.last_name].filter(Boolean).join(' ')
    return name.toLowerCase().includes(search.toLowerCase()) ||
      (d.email || '').toLowerCase().includes(search.toLowerCase()) ||
      (d.specialization || '').toLowerCase().includes(search.toLowerCase())
  })

  if (loading) return (
    <div className="bg-white border border-gray-100 rounded-xl overflow-hidden">
      {[1,2,3].map(i => (
        <div key={i} className="flex items-center gap-3 px-5 py-3.5 border-b border-gray-50">
          <div className="w-10 h-10 rounded-full bg-gray-100 animate-pulse" />
          <div className="flex-1 space-y-1.5">
            <div className="h-3 bg-gray-100 rounded animate-pulse w-32" />
            <div className="h-2.5 bg-gray-100 rounded animate-pulse w-24" />
          </div>
        </div>
      ))}
    </div>
  )

  return (
    <div className="bg-white border border-gray-100 rounded-xl overflow-hidden">
      <div className="p-4 border-b border-gray-100">
        <input className="input-field w-full" placeholder="Поиск по имени, email или специализации..."
          value={search} onChange={e => setSearch(e.target.value)} />
      </div>
      <div className="overflow-x-auto">
        <table className="w-full">
          <thead className="bg-gray-50 border-b border-gray-100">
            <tr>
              <th className="text-left text-xs font-medium text-gray-500 px-5 py-3">Врач</th>
              <th className="text-left text-xs font-medium text-gray-500 px-5 py-3">Специализация</th>
              <th className="text-left text-xs font-medium text-gray-500 px-5 py-3">Телефон</th>
              <th className="text-left text-xs font-medium text-gray-500 px-5 py-3">Статус</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-50">
            {filtered.length === 0 ? (
              <tr><td colSpan="4" className="px-5 py-10 text-center text-sm text-gray-400">Врачи не найдены</td></tr>
            ) : filtered.map(d => {
              const name = [d.first_name, d.last_name].filter(Boolean).join(' ') || d.login || `Врач #${d.id}`
              return (
                <tr key={d.id} className="hover:bg-gray-50 transition-colors">
                  <td className="px-5 py-3.5">
                    <div className="flex items-center gap-3">
                      <Avatar name={name} id={d.id} />
                      <div>
                        <p className="text-sm font-medium text-gray-800">{name}</p>
                        <p className="text-xs text-gray-400">{d.email}</p>
                      </div>
                    </div>
                  </td>
                  <td className="px-5 py-3.5 text-sm text-gray-600">{d.specialization || '—'}</td>
                  <td className="px-5 py-3.5 text-sm text-gray-600">{d.phone || '—'}</td>
                  <td className="px-5 py-3.5">
                    <span className={`text-xs font-medium px-2.5 py-1 rounded-full ${d.is_verified ? 'bg-emerald-100 text-emerald-700' : 'bg-gray-100 text-gray-600'}`}>
                      {d.is_verified ? 'Верифицирован' : 'Не верифицирован'}
                    </span>
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>
      <div className="px-5 py-3 border-t border-gray-100 bg-gray-50">
        <p className="text-xs text-gray-500">Показано {filtered.length} из {doctors.length} врачей</p>
      </div>
    </div>
  )
}

function ClinicCard({ clinic, onBindDoctor }) {
  return (
    <div className="bg-white border border-gray-100 rounded-xl p-5 hover:shadow-md transition-shadow">
      <div className="flex items-start justify-between mb-3">
        <div>
          <h3 className="text-sm font-semibold text-gray-800">{clinic.name}</h3>
          <p className="text-xs text-gray-400 mt-0.5">{clinic.address || 'Адрес не указан'}</p>
        </div>
        <button onClick={() => onBindDoctor(clinic)}
          className="text-xs btn-primary px-3 py-1.5 flex items-center gap-1">
          <svg width="11" height="11" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round"><path d="M8 2v12M2 8h12"/></svg>
          Врача
        </button>
      </div>
      {clinic.email && <p className="text-xs text-gray-400 mt-1">{clinic.email}</p>}
      {clinic.phone && <p className="text-xs text-gray-400">{clinic.phone}</p>}
    </div>
  )
}

export default function AdminDashboard() {
  const [activeTab, setActiveTab] = useState('doctors')
  const [doctors, setDoctors] = useState([])
  const [clinics, setClinics] = useState([])
  const [loading, setLoading] = useState(true)
  const [showModal, setShowModal] = useState(false)
  const [selectedClinic, setSelectedClinic] = useState(null)
  const [selectedDoctorId, setSelectedDoctorId] = useState('')
  const [binding, setBinding] = useState(false)
  const [bindError, setBindError] = useState(null)
  const [bindSuccess, setBindSuccess] = useState(false)

  const loadData = async () => {
    setLoading(true)
    try {
      const [doctorsRes, clinicsRes] = await Promise.all([
        catalogApi.getDoctors(),
        catalogApi.getClinics(),
      ])
      const dl = doctorsRes.data?.items ?? doctorsRes.data?.doctors ?? doctorsRes.data ?? []
      const cl = clinicsRes.data?.items ?? clinicsRes.data?.clinics ?? clinicsRes.data ?? []
      setDoctors(Array.isArray(dl) ? dl : [])
      setClinics(Array.isArray(cl) ? cl : [])
    } catch (err) {
      console.error('Ошибка загрузки:', err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { loadData() }, [])

  const openBindModal = (clinic) => {
    setSelectedClinic(clinic)
    setSelectedDoctorId('')
    setBindError(null)
    setBindSuccess(false)
    setShowModal(true)
  }

  const handleBind = async () => {
    if (!selectedDoctorId) { setBindError('Выберите врача'); return }
    setBinding(true)
    setBindError(null)
    try {
      await adminApi.bindDoctorToClinic(selectedClinic.id, Number(selectedDoctorId))
      setBindSuccess(true)
      setTimeout(() => { setShowModal(false); setBindSuccess(false) }, 1500)
    } catch (err) {
      setBindError(err?.response?.data?.message || 'Не удалось привязать врача')
    } finally {
      setBinding(false)
    }
  }

  const tabs = [
    { id: 'doctors', label: 'Врачи' },
    { id: 'clinics', label: 'Клиники' },
  ]

  return (
    <div className="p-6 max-w-7xl">
      <div className="mb-6">
        <h1 className="text-lg font-medium text-gray-800">Панель администратора</h1>
        <p className="text-sm text-gray-400 mt-0.5">Управление врачами и клиниками</p>
      </div>

      {/* Статистика */}
      <div className="grid grid-cols-2 gap-3 mb-6">
        <StatCard
          title="Врачей в системе"
          value={loading ? '—' : doctors.length}
          icon={<svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="#059669" strokeWidth="1.5"><rect x="3" y="1" width="10" height="14" rx="1.5"/><path d="M6 5h4M6 8h4M6 11h2"/></svg>}
        />
        <StatCard
          title="Клиник"
          value={loading ? '—' : clinics.length}
          icon={<svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="#d97706" strokeWidth="1.5"><path d="M2 4h12v10H2z"/><path d="M8 2v12M4 8h8"/></svg>}
        />
      </div>

      {/* Табы */}
      <div className="bg-white border border-gray-100 rounded-xl">
        <div className="flex border-b border-gray-100">
          {tabs.map(tab => (
            <button key={tab.id} onClick={() => setActiveTab(tab.id)}
              className={`px-5 py-3 text-sm font-medium transition-colors ${
                activeTab === tab.id ? 'text-brand-600 border-b-2 border-brand-400' : 'text-gray-400 hover:text-gray-600'
              }`}>
              {tab.label}
            </button>
          ))}
        </div>

        <div className="p-6">
          {activeTab === 'doctors' && (
            <DoctorTable doctors={doctors} loading={loading} />
          )}

          {activeTab === 'clinics' && (
            <div>
              <div className="flex justify-between items-center mb-4">
                <p className="text-sm text-gray-600">{loading ? 'Загрузка...' : `${clinics.length} клиник`}</p>
              </div>
              {clinics.length === 0 && !loading ? (
                <div className="text-center py-10 text-sm text-gray-400">Клиники не найдены</div>
              ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                  {clinics.map(clinic => (
                    <ClinicCard key={clinic.id} clinic={clinic} onBindDoctor={openBindModal} />
                  ))}
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Модалка привязки врача */}
      {showModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-xl p-6 w-full max-w-md">
            <h2 className="text-base font-semibold text-gray-800 mb-1">Привязать врача к клинике</h2>
            <p className="text-sm text-gray-400 mb-4">{selectedClinic?.name}</p>

            <div>
              <label className="text-xs text-gray-500 mb-1 block">Выберите врача</label>
              <select
                className="input-field w-full"
                value={selectedDoctorId}
                onChange={e => { setSelectedDoctorId(e.target.value); setBindError(null) }}
              >
                <option value="">— выберите врача —</option>
                {doctors.map(d => {
                  const name = [d.first_name, d.last_name].filter(Boolean).join(' ') || d.login || `Врач #${d.id}`
                  const spec = d.specialization ? ` (${d.specialization})` : ''
                  return <option key={d.id} value={d.id}>{name}{spec}</option>
                })}
              </select>
            </div>

            {bindError && <p className="text-xs text-red-500 mt-2">{bindError}</p>}
            {bindSuccess && <p className="text-xs text-emerald-600 mt-2">✓ Врач успешно привязан!</p>}

            <div className="flex gap-2 mt-6">
              <button onClick={handleBind} disabled={binding || bindSuccess} className="btn-primary flex-1 disabled:opacity-50">
                {binding ? 'Привязываем...' : 'Привязать'}
              </button>
              <button onClick={() => setShowModal(false)} className="btn-secondary flex-1">Отмена</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
