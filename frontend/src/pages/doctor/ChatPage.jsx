import { useLocation } from 'react-router-dom'
import { ChatView } from '../../components/chat/ChatView'

export default function ChatPage() {
  const { chatId, chat } = useLocation().state || {}
  const patientName = chat?.other_name || chat?.other_login || (chat?.patient_id ? `Пациент #${chat.patient_id}` : null)

  return (
    <div className="h-screen">
      <ChatView
        chatId={chatId}
        chat={chat}
        otherName={patientName || 'Сообщения'}
        otherSub={patientName ? 'Чат с пациентом' : 'Выберите пациента из списка'}
      />
    </div>
  )
}
