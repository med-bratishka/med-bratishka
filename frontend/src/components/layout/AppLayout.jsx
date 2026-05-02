import { Navigate, Outlet } from 'react-router-dom'
import { useAuth } from '../../context/AuthContext'
import Sidebar from './Sidebar'

export default function AppLayout({ requiredRole }) {
  const { user, isLoading } = useAuth()

  // Пока проверяем localStorage — не редиректим, показываем заглушку
  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="w-6 h-6 border-2 border-brand-400 border-t-transparent rounded-full animate-spin" />
      </div>
    )
  }

  if (!user) return <Navigate to="/auth" replace />

  if (requiredRole && user.role !== requiredRole) {
    const fallback =
      user.role === 'admin' ? '/admin' :
      user.role === 'doctor' ? '/doctor' :
      '/patient'
    return <Navigate to={fallback} replace />
  }

  return (
    <div className="flex min-h-screen bg-gray-100">
      <Sidebar />
      <main className="flex-1 overflow-y-auto">
        <Outlet />
      </main>
    </div>
  )
}
