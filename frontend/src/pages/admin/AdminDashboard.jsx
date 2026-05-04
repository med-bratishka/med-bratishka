import { useState, useEffect } from 'react'
import { adminApi } from '../../api/index'

const ROLE_COLORS = {
  admin: 'bg-purple-100 text-purple-700',
  doctor: 'bg-blue-100 text-blue-700',
  patient: 'bg-green-100 text-green-700',
}

const STATUS_COLORS = {
  active: 'bg-emerald-100 text-emerald-700',
  inactive: 'bg-gray-100 text-gray-600',
  blocked: 'bg-red-100 text-red-700',
}

function Avatar({ name, id }) {
  const colors = [
    'bg-blue-100 text-blue-600',
    'bg-violet-100 text-violet-600',
    'bg-emerald-100 text-emerald-600',
    'bg-amber-100 text-amber-600',
    'bg-rose-100 text-rose-600',
  ]
  const color = colors[(id ?? 0) % colors.length]
  const initials = name
      ? name.split(' ').map(w => w[0]).join('').slice(0, 2).toUpperCase()
      : '??'
  return (
      <div className={`w-10 h-10 rounded-full flex items-center justify-center flex-shrink-0 text-sm font-semibold ${color}`}>
        {initials}
      </div>
  )
}

function Badge({ children, color }) {
  return (
      <span className={`text-xs font-medium px-2.5 py-1 rounded-full ${color}`}>
      {children}
    </span>
  )
}

function StatCard({ title, value, icon, trend }) {
  return (
      <div className="bg-white border border-gray-100 rounded-xl p-5">
        <div className="flex items-center justify-between mb-3">
          <p className="text-xs text-gray-400">{title}</p>
          <div className="w-8 h-8 rounded-lg bg-brand-50 flex items-center justify-center">
            {icon}
          </div>
        </div>
        <p className="text-2xl font-semibold text-gray-800">{value}</p>
        {trend && (
            <p className={`text-xs mt-1 ${trend > 0 ? 'text-emerald-600' : 'text-red-600'}`}>
              {trend > 0 ? '↑' : '↓'} {Math.abs(trend)}% за неделю
            </p>
        )}
      </div>
  )
}

