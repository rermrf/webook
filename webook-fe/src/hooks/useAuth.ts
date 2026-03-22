import { useState, useCallback } from 'react'
import { api } from '../services/api'

export function useAuth() {
  const [isAuthenticated, setIsAuthenticated] = useState(
    () => !!localStorage.getItem('accessToken')
  )

  const login = useCallback(async (phone: string, code: string) => {
    const res = await api.postWithHeaders('/users/login_sms', { phone, code })
    if (res.headers.accessToken) {
      localStorage.setItem('accessToken', res.headers.accessToken)
      localStorage.setItem('refreshToken', res.headers.refreshToken || '')
      setIsAuthenticated(true)
    }
    return res
  }, [])

  const logout = useCallback(() => {
    localStorage.removeItem('accessToken')
    localStorage.removeItem('refreshToken')
    setIsAuthenticated(false)
  }, [])

  return { isAuthenticated, login, logout }
}
