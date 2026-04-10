import { useState } from 'react'

export default function MedicationsPage() {
  const [meds, setMeds] = useState([])
  const [showForm, setShowForm] = useState(false)
  const [form, setForm] = useState({ name: '', dose: '', times: '', reminder: true })

  const handleAdd = () => {
    if (!form.name.trim()) return
    setMeds((prev) => [...prev, { id: Date.now(), ...form }])
    setForm({ name: '', dose: '', times: '', reminder: true })
    setShowForm(false)
  }

  const deleteMed = (id) => setMeds((prev) => prev.filter((m) => m.id !== id))
  const toggleReminder = (id) => setMeds((prev) => prev.map((m) => m.id === id ? { ...m, reminder: !m.reminder } : m))

  return (
      <div className="p-6 max-w-2xl">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-lg font-medium text-gray-800">Назначения</h1>
            <p className="text-sm text-gray-400 mt-0.5">Управление лекарствами пациента</p>
          </div>
          <button onClick={() => setShowForm(true)} className="btn-primary flex items-center gap-1.5">
            <svg width="13" height="13" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round"><path d="M8 2v12M2 8h12"/></svg>
            Добавить
          </button>
        </div>
        {showForm && (
            <div className="bg-white border border-brand-100 rounded-xl p-4 mb-4 flex flex-col gap-3">
              <p className="text-sm font-medium text-gray-700">Новое лекарство</p>
              <input className="input-field" placeholder="Название и дозировка" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} />
              <input className="input-field" placeholder="Режим (напр. 2 раза в день)" value={form.dose} onChange={(e) => setForm({ ...form, dose: e.target.value })} />
              <input className="input-field" placeholder="Время (напр. 08:00, 20:00)" value={form.times} onChange={(e) => setForm({ ...form, times: e.target.value })} />
              <label className="flex items-center gap-2 text-sm text-gray-600 cursor-pointer">
                <input type="checkbox" checked={form.reminder} onChange={(e) => setForm({ ...form, reminder: e.target.checked })} />
                Включить напоминание пациенту
              </label>
              <div className="flex gap-2">
                <button onClick={handleAdd} className="btn-primary">Добавить</button>
                <button onClick={() => setShowForm(false)} className="btn-secondary">Отмена</button>
              </div>
            </div>
        )}
        {meds.length === 0 && !showForm && (
            <div className="bg-white border border-gray-100 rounded-xl p-10 flex flex-col items-center justify-center text-center">
              <div className="w-12 h-12 rounded-full bg-gray-100 flex items-center justify-center mb-3">
                <svg width="20" height="20" viewBox="0 0 16 16" fill="none" stroke="#9ca3af" strokeWidth="1.5"><path d="M11.5 4.5l-7 7M5 3a2 2 0 100 4 2 2 0 000-4zM11 9a2 2 0 100 4 2 2 0 000-4z"/></svg>
              </div>
              <p className="text-sm font-medium text-gray-600 mb-1">Назначений пока нет</p>
              <p className="text-xs text-gray-400">Нажмите «Добавить» чтобы назначить лекарство</p>
            </div>
        )}
        <div className="flex flex-col gap-2">
          {meds.map((med) => (
              <div key={med.id} className="bg-white border border-gray-100 rounded-xl px-4 py-3.5 flex items-center gap-3">
                <div className="w-9 h-9 rounded-lg bg-brand-50 flex items-center justify-center flex-shrink-0">
                  <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="#378add" strokeWidth="1.5"><path d="M11.5 4.5l-7 7M5 3a2 2 0 100 4 2 2 0 000-4zM11 9a2 2 0 100 4 2 2 0 000-4z"/></svg>
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium text-gray-800">{med.name}</p>
                  <p className="text-xs text-gray-400">{med.dose}{med.times ? ` · ${med.times}` : ''}</p>
                </div>
                <button onClick={() => toggleReminder(med.id)} className={`status-badge text-xs cursor-pointer ${med.reminder ? 'bg-emerald-50 text-emerald-700' : 'bg-gray-50 text-gray-400'}`}>
                  {med.reminder ? 'Напоминание вкл' : 'Без напоминания'}
                </button>
                <button onClick={() => deleteMed(med.id)} className="text-gray-300 hover:text-red-400 transition-colors ml-1">
                  <svg width="15" height="15" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5"><path d="M2 4h12M5 4V2h6v2M6 7v5M10 7v5M3 4l1 9a1 1 0 001 1h6a1 1 0 001-1l1-9"/></svg>
                </button>
              </div>
          ))}
        </div>
      </div>
  )
}