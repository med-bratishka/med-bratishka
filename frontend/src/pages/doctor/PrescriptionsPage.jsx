import { useState, useEffect } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { prescriptionApi } from '../../api/index'

const FREQUENCY_OPTIONS = ['1 раз в день', '2 раза в день', '3 раза в день', 'По необходимости']

function PrescriptionCard({ item, onDelete }) {
  return (
    <div className="bg-white border border-gray-100 rounded-xl p-5 flex items-start justify-between gap-4">
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2 mb-1">
          <p className="text-sm font-medium text-gray-800">{item.medication_name || item.name}</p>
          <span className="text-xs bg-brand-50 text-brand-600 px-2 py-0.5 rounded-full">
            {item.dosage || item.dose}
          </span>
        </div>
        <p className="text-xs text-gray-400">{item.frequency} · {item.duration || 'бессрочно'}</p>
        {item.notes && <p className="text-xs text-gray-500 mt-1 italic">{item.notes}</p>}
      </div>
      <button
        onClick={() => onDelete(item.id)}
        className="w-7 h-7 rounded-lg flex items-center justify-center text-gray-300 hover:text-red-500 hover:bg-red-50 transition-colors flex-shrink-0"
        title="Удалить назначение"
      >
        <svg width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round">
          <path d="M3 4h10M6 4V2h4v2M5 4l.5 9h5l.5-9"/>
        </svg>
      </button>
    </div>
  )
}

