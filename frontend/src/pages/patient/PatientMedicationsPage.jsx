export default function PatientMedicationsPage() {
    return (
        <div className="p-6 max-w-xl">
            <div className="mb-6">
                <h1 className="text-lg font-medium text-gray-800">Мои лекарства</h1>
                <p className="text-sm text-gray-400 mt-0.5">Назначения от вашего врача</p>
            </div>
            <div className="bg-white border border-gray-100 rounded-xl p-10 flex flex-col items-center justify-center text-center">
                <div className="w-12 h-12 rounded-full bg-gray-100 flex items-center justify-center mb-3">
                    <svg width="20" height="20" viewBox="0 0 16 16" fill="none" stroke="#9ca3af" strokeWidth="1.5">
                        <path d="M11.5 4.5l-7 7M5 3a2 2 0 100 4 2 2 0 000-4zM11 9a2 2 0 100 4 2 2 0 000-4z"/>
                    </svg>
                </div>
                <p className="text-sm font-medium text-gray-600 mb-1">Назначений пока нет</p>
                <p className="text-xs text-gray-400">Лекарства появятся после назначения врачом</p>
            </div>
        </div>
    )
}