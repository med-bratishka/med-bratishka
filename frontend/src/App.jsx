import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { AuthProvider } from './context/AuthContext'
import AppLayout from './components/layout/AppLayout'

import StaffAuthPage from './pages/StaffAuthPage'
import PatientAuthPage from './pages/PatientAuthPage'

import DoctorDashboard from './pages/doctor/DoctorDashboard'
import ChatPage from './pages/doctor/ChatPage'
import InvitePage from './pages/doctor/InvitePage'
import MessagesPage from './pages/doctor/MessagesPage'

import PatientDashboard from './pages/patient/PatientDashboard'
import PatientChatPage from './pages/patient/PatientChatPage'
import PatientMedicationsPage from './pages/patient/PatientMedicationsPage'
import PatientRemindersPage from './pages/patient/PatientRemindersPage'
import SecuritySettingsPage from './pages/SecuritySettingsPage'

import AdminDashboard from './pages/admin/AdminDashboard'

const APP_MODE = import.meta.env.VITE_APP_MODE || 'staff'
const isStaff = APP_MODE === 'staff'

export default function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <Routes>
          <Route
            path="/auth"
            element={isStaff ? <StaffAuthPage /> : <PatientAuthPage />}
          />

          <Route path="/" element={<Navigate to="/auth" replace />} />

          {isStaff && (
            <>
              <Route element={<AppLayout requiredRole="doctor" />}>
                <Route path="/doctor" element={<DoctorDashboard />} />
                <Route path="/doctor/chat" element={<ChatPage />} />
                <Route path="/doctor/invite" element={<InvitePage />} />
                <Route path="/doctor/messages" element={<MessagesPage />} />
                <Route path="/doctor/security" element={<SecuritySettingsPage />} />
              </Route>

              <Route element={<AppLayout requiredRole="admin" />}>
                <Route path="/admin" element={<AdminDashboard />} />
                <Route path="/admin/security" element={<SecuritySettingsPage />} />
              </Route>
            </>
          )}

          {!isStaff && (
            <Route element={<AppLayout requiredRole="patient" />}>
              <Route path="/patient" element={<PatientDashboard />} />
              <Route path="/patient/chat" element={<PatientChatPage />} />
              <Route path="/patient/medications" element={<PatientMedicationsPage />} />
              <Route path="/patient/reminders" element={<PatientRemindersPage />} />
              <Route path="/patient/security" element={<SecuritySettingsPage />} />
            </Route>
          )}

          <Route path="*" element={<Navigate to="/auth" replace />} />
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  )
}
