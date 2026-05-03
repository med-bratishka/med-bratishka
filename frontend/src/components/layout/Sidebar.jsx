import { NavLink, useNavigate } from 'react-router-dom'
import { useAuth } from '../../context/AuthContext'

const DoctorNav = [
  { to: '/doctor', label: 'Пациенты', icon: (<svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5"><circle cx="8" cy="5" r="3"/><path d="M2 14c0-3.3 2.7-6 6-6s6 2.7 6 6"/></svg>) },
  { to: '/doctor/messages', label: 'Сообщения', icon: (<svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5"><path d="M14 10c0 .6-.5 1-1 1H4l-2 2V3c0-.6.5-1 1-1h10c.6 0 1 .4 1 1v7z"/></svg>) },
  { to: '/doctor/invite', label: 'Инвайт-коды', icon: (<svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5"><path d="M13 4H3a1 1 0 00-1 1v6a1 1 0 001 1h10a1 1 0 001-1V5a1 1 0 00-1-1z"/><path d="M3 4l5 4 5-4"/></svg>) },
]

const AdminNav = [
  { to: '/admin', label: 'Управление', icon: (<svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5"><rect x="2" y="2" width="5" height="5" rx="1"/><rect x="9" y="2" width="5" height="5" rx="1"/><rect x="2" y="9" width="5" height="5" rx="1"/><rect x="9" y="9" width="5" height="5" rx="1"/></svg>) },
]

const PatientNav = [
  { to: '/patient', label: 'Мой врач', icon: (<svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5"><rect x="3" y="1" width="10" height="14" rx="1.5"/><path d="M6 5h4M6 8h4M6 11h2"/></svg>) },
  { to: '/patient/chat', label: 'Чаты', icon: (<svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5"><path d="M14 10c0 .6-.5 1-1 1H4l-2 2V3c0-.6.5-1 1-1h10c.6 0 1 .4 1 1v7z"/></svg>) },
  { to: '/patient/reminders', label: 'Напоминания', icon: (<svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5"><path d="M8 1a5 5 0 015 5v3l1 1v1H2v-1l1-1V6a5 5 0 015-5zM6.5 13.5a1.5 1.5 0 003 0"/></svg>) },
]

export default function Sidebar() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()
  const isDoctor = user?.role === 'doctor'
  const isAdmin = user?.role === 'admin'
  const navItems = isAdmin ? AdminNav : isDoctor ? DoctorNav : PatientNav
  const roleLabel = isAdmin ? 'Администратор' : isDoctor ? 'Панель врача' : 'Панель пациента'
  const handleLogout = () => { logout(); navigate('/auth') }
  const name = user?.name || 'Пользователь'
  const initials = name.split(' ').map((w) => w[0]).filter(Boolean).slice(0, 2).join('').toUpperCase() || '??'

  return (
    <aside className="w-56 bg-gray-900 border-r border-gray-800 flex flex-col h-screen sticky top-0">
      <div className="px-5 py-5 border-b border-gray-800">
        <div className="flex items-center gap-2">
          <div className="w-7 h-7 bg-brand-400 rounded-lg flex items-center justify-center flex-shrink-0">
            <svg width="13" height="13" viewBox="0 0 16 16" fill="none"><path d="M8 2v12M2 8h12" stroke="white" strokeWidth="2.5" strokeLinecap="round"/></svg>
          </div>
          <span className="font-medium text-white text-sm">MedCare</span>
        </div>
        <p className="text-xs text-gray-500 mt-1 ml-9">{roleLabel}</p>
      </div>
      <nav className="flex-1 px-3 py-3 overflow-y-auto">
        <div className="flex flex-col gap-0.5">
          {navItems.map((item) => (
            <NavLink key={item.to} to={item.to} end={item.to === '/doctor' || item.to === '/patient' || item.to === '/admin'}
              className={({ isActive }) => `flex items-center gap-2.5 px-3 py-2 rounded-lg text-sm transition-colors ${isActive ? 'bg-gray-800 text-white font-medium' : 'text-gray-400 hover:bg-gray-800 hover:text-gray-200'}`}>
              {item.icon}{item.label}
            </NavLink>
          ))}
        </div>
      </nav>
      <div className="px-3 py-3 border-t border-gray-800">
        <div className="flex items-center gap-2.5 px-2 py-2">
          <div className="w-8 h-8 rounded-full bg-gray-700 flex items-center justify-center text-xs font-medium text-gray-200 flex-shrink-0">{initials}</div>
          <div className="flex-1 min-w-0">
            <p className="text-sm font-medium text-gray-200 truncate">{name}</p>
            <p className="text-xs text-gray-500 truncate">{isAdmin ? 'Администратор' : isDoctor ? 'Врач' : 'Пациент'}</p>
          </div>
          <button onClick={handleLogout} className="text-gray-600 hover:text-red-400 transition-colors">
            <svg width="15" height="15" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5"><path d="M6 2H3a1 1 0 00-1 1v10a1 1 0 001 1h3M11 11l3-3-3-3M14 8H6"/></svg>
          </button>
        </div>
      </div>
    </aside>
  )
}
