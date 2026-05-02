import { createContext, useContext, useState, useEffect } from 'react'

const AuthContext = createContext(null)

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null)
  const [isLoading, setIsLoading] = useState(true) // true пока не проверили localStorage
  const [authError, setAuthError] = useState(null)

  // Автоматически восстанавливаем сессию при монтировании
  useEffect(() => {
    const stored = localStorage.getItem('medcare_user')
    if (stored) {
      try {
        setUser(JSON.parse(stored))
      } catch {
        localStorage.removeItem('medcare_user')
      }
    }
    setIsLoading(false)
  }, [])

  const login = (userData) => {
    setUser(userData)
    setAuthError(null)
    localStorage.setItem('medcare_user', JSON.stringify(userData))
  }

  const logout = () => {
    setUser(null)
    setAuthError(null)
    localStorage.removeItem('medcare_user')
  }

  const isAuthenticated = !!user && !isLoading

  return (
    <AuthContext.Provider value={{
      user,
      isLoading,
      isAuthenticated,
      authError,
      setAuthError,
      login,
      logout,
    }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  return useContext(AuthContext)
}
