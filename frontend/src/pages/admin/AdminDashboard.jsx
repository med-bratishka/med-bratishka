import { useState } from 'react'
import { adminApi } from '../../api/index'

export default function AdminDashboard() {
  const [clinicId, setClinicId] = useState('')
  const [doctorId, setDoctorId] = useState('')
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState(null)
  const [error, setError] = useState(null)

  const handleBind = async () => {
    if (!clinicId.trim() || !doctorId.trim()) {
      setError('Заполните оба поля')
      return
    }
    setLoading(true)
    setError(null)
    setResult(null)
    try {
      await adminApi.bindDoctorToClinic(Number(clinicId), Number(doctorId))
      setResult(`Врач #${doctorId} успешно привязан к клинике #${clinicId}`)
      setDoctorId('')
    } catch (err) {
      const msg = err?.response?.data?.message || err?.response?.data?.error || ''
      if (msg === 'forbidden') setError('Нет доступа — убедитесь что вы привязаны к этой клинике как администратор')
      else if (msg === 'doctor not found') setError('Врач с таким ID не найден')
      else setError(msg || 'Не удалось привязать врача')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="p-6 max-w-2xl">
      <div className="mb-6">
        <h1 className="text-lg font-medium text-gray-800">Панель администратора</h1>
        <p className="text-sm text-gray-400 mt-0.5">Управление клиникой и врачами</p>
      </div>

      {/* Info block */}
      <div className="bg-blue-50 border border-blue-100 rounded-xl p-4 mb-6">
        <p className="text-sm font-medium text-blue-700 mb-1">Как это работает</p>
        <p className="text-xs text-blue-600 leading-relaxed">
          Чтобы врач мог работать с пациентами, его нужно привязать к клинике. 
          После привязки врач сможет генерировать инвайт-коды и вести чаты с пациентами.
        </p>
      </div>

      {/* Bind doctor form */}
      <div className="bg-white border border-gray-100 rounded-xl p-5 mb-4">
        <p className="text-sm font-medium text-gray-700 mb-4">Привязать врача к клинике</p>

        <div className="grid grid-cols-2 gap-3 mb-4">
          <div className="flex flex-col gap-1.5">
            <label className="text-xs text-gray-500">ID клиники</label>
            <input
              className="input-field"
              type="number"
              placeholder="например: 1"
              value={clinicId}
              onChange={(e) => { setClinicId(e.target.value); setError(null); setResult(null) }}
            />
          </div>
          <div className="flex flex-col gap-1.5">
            <label className="text-xs text-gray-500">ID врача</label>
            <input
              className="input-field"
              type="number"
              placeholder="например: 2"
              value={doctorId}
              onChange={(e) => { setDoctorId(e.target.value); setError(null); setResult(null) }}
            />
          </div>
        </div>

        <button
          onClick={handleBind}
          disabled={loading || !clinicId || !doctorId}
          className="btn-primary w-full disabled:opacity-50"
        >
          {loading ? 'Привязываю...' : 'Привязать врача'}
        </button>

        {error && <p className="text-xs text-red-500 mt-3">{error}</p>}
        {result && <p className="text-xs text-green-600 mt-3">✓ {result}</p>}
      </div>

      {/* SQL hint */}
      <div className="bg-gray-50 border border-gray-100 rounded-xl p-4">
        <p className="text-xs font-medium text-gray-600 mb-2">Как узнать ID пользователей и клиник</p>
        <p className="text-xs text-gray-500 mb-2">Выполни в терминале:</p>
        <div className="bg-white border border-gray-100 rounded-lg p-3 font-mono text-xs text-gray-700 space-y-1">
          <p>docker exec -it medbratishka-postgres \</p>
          <p className="pl-4">psql -U postgres -d medbratishka</p>
          <p className="mt-2 text-gray-400">-- Посмотреть всех пользователей:</p>
          <p>SELECT id, email, role FROM users;</p>
          <p className="mt-1 text-gray-400">-- Посмотреть клиники:</p>
          <p>SELECT id, name FROM clinics;</p>
          <p className="mt-1 text-gray-400">-- Создать клинику если нет:</p>
          <p>INSERT INTO clinics (name, created_at, updated_at)</p>
          <p className="pl-4">VALUES ('Моя клиника', NOW(), NOW());</p>
        </div>
      </div>
    </div>
  )
}