function UserTable({ users, loading, onEdit, onDelete, onBlock }) {
  const [search, setSearch] = useState('')
  const [filterRole, setFilterRole] = useState('all')

  const filtered = users.filter(user => {
    const matchesSearch = (user.name || '').toLowerCase().includes(search.toLowerCase()) ||
        (user.email || '').toLowerCase().includes(search.toLowerCase())
    const matchesRole = filterRole === 'all' || user.role === filterRole
    return matchesSearch && matchesRole
  })

  if (loading) {
    return (
        <div className="bg-white border border-gray-100 rounded-xl overflow-hidden">
          {[1, 2, 3].map(i => (
              <div key={i} className="flex items-center gap-3 px-5 py-3.5 border-b border-gray-50">
                <div className="w-10 h-10 rounded-full bg-gray-100 animate-pulse" />
                <div className="flex-1 space-y-1.5">
                  <div className="h-3 bg-gray-100 rounded animate-pulse w-32" />
                  <div className="h-2.5 bg-gray-100 rounded animate-pulse w-24" />
                </div>
              </div>
          ))}
        </div>
    )
  }

  return (
      <div className="bg-white border border-gray-100 rounded-xl overflow-hidden">
        <div className="p-4 border-b border-gray-100 flex flex-col sm:flex-row gap-3">
          <input
              className="input-field flex-1"
              placeholder="Поиск по имени или email..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
          />
          <select
              className="input-field w-full sm:w-40"
              value={filterRole}
              onChange={(e) => setFilterRole(e.target.value)}
          >
            <option value="all">Все роли</option>
            <option value="admin">Администраторы</option>
            <option value="doctor">Врачи</option>
            <option value="patient">Пациенты</option>
          </select>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-gray-50 border-b border-gray-100">
            <tr>
              <th className="text-left text-xs font-medium text-gray-500 px-5 py-3">Пользователь</th>
              <th className="text-left text-xs font-medium text-gray-500 px-5 py-3">Роль</th>
              <th className="text-left text-xs font-medium text-gray-500 px-5 py-3">Статус</th>
              <th className="text-left text-xs font-medium text-gray-500 px-5 py-3">Дата регистрации</th>
              <th className="text-right text-xs font-medium text-gray-500 px-5 py-3">Действия</th>
            </tr>
            </thead>
            <tbody className="divide-y divide-gray-50">
            {filtered.length === 0 ? (
                <tr>
                  <td colSpan="5" className="px-5 py-10 text-center text-sm text-gray-400">
                    Пользователи не найдены
                  </td>
                </tr>
            ) : (
                filtered.map(user => (
                    <tr key={user.id} className="hover:bg-gray-50 transition-colors">
                      <td className="px-5 py-3.5">
                        <div className="flex items-center gap-3">
                          <Avatar name={user.name} id={user.id} />
                          <div>
                            <p className="text-sm font-medium text-gray-800">{user.name || '—'}</p>
                            <p className="text-xs text-gray-400">{user.email}</p>
                          </div>
                        </div>
                      </td>
                      <td className="px-5 py-3.5">
                        <Badge color={ROLE_COLORS[user.role] || 'bg-gray-100 text-gray-600'}>
                          {user.role === 'admin' ? 'Администратор' :
                              user.role === 'doctor' ? 'Врач' : 'Пациент'}
                        </Badge>
                      </td>
                      <td className="px-5 py-3.5">
                        <Badge color={STATUS_COLORS[user.is_blocked ? 'blocked' : 'active']}>
                          {user.is_blocked ? 'Заблокирован' : 'Активен'}
                        </Badge>
                      </td>
                      <td className="px-5 py-3.5 text-sm text-gray-600">
                        {new Date(user.created_at).toLocaleDateString('ru')}
                      </td>
                      <td className="px-5 py-3.5">
                        <div className="flex items-center justify-end gap-1">
                          <button
                              onClick={() => onEdit(user)}
                              className="w-8 h-8 rounded-lg flex items-center justify-center text-gray-400 hover:text-brand-600 hover:bg-brand-50 transition-colors"
                              title="Редактировать"
                          >
                            <svg width="15" height="15" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
                              <path d="M11.5 2.5l2 2L6 12H4v-2l7.5-7.5zM3 13h10"/>
                            </svg>
                          </button>
                          <button
                              onClick={() => onBlock(user)}
                              className="w-8 h-8 rounded-lg flex items-center justify-center text-gray-400 hover:text-amber-600 hover:bg-amber-50 transition-colors"
                              title={user.is_blocked ? 'Разблокировать' : 'Заблокировать'}
                          >
                            {user.is_blocked ? (
                                <svg width="15" height="15" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
                                  <path d="M3 8l4 4 6-7"/>
                                </svg>
                            ) : (
                                <svg width="15" height="15" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
                                  <rect x="3" y="6" width="10" height="8" rx="1"/>
                                  <path d="M5 6V4a3 3 0 016 0v2"/>
                                </svg>
                            )}
                          </button>
                          <button
                              onClick={() => onDelete(user)}
                              className="w-8 h-8 rounded-lg flex items-center justify-center text-gray-400 hover:text-red-500 hover:bg-red-50 transition-colors"
                              title="Удалить"
                          >
                            <svg width="15" height="15" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
                              <path d="M3 4h10M6 4V2h4v2M5 4l.5 9h5l.5-9"/>
                            </svg>
                          </button>
                        </div>
                      </td>
                    </tr>
                ))
            )}
            </tbody>
          </table>
        </div>
        <div className="px-5 py-3 border-t border-gray-100 bg-gray-50">
          <p className="text-xs text-gray-500">
            Показано {filtered.length} из {users.length} пользователей
          </p>
        </div>
      </div>
  )
}

