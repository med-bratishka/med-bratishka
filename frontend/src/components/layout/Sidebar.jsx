import { NavLink, useNavigate } from 'react-router-dom'
import { useAuth } from '../../context/AuthContext'

const DoctorNav = [
  { to: '/doctor', label: 'Пациенты', icon: (<svg width="18" height="18" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5"><circle cx="8" cy="5" r="3"/><path d="M2 14c0-3.3 2.7-6 6-6s6 2.7 6 6"/></svg>) },
  { to: '/doctor/messages', label: 'Сообщения', icon: (<svg width="18" height="18" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5"><path d="M14 10c0 .6-.5 1-1 1H4l-2 2V3c0-.6.5-1 1-1h10c.6 0 1 .4 1 1v7z"/></svg>) },
  { to: '/doctor/invite', label: 'Инвайт-коды', icon: (<svg width="18" height="18" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5"><path d="M13 4H3a1 1 0 00-1 1v6a1 1 0 001 1h10a1 1 0 001-1V5a1 1 0 00-1-1z"/><path d="M3 4l5 4 5-4"/></svg>) },
]

const AdminNav = [
  { to: '/admin', label: 'Управление', icon: (<svg width="18" height="18" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5"><rect x="2" y="2" width="5" height="5" rx="1"/><rect x="9" y="2" width="5" height="5" rx="1"/><rect x="2" y="9" width="5" height="5" rx="1"/><rect x="9" y="9" width="5" height="5" rx="1"/></svg>) },
]

const PatientNav = [
  { to: '/patient', label: 'Мой врач', icon: (<svg width="18" height="18" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5"><rect x="3" y="1" width="10" height="14" rx="1.5"/><path d="M6 5h4M6 8h4M6 11h2"/></svg>) },
  { to: '/patient/chat', label: 'Чаты', icon: (<svg width="18" height="18" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5"><path d="M14 10c0 .6-.5 1-1 1H4l-2 2V3c0-.6.5-1 1-1h10c.6 0 1 .4 1 1v7z"/></svg>) },
]

export default function Sidebar() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()
  const isDoctor = user?.role === 'doctor'
  const isAdmin = user?.role === 'admin'
  const navItems = isAdmin ? AdminNav : isDoctor ? DoctorNav : PatientNav
  const roleLabel = isAdmin ? 'Администратор' : isDoctor ? 'Панель врача' : 'Панель пациента'
  const handleLogout = () => { logout(); navigate('/auth') }
  const displayName = user?.name || [user?.first_name, user?.last_name].filter(Boolean).join(' ') || user?.login || user?.email || 'Пользователь'
  const initials = displayName.split(' ').map((w) => w[0]).filter(Boolean).slice(0, 2).join('').toUpperCase() || (displayName[0] || '?').toUpperCase()

  return (
      <aside className="w-64 bg-gradient-to-b from-zinc-950 via-black to-zinc-950 border-r border-amber-600/20 flex flex-col h-screen sticky top-0 shadow-2xl shadow-black z-50">
        <div className="px-6 py-6 border-b border-amber-600/20 bg-gradient-to-r from-amber-600/10 to-transparent">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-gradient-to-br from-amber-400 via-amber-500 to-amber-600 rounded-xl flex items-center justify-center flex-shrink-0 shadow-lg shadow-amber-900/40 border border-amber-400/30">
              <svg width="18" height="18" viewBox="0 0 16 16" fill="none">
                <path d="M8 2v12M2 8h12" stroke="black" strokeWidth="2.5" strokeLinecap="round"/>
              </svg>
            </div>
            <div>
              <span className="font-bold text-white text-lg tracking-tight">MedBratishka</span>
              <p className="text-xs text-amber-500/90 mt-0.5 font-medium">{roleLabel}</p>
            </div>
          </div>
        </div>

        <nav className="flex-1 px-3 py-6 overflow-y-auto">
          <div className="flex flex-col gap-1">
            {navItems.map((item) => (
                <NavLink
                    key={item.to}
                    to={item.to}
                    end={item.to === '/doctor' || item.to === '/patient' || item.to === '/admin'}
                    className={({ isActive }) =>
                        `flex items-center gap-3 px-4 py-3 rounded-xl text-sm transition-all duration-300 group ${
                            isActive
                                ? 'bg-gradient-to-r from-amber-600/30 to-amber-600/10 text-amber-400 border border-amber-600/30 shadow-lg shadow-amber-900/20'
                                : 'text-zinc-400 hover:text-amber-400 hover:bg-zinc-900/60 hover:border hover:border-zinc-800'
                        }`
                    }
                >
                  {({ isActive }) => (
                      <>
                  <span className={isActive ? 'text-amber-400' : 'text-zinc-500 group-hover:text-amber-400 transition-colors'}>
                    {item.icon}
                  </span>
                        <span className="font-medium">{item.label}</span>
                      </>
                  )}
                </NavLink>
            ))}
          </div>
        </nav>

        <div className="px-3 py-4 border-t border-amber-600/20 bg-gradient-to-r from-amber-600/5 to-transparent">
          <div className="flex items-center gap-3 px-3 py-3 rounded-xl bg-zinc-900/80 border border-zinc-800 hover:border-amber-600/30 transition-all duration-300">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-amber-500 to-amber-700 flex items-center justify-center text-xs font-bold text-black flex-shrink-0 shadow-lg shadow-amber-900/30">
              {initials}
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-semibold text-gray-200 truncate">{displayName}</p>
              <p className="text-xs text-zinc-500 truncate">
                {isAdmin ? 'Администратор' : isDoctor ? 'Врач' : 'Пациент'}
              </p>
            </div>
            <button
                onClick={handleLogout}
                className="w-8 h-8 rounded-lg flex items-center justify-center text-zinc-600 hover:text-amber-500 hover:bg-amber-500/10 transition-all duration-200"
                title="Выйти"
            >
              <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
                <path d="M6 2H3a1 1 0 00-1 1v10a1 1 0 001 1h3M11 11l3-3-3-3M14 8H6"/>
              </svg>
            </button>
          </div>
        </div>
      </aside>
  )
}