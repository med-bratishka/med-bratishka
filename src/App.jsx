import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { AuthProvider } from './context/AuthContext'
import AppLayout from './components/layout/AppLayout'
import AuthPage from './pages/AuthPage'
import DoctorDashboard from './pages/doctor/DoctorDashboard'
import ChatPage from './pages/doctor/ChatPage'
import MedicationsPage from './pages/doctor/MedicationsPage'
import InvitePage from './pages/doctor/InvitePage'
import PatientDashboard from './pages/patient/PatientDashboard'
import PatientChatPage from './pages/patient/PatientChatPage'
import PatientMedicationsPage from './pages/patient/PatientMedicationsPage'
import PatientRemindersPage from './pages/patient/PatientRemindersPage'

export default function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <Routes>
          <Route path="/auth" element={<AuthPage />} />
          <Route path="/" element={<Navigate to="/auth" replace />} />

          {/* Врач */}
          <Route element={<AppLayout requiredRole="doctor" />}>
            <Route path="/doctor" element={<DoctorDashboard />} />
            <Route path="/doctor/chat" element={<ChatPage />} />
            <Route path="/doctor/medications" element={<MedicationsPage />} />
            <Route path="/doctor/invite" element={<InvitePage />} />
          </Route>

          {/* Пациент */}
          <Route element={<AppLayout requiredRole="patient" />}>
            <Route path="/patient" element={<PatientDashboard />} />
            <Route path="/patient/chat" element={<PatientChatPage />} />
            <Route path="/patient/medications" element={<PatientMedicationsPage />} />
            <Route path="/patient/reminders" element={<PatientRemindersPage />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  )
}
