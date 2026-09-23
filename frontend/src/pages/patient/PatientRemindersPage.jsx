export default function PatientRemindersPage() {
    return (
        <div className="p-6 max-w-xl">
            <div className="mb-6">
                <h1 className="text-lg font-medium text-gray-800">Напоминания</h1>
                <p className="text-sm text-gray-400 mt-0.5">Расписание приёма лекарств</p>
            </div>
            <div className="bg-white border border-gray-100 rounded-xl p-10 flex flex-col items-center justify-center text-center">
                <div className="w-12 h-12 rounded-full bg-gray-100 flex items-center justify-center mb-3">
                    <svg width="20" height="20" viewBox="0 0 16 16" fill="none" stroke="#9ca3af" strokeWidth="1.5">
                        <path d="M8 1a5 5 0 015 5v3l1 1v1H2v-1l1-1V6a5 5 0 015-5zM6.5 13.5a1.5 1.5 0 003 0"/>
                    </svg>
                </div>
                <p className="text-sm font-medium text-gray-600 mb-1">Напоминаний пока нет</p>
                <p className="text-xs text-gray-400">Напоминания появятся когда врач назначит лекарства</p>
            </div>
        </div>
    )
}