function ClinicCard({ clinic, doctorsCount, patientsCount, onEdit, onDelete }) {
  return (
      <div className="bg-white border border-gray-100 rounded-xl p-5 hover:shadow-md transition-shadow">
        <div className="flex items-start justify-between mb-3">
          <div>
            <h3 className="text-sm font-semibold text-gray-800">{clinic.name}</h3>
            <p className="text-xs text-gray-400 mt-0.5">ID: {clinic.id}</p>
          </div>
          <div className="flex gap-1">
            <button
                onClick={() => onEdit(clinic)}
                className="w-7 h-7 rounded-lg flex items-center justify-center text-gray-400 hover:text-brand-600 hover:bg-brand-50 transition-colors"
            >
              <svg width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
                <path d="M11.5 2.5l2 2L6 12H4v-2l7.5-7.5zM3 13h10"/>
              </svg>
            </button>
            <button
                onClick={() => onDelete(clinic)}
                className="w-7 h-7 rounded-lg flex items-center justify-center text-gray-400 hover:text-red-500 hover:bg-red-50 transition-colors"
            >
              <svg width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
                <path d="M3 4h10M6 4V2h4v2M5 4l.5 9h5l.5-9"/>
              </svg>
            </button>
          </div>
        </div>
        <div className="flex gap-3 mt-4">
          <div className="flex-1 bg-blue-50 rounded-lg px-3 py-2 text-center">
            <p className="text-lg font-semibold text-blue-700">{doctorsCount}</p>
            <p className="text-xs text-blue-600">врачей</p>
          </div>
          <div className="flex-1 bg-green-50 rounded-lg px-3 py-2 text-center">
            <p className="text-lg font-semibold text-green-700">{patientsCount}</p>
            <p className="text-xs text-green-600">пациентов</p>
          </div>
        </div>
      </div>
  )
}

