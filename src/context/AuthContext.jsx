import { createContext, useContext, useState } from 'react'

const AuthContext = createContext(null)

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null)
  // user: { id, name, role: 'doctor' | 'patient', token }

  const login = (userData) => {
    setUser(userData)
    localStorage.setItem('medcare_user', JSON.stringify(userData))
  }

  const logout = () => {
    setUser(null)
    localStorage.removeItem('medcare_user')
  }

  const loadFromStorage = () => {
    const stored = localStorage.getItem('medcare_user')
    if (stored) setUser(JSON.parse(stored))
  }

  return (
    <AuthContext.Provider value={{ user, login, logout, loadFromStorage }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  return useContext(AuthContext)
}
