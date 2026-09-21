import { useState } from 'react'
import { useNavigate } from 'react-router-dom'

export default function DoctorDashboard() {
    const [patients] = useState([])
    const navigate = useNavigate()

    return (
        <div className="p-6 max-w-3xl">
            <div className="flex items-center justify-between mb-6">
                <div>
                    <h1 className="text-lg font-medium text-gray-800">Мои пациенты</h1>
                    <p className="text-sm text-gray-400 mt-0.5">{patients.length} на курировании</p>
                </div>
                <button onClick={() => navigate('/doctor/invite')} className="btn-primary flex items-center gap-1.5">
                    <svg width="13" height="13" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round"><path d="M8 2v12M2 8h12"/></svg>
                    Инвайт-код
                </button>
            </div>
            <div className="grid grid-cols-3 gap-3 mb-6">
                {[{ label: 'Всего пациентов', value: 0 }, { label: 'Новых сообщений', value: 0 }, { label: 'Напоминаний сегодня', value: 0 }].map((m) => (
                    <div key={m.label} className="bg-white rounded-xl border border-gray-100 p-4">
                        <p className="text-xs text-gray-400 mb-1">{m.label}</p>
                        <p className="text-2xl font-medium text-gray-800">{m.value}</p>
                    </div>
                ))}
            </div>
            <div className="bg-white border border-gray-100 rounded-xl p-10 flex flex-col items-center justify-center text-center">
                <div className="w-12 h-12 rounded-full bg-gray-100 flex items-center justify-center mb-3">
                    <svg width="20" height="20" viewBox="0 0 16 16" fill="none" stroke="#9ca3af" strokeWidth="1.5"><circle cx="8" cy="5" r="3"/><path d="M2 14c0-3.3 2.7-6 6-6s6 2.7 6 6"/></svg>
                </div>
                <p className="text-sm font-medium text-gray-600 mb-1">Пациентов пока нет</p>
                <p className="text-xs text-gray-400">Создайте инвайт-код и отправьте его пациенту</p>
            </div>
        </div>
    )
}