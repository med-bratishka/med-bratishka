import { useState } from 'react'

export default function ChatPage() {
  const [text, setText] = useState('')

  return (
      <div className="flex flex-col h-screen">
        <div className="bg-white border-b border-gray-100 px-6 py-4">
          <p className="text-sm font-medium text-gray-800">Сообщения</p>
          <p className="text-xs text-gray-400">Выберите пациента из списка</p>
        </div>
        <div className="flex-1 flex items-center justify-center">
          <div className="text-center">
            <div className="w-12 h-12 rounded-full bg-gray-100 flex items-center justify-center mx-auto mb-3">
              <svg width="20" height="20" viewBox="0 0 16 16" fill="none" stroke="#9ca3af" strokeWidth="1.5">
                <path d="M14 10c0 .6-.5 1-1 1H4l-2 2V3c0-.6.5-1 1-1h10c.6 0 1 .4 1 1v7z"/>
              </svg>
            </div>
            <p className="text-sm font-medium text-gray-600 mb-1">Чат пока пуст</p>
            <p className="text-xs text-gray-400">Сообщения появятся когда подключатся пациенты</p>
          </div>
        </div>
        <div className="bg-white border-t border-gray-100 px-6 py-4 flex gap-3 items-end">
          <textarea className="input-field flex-1 resize-none min-h-[40px] max-h-[120px]" rows={1} placeholder="Напишите сообщение..." value={text} onChange={(e) => setText(e.target.value)} disabled />
          <button className="btn-primary px-4 py-2.5 flex-shrink-0 opacity-50" disabled>Отправить</button>
        </div>
      </div>
  )
}