export default function DoctorPrescriptionsPage() {
  const location = useLocation()
  const navigate = useNavigate()
  const { chat } = location.state || {}
  const patientId = chat?.patient_id
  const patientName = chat?.other_name || chat?.other_login || (patientId ? `Пациент #${patientId}` : null)

  const [prescriptions, setPrescriptions] = useState([])
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const [showForm, setShowForm] = useState(false)
  const [form, setForm] = useState({ medication_name: '', dosage: '', frequency: FREQUENCY_OPTIONS[0], duration: '', notes: '' })
  const [errors, setErrors] = useState({})

  useEffect(() => {
    if (!patientId) { setLoading(false); return }
    prescriptionApi.getForPatient(patientId)
      .then((res) => {
        const list = res.data?.items ?? res.data?.prescriptions ?? res.data
        setPrescriptions(Array.isArray(list) ? list : [])
      })
      .catch(console.error)
      .finally(() => setLoading(false))
  }, [patientId])

  const validate = () => {
    const e = {}
    if (!form.medication_name.trim()) e.medication_name = 'Введите название'
    if (!form.dosage.trim()) e.dosage = 'Введите дозировку'
    return e
  }

  const handleSubmit = async () => {
    const e = validate()
    if (Object.keys(e).length) { setErrors(e); return }
    setSubmitting(true)
    try {
      const res = await prescriptionApi.create({ ...form, patient_id: patientId })
      const created = res.data?.prescription || res.data
      setPrescriptions(prev => [...prev, created])
      setShowForm(false)
      setForm({ medication_name: '', dosage: '', frequency: FREQUENCY_OPTIONS[0], duration: '', notes: '' })
      setErrors({})
    } catch (err) {
      setErrors({ submit: err?.response?.data?.message || 'Ошибка при создании назначения' })
    } finally {
      setSubmitting(false)
    }
  }

  const handleDelete = async (id) => {
    if (!window.confirm('Удалить назначение?')) return
    try {
      await prescriptionApi.delete(id)
      setPrescriptions(prev => prev.filter(p => p.id !== id))
    } catch (err) {
      alert(err?.response?.data?.message || 'Не удалось удалить')
    }
  }

  return (
    <div className="p-6 max-w-2xl">
      <div className="flex items-center gap-3 mb-6">
        <button onClick={() => navigate(-1)} className="w-8 h-8 rounded-lg flex items-center justify-center text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition-colors">
          <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round"><path d="M10 4L6 8l4 4"/></svg>
        </button>
        <div>
          <h1 className="text-lg font-medium text-gray-800">Назначения</h1>
          <p className="text-sm text-gray-400 mt-0.5">{patientName || 'Пациент не выбран'}</p>
        </div>
        <div className="ml-auto">
          <button
            onClick={() => { setShowForm(true); setErrors({}) }}
            className="btn-primary flex items-center gap-1.5"
          >
            <svg width="13" height="13" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round"><path d="M8 2v12M2 8h12"/></svg>
            Добавить
          </button>
        </div>
      </div>

      {/* Форма создания */}
      {showForm && (
        <div className="bg-white border border-gray-100 rounded-xl p-5 mb-4">
          <p className="text-sm font-medium text-gray-700 mb-4">Новое назначение</p>
          <div className="grid grid-cols-2 gap-3 mb-3">
            <div>
              <label className="text-xs text-gray-500 mb-1 block">Препарат *</label>
              <input
                className={`input-field w-full ${errors.medication_name ? 'border-red-300' : ''}`}
                placeholder="Название препарата"
                value={form.medication_name}
                onChange={(e) => setForm(p => ({ ...p, medication_name: e.target.value }))}
              />
              {errors.medication_name && <p className="text-xs text-red-500 mt-1">{errors.medication_name}</p>}
            </div>
            <div>
              <label className="text-xs text-gray-500 mb-1 block">Дозировка *</label>
              <input
                className={`input-field w-full ${errors.dosage ? 'border-red-300' : ''}`}
                placeholder="напр. 500мг"
                value={form.dosage}
                onChange={(e) => setForm(p => ({ ...p, dosage: e.target.value }))}
              />
              {errors.dosage && <p className="text-xs text-red-500 mt-1">{errors.dosage}</p>}
            </div>
          </div>
          <div className="grid grid-cols-2 gap-3 mb-3">
            <div>
              <label className="text-xs text-gray-500 mb-1 block">Частота</label>
              <select
                className="input-field w-full"
                value={form.frequency}
                onChange={(e) => setForm(p => ({ ...p, frequency: e.target.value }))}
              >
                {FREQUENCY_OPTIONS.map(o => <option key={o}>{o}</option>)}
              </select>
            </div>
            <div>
              <label className="text-xs text-gray-500 mb-1 block">Длительность</label>
              <input
                className="input-field w-full"
                placeholder="напр. 7 дней"
                value={form.duration}
                onChange={(e) => setForm(p => ({ ...p, duration: e.target.value }))}
              />
            </div>
          </div>
          <div className="mb-4">
            <label className="text-xs text-gray-500 mb-1 block">Примечания</label>
            <textarea
              className="input-field w-full resize-none"
              rows={2}
              placeholder="Дополнительные инструкции..."
              value={form.notes}
              onChange={(e) => setForm(p => ({ ...p, notes: e.target.value }))}
            />
          </div>
          {errors.submit && <p className="text-xs text-red-500 mb-3">{errors.submit}</p>}
          <div className="flex gap-2">
            <button onClick={handleSubmit} disabled={submitting} className="btn-primary disabled:opacity-50">
              {submitting ? 'Сохранение...' : 'Сохранить'}
            </button>
            <button onClick={() => { setShowForm(false); setErrors({}) }} className="btn-secondary">
              Отмена
            </button>
          </div>
        </div>
      )}

      {/* Список назначений */}
      {loading ? (
        <div className="space-y-3">
          {[1, 2].map(i => (
            <div key={i} className="bg-white border border-gray-100 rounded-xl p-5">
              <div className="h-3 bg-gray-100 rounded animate-pulse w-40 mb-2" />
              <div className="h-2.5 bg-gray-100 rounded animate-pulse w-28" />
            </div>
          ))}
        </div>
      ) : !patientId ? (
        <div className="bg-white border border-gray-100 rounded-xl p-10 text-center">
          <p className="text-sm text-gray-500">Пациент не выбран. Перейдите из карточки пациента.</p>
        </div>
      ) : prescriptions.length === 0 ? (
        <div className="bg-white border border-gray-100 rounded-xl p-10 flex flex-col items-center text-center">
          <div className="w-12 h-12 rounded-full bg-gray-100 flex items-center justify-center mb-3">
            <svg width="20" height="20" viewBox="0 0 16 16" fill="none" stroke="#9ca3af" strokeWidth="1.5">
              <rect x="3" y="1.5" width="10" height="13" rx="1.5"/><path d="M6 5h4M6 8h4M6 11h2"/>
            </svg>
          </div>
          <p className="text-sm font-medium text-gray-600 mb-1">Назначений пока нет</p>
          <p className="text-xs text-gray-400">Нажмите «Добавить» чтобы создать первое назначение</p>
        </div>
      ) : (
        <div className="space-y-3">
          {prescriptions.map(p => (
            <PrescriptionCard key={p.id} item={p} onDelete={handleDelete} />
          ))}
        </div>
      )}
    </div>
  )
}
