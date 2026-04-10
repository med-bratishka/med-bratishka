import { useNavigate } from 'react-router-dom'

export default function PatientDashboard() {
    const navigate = useNavigate()

    return (
        <div className="p-6 max-w-xl">
            <div className="mb-6">
                <h1 className="text-lg font-medium text-gray-800">Мой врач</h1>
                <p className="text-sm text-gray-400 mt-0.5">Информация о вашем враче</p>
            </div>
            <div className="bg-white border border-gray-100 rounded-xl p-10 flex flex-col items-center justify-center text-center">
                <div className="w-12 h-12 rounded-full bg-gray-100 flex items-center justify-center mb-3">
                    <svg width="20" height="20" viewBox="0 0 16 16" fill="none" stroke="#9ca3af" strokeWidth="1.5">
                        <rect x="3" y="1" width="10" height="14" rx="1.5"/>
                        <path d="M6 5h4M6 8h4M6 11h2"/>
                    </svg>
                </div>
                <p className="text-sm font-medium text-gray-600 mb-1">Врач не подключён</p>
                <p className="text-xs text-gray-400">Попросите врача выдать вам инвайт-код для подключения</p>
            </div>
        </div>
    )
}