import { useState } from "react"
import { useNavigate } from "react-router-dom"
import { useAuth } from "../context/AuthContext"

export default function AuthPage() {
  const [role, setRole] = useState("doctor")
  const [mode, setMode] = useState("login")
  const [form, setForm] = useState({ email: "", password: "", name: "", inviteCode: "" })
  const [error, setError] = useState("")
  const [loading, setLoading] = useState(false)
  const { login } = useAuth()
  const navigate = useNavigate()

  const handleChange = (e) => {
    setForm((prev) => ({ ...prev, [e.target.name]: e.target.value }))
    setError("")
  }

  const handleSubmit = () => {
    const mockUser = {
      id: 1,
      name: role === "doctor" ? "Д-р Иванов" : "Алёна",
      role,
      token: "mock-token",
    }
    login(mockUser)
    navigate(role === "doctor" ? "/doctor" : "/patient")
  }

  const isPatient = role === "patient"

  return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-slate-100 flex items-center justify-center p-4">
        <div className="w-full max-w-md">
          <div className="flex items-center gap-2 mb-6 justify-center">
            <div className="w-8 h-8 bg-brand-400 rounded-lg flex items-center justify-center">
              <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
                <path d="M8 2v12M2 8h12" stroke="white" strokeWidth="2.5" strokeLinecap="round"/>
              </svg>
            </div>
            <span className="text-xl font-medium text-gray-800">МедБратишка</span>
          </div>
          <div className="bg-white rounded-2xl border border-gray-100 shadow-sm overflow-hidden">
            <div className="grid grid-cols-2 border-b border-gray-100">
              <button onClick={() => setRole("doctor")} className={`flex items-center justify-center gap-2 py-3.5 text-sm font-medium transition-colors ${!isPatient ? "bg-brand-50 text-brand-600 border-b-2 border-brand-400" : "text-gray-400 hover:bg-gray-50"}`}>
                <svg width="15" height="15" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5"><rect x="3" y="1" width="10" height="14" rx="1.5"/><path d="M6 5h4M6 8h4M6 11h2"/></svg>
                Врач
              </button>
              <button onClick={() => setRole("patient")} className={`flex items-center justify-center gap-2 py-3.5 text-sm font-medium transition-colors ${isPatient ? "bg-brand-50 text-brand-600 border-b-2 border-brand-400" : "text-gray-400 hover:bg-gray-50"}`}>
                <svg width="15" height="15" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5"><circle cx="8" cy="5" r="3"/><path d="M2 14c0-3.3 2.7-6 6-6s6 2.7 6 6"/></svg>
                Пациент
              </button>
            </div>
            <div className="p-7 flex flex-col gap-4">
              {mode === "register" && (
                  <div className="flex flex-col gap-1.5">
                    <label className="text-xs text-gray-500">Имя</label>
                    <input className="input-field" name="name" placeholder="Иван Иванов" value={form.name} onChange={handleChange} />
                  </div>
              )}
              <div className="flex flex-col gap-1.5">
                <label className="text-xs text-gray-500">Email</label>
                <input className="input-field" name="email" type="email" placeholder={isPatient ? "patient@mail.ru" : "doctor@clinic.ru"} value={form.email} onChange={handleChange} />
              </div>
              <div className="flex flex-col gap-1.5">
                <label className="text-xs text-gray-500">Пароль</label>
                <input className="input-field" name="password" type="password" placeholder="••••••••" value={form.password} onChange={handleChange} />
              </div>
              {isPatient && mode === "register" && (
                  <div className="flex flex-col gap-1.5">
                    <label className="text-xs text-gray-500">Код приглашения</label>
                    <input className="input-field font-mono tracking-widest uppercase" name="inviteCode" placeholder="MED-XXXXX" value={form.inviteCode} onChange={handleChange} />
                  </div>
              )}
              {error && <p className="text-xs text-red-500 bg-red-50 rounded-lg px-3 py-2">{error}</p>}
              <button onClick={handleSubmit} disabled={loading} className="btn-primary w-full py-2.5 mt-1 disabled:opacity-50">
                {loading ? "Загрузка..." : mode === "login" ? `Войти как ${isPatient ? "пациент" : "врач"}` : "Зарегистрироваться"}
              </button>
              <div className="flex items-center gap-3">
                <div className="flex-1 h-px bg-gray-100" />
                <span className="text-xs text-gray-400">{mode === "login" ? "нет аккаунта?" : "есть аккаунт?"}</span>
                <div className="flex-1 h-px bg-gray-100" />
              </div>
              <button onClick={() => setMode(mode === "login" ? "register" : "login")} className="btn-secondary w-full py-2.5 text-xs">
                {mode === "login" ? "Зарегистрироваться" : "Войти"}
              </button>
            </div>
          </div>
        </div>
      </div>
  )
}