export default function AdminDashboard() {
  const [activeTab, setActiveTab] = useState('users')
  const [users, setUsers] = useState([])
  const [clinics, setClinics] = useState([])
  const [loading, setLoading] = useState(true)
  const [showModal, setShowModal] = useState(false)
  const [modalType, setModalType] = useState('')
  const [formData, setFormData] = useState({})
  const [stats, setStats] = useState({
    totalUsers: 0,
    totalDoctors: 0,
    totalPatients: 0,
    totalClinics: 0,
  })

  const loadData = async () => {
    setLoading(true)
    try {
      const [usersRes, clinicsRes] = await Promise.all([
        adminApi.getUsers(),
        adminApi.getClinics()
      ])

      const usersList = usersRes.data?.items ?? usersRes.data?.users ?? usersRes.data ?? []
      const clinicsList = clinicsRes.data?.items ?? clinicsRes.data?.clinics ?? clinicsRes.data ?? []

      setUsers(Array.isArray(usersList) ? usersList : [])
      setClinics(Array.isArray(clinicsList) ? clinicsList : [])

      setStats({
        totalUsers: usersList.length,
        totalDoctors: usersList.filter(u => u.role === 'doctor').length,
        totalPatients: usersList.filter(u => u.role === 'patient').length,
        totalClinics: clinicsList.length,
      })
    } catch (err) {
      console.error('Ошибка загрузки:', err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadData()
  }, [])

  const handleEditUser = (user) => {
    setModalType('editUser')
    setFormData(user)
    setShowModal(true)
  }

  const handleBlockUser = async (user) => {
    const action = user.is_blocked ? 'разблокировать' : 'заблокировать'
    if (!window.confirm(`${action} пользователя ${user.name}?`)) return

    try {
      await adminApi.blockUser(user.id, !user.is_blocked)
      await loadData()
    } catch (err) {
      alert('Не удалось изменить статус пользователя')
    }
  }

  const handleDeleteUser = async (user) => {
    if (!window.confirm(`Удалить пользователя ${user.name}? Это действие нельзя отменить.`)) return

    try {
      await adminApi.deleteUser(user.id)
      await loadData()
    } catch (err) {
      alert('Не удалось удалить пользователя')
    }
  }

  const handleBindDoctor = async () => {
    if (!formData.clinicId || !formData.doctorId) {
      alert('Заполните оба поля')
      return
    }

    try {
      await adminApi.bindDoctorToClinic(Number(formData.clinicId), Number(formData.doctorId))
      setShowModal(false)
      setFormData({})
      await loadData()
      alert('Врач успешно привязан к клинике')
    } catch (err) {
      alert(err?.response?.data?.message || 'Не удалось привязать врача')
    }
  }

  const handleCreateClinic = async () => {
    if (!formData.name?.trim()) {
      alert('Введите название клиники')
      return
    }

    try {
      await adminApi.createClinic(formData.name)
      setShowModal(false)
      setFormData({})
      await loadData()
    } catch (err) {
      alert('Не удалось создать клинику')
    }
  }

  return (
      <div className="p-6 max-w-7xl">
        <div className="mb-6">
          <h1 className="text-lg font-medium text-gray-800">Панель администратора</h1>
          <p className="text-sm text-gray-400 mt-0.5">Управление пользователями, врачами и клиниками</p>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3 mb-6">
          <StatCard
              title="Всего пользователей"
              value={stats.totalUsers}
              trend={12}
              icon={
                <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="#378add" strokeWidth="1.5">
                  <circle cx="8" cy="5" r="3"/>
                  <path d="M2 14c0-3.3 2.7-6 6-6s6 2.7 6 6"/>
                </svg>
              }
          />
          <StatCard
              title="Врачей"
              value={stats.totalDoctors}
              trend={5}
              icon={
                <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="#059669" strokeWidth="1.5">
                  <rect x="3" y="1" width="10" height="14" rx="1.5"/>
                  <path d="M6 5h4M6 8h4M6 11h2"/>
                </svg>
              }
          />
          <StatCard
              title="Пациентов"
              value={stats.totalPatients}
              trend={18}
              icon={
                <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="#7c3aed" strokeWidth="1.5">
                  <circle cx="8" cy="5" r="3"/>
                  <path d="M2 14c0-3.3 2.7-6 6-6s6 2.7 6 6"/>
                </svg>
              }
          />
          <StatCard
              title="Клиник"
              value={stats.totalClinics}
              icon={
                <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="#d97706" strokeWidth="1.5">
                  <path d="M2 4h12v10H2z"/>
                  <path d="M8 2v12M4 8h8"/>
                </svg>
              }
          />
        </div>

        <div className="bg-white border border-gray-100 rounded-xl mb-6">
          <div className="flex border-b border-gray-100">
            <button
                onClick={() => setActiveTab('users')}
                className={`px-5 py-3 text-sm font-medium transition-colors ${
                    activeTab === 'users'
                        ? 'text-brand-600 border-b-2 border-brand-400'
                        : 'text-gray-400 hover:text-gray-600'
                }`}
            >
              Пользователи
            </button>
            <button
                onClick={() => setActiveTab('clinics')}
                className={`px-5 py-3 text-sm font-medium transition-colors ${
                    activeTab === 'clinics'
                        ? 'text-brand-600 border-b-2 border-brand-400'
                        : 'text-gray-400 hover:text-gray-600'
                }`}
            >
              Клиники
            </button>
            <button
                onClick={() => { setModalType('bindDoctor'); setShowModal(true) }}
                className={`px-5 py-3 text-sm font-medium transition-colors ${
                    activeTab === 'bind'
                        ? 'text-brand-600 border-b-2 border-brand-400'
                        : 'text-gray-400 hover:text-gray-600'
                }`}
            >
              Привязка врачей
            </button>
          </div>

          <div className="p-6">
            {activeTab === 'users' && (
                <UserTable
                    users={users}
                    loading={loading}
                    onEdit={handleEditUser}
                    onDelete={handleDeleteUser}
                    onBlock={handleBlockUser}
                />
            )}

            {activeTab === 'clinics' && (
                <div>
                  <div className="flex justify-between items-center mb-4">
                    <p className="text-sm text-gray-600">
                      {loading ? 'Загрузка...' : `${clinics.length} клиник`}
                    </p>
                    <button
                        onClick={() => { setModalType('createClinic'); setShowModal(true) }}
                        className="btn-primary flex items-center gap-1.5"
                    >
                      <svg width="13" height="13" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round">
                        <path d="M8 2v12M2 8h12"/>
                      </svg>
                      Добавить клинику
                    </button>
                  </div>
                  <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                    {clinics.map(clinic => (
                        <ClinicCard
                            key={clinic.id}
                            clinic={clinic}
                            doctorsCount={0}
                            patientsCount={0}
                            onEdit={(c) => { setModalType('editClinic'); setFormData(c); setShowModal(true) }}
                            onDelete={async (c) => {
                              if (window.confirm(`Удалить клинику "${c.name}"?`)) {
                                try {
                                  await adminApi.deleteClinic(c.id)
                                  await loadData()
                                } catch {
                                  alert('Не удалось удалить клинику')
                                }
                              }
                            }}
                        />
                    ))}
                  </div>
                </div>
            )}
          </div>
        </div>

        {showModal && (
            <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
              <div className="bg-white rounded-xl p-6 w-full max-w-md">
                <h2 className="text-base font-semibold text-gray-800 mb-4">
                  {modalType === 'bindDoctor' && 'Привязать врача к клинике'}
                  {modalType === 'createClinic' && 'Создать клинику'}
                  {modalType === 'editUser' && 'Редактировать пользователя'}
                </h2>

                {modalType === 'bindDoctor' && (
                    <div className="space-y-3">
                      <div>
                        <label className="text-xs text-gray-500 mb-1 block">ID клиники</label>
                        <input
                            className="input-field"
                            type="number"
                            value={formData.clinicId || ''}
                            onChange={(e) => setFormData({ ...formData, clinicId: e.target.value })}
                            placeholder="Например: 1"
                        />
                      </div>
                      <div>
                        <label className="text-xs text-gray-500 mb-1 block">ID врача</label>
                        <input
                            className="input-field"
                            type="number"
                            value={formData.doctorId || ''}
                            onChange={(e) => setFormData({ ...formData, doctorId: e.target.value })}
                            placeholder="Например: 2"
                        />
                      </div>
                    </div>
                )}

                {modalType === 'createClinic' && (
                    <div>
                      <label className="text-xs text-gray-500 mb-1 block">Название клиники</label>
                      <input
                          className="input-field"
                          value={formData.name || ''}
                          onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                          placeholder="Например: Городская больница №1"
                      />
                    </div>
                )}

                {modalType === 'editUser' && (
                    <div className="space-y-3">
                      <div>
                        <label className="text-xs text-gray-500 mb-1 block">Имя</label>
                        <input
                            className="input-field"
                            value={formData.name || ''}
                            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                        />
                      </div>
                      <div>
                        <label className="text-xs text-gray-500 mb-1 block">Email</label>
                        <input
                            className="input-field"
                            type="email"
                            value={formData.email || ''}
                            onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                        />
                      </div>
                    </div>
                )}

                <div className="flex gap-2 mt-6">
                  <button
                      onClick={() => {
                        if (modalType === 'bindDoctor') handleBindDoctor()
                        else if (modalType === 'createClinic') handleCreateClinic()
                        else setShowModal(false)
                      }}
                      className="btn-primary flex-1"
                  >
                    Сохранить
                  </button>
                  <button
                      onClick={() => { setShowModal(false); setFormData({}) }}
                      className="btn-secondary flex-1"
                  >
                    Отмена
                  </button>
                </div>
              </div>
            </div>
        )}
      </div>
  )
}