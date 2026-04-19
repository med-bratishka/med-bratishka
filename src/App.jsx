import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { AuthProvider } from './context/AuthContext'
import AppLayout from './components/layout/AppLayout'

// Страницы авторизации — разные в зависимости от режима
import StaffAuthPage from './pages/StaffAuthPage'
import PatientAuthPage from './pages/PatientAuthPage'

// Врач
import DoctorDashboard from './pages/doctor/DoctorDashboard'
import ChatPage from './pages/doctor/ChatPage'
import MedicationsPage from './pages/doctor/MedicationsPage'
import InvitePage from './pages/doctor/InvitePage'

// Пациент
import PatientDashboard from './pages/patient/PatientDashboard'
import PatientChatPage from './pages/patient/PatientChatPage'
import PatientMedicationsPage from './pages/patient/PatientMedicationsPage'
import PatientRemindersPage from './pages/patient/PatientRemindersPage'

/**
 * APP_MODE управляется переменной окружения VITE_APP_MODE:
 *   VITE_APP_MODE=staff   → портал врачей и администраторов
 *   VITE_APP_MODE=patient → портал пациентов
 *
 * По умолчанию — 'staff', чтобы не сломать текущий дев-запуск.
 */
const APP_MODE = import.meta.env.VITE_APP_MODE || 'staff'

const isStaff = APP_MODE === 'staff'

export default function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <Routes>
          {/* Авторизация — своя страница для каждого портала */}
          <Route
            path="/auth"
            element={isStaff ? <StaffAuthPage /> : <PatientAuthPage />}
          />

          {/* Редирект с корня на /auth */}
          <Route path="/" element={<Navigate to="/auth" replace />} />

          {/* ── STAFF-портал ──────────────────────────────────── */}
          {isStaff && (
            <>
              <Route element={<AppLayout requiredRole="doctor" />}>
                <Route path="/doctor" element={<DoctorDashboard />} />
                <Route path="/doctor/chat" element={<ChatPage />} />
                <Route path="/doctor/medications" element={<MedicationsPage />} />
                <Route path="/doctor/invite" element={<InvitePage />} />
              </Route>

              {/* TODO: добавить AdminDashboard когда будет готов */}
              <Route element={<AppLayout requiredRole="admin" />}>
                <Route path="/admin" element={<DoctorDashboard />} />
              </Route>
            </>
          )}

          {/* ── PATIENT-портал ────────────────────────────────── */}
          {!isStaff && (
            <Route element={<AppLayout requiredRole="patient" />}>
              <Route path="/patient" element={<PatientDashboard />} />
              <Route path="/patient/chat" element={<PatientChatPage />} />
              <Route path="/patient/medications" element={<PatientMedicationsPage />} />
              <Route path="/patient/reminders" element={<PatientRemindersPage />} />
            </Route>
          )}

          {/* Fallback */}
          <Route path="*" element={<Navigate to="/auth" replace />} />
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  )
}
