import { useState, useEffect, useCallback } from 'react'
import { TextField, Input, Button, Label, Link } from '@heroui/react'
import { useNavigate } from 'react-router-dom'
import { Mail, MessageSquare, Smartphone, Lock } from 'lucide-react'
import { useAuth } from '../hooks/useAuth'
import { api } from '../services/api'

export default function Login() {
  const navigate = useNavigate()
  const { login, isAuthenticated } = useAuth()
  const [phone, setPhone] = useState('')
  const [code, setCode] = useState('')
  const [countdown, setCountdown] = useState(0)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    if (isAuthenticated) {
      navigate('/', { replace: true })
    }
  }, [isAuthenticated, navigate])

  useEffect(() => {
    if (countdown <= 0) return
    const timer = setTimeout(() => setCountdown((c) => c - 1), 1000)
    return () => clearTimeout(timer)
  }, [countdown])

  // Backend: POST /users/login_sms/code/send  body: { phone }
  const sendCode = useCallback(async () => {
    if (!phone || countdown > 0) return
    try {
      await api.post('/users/login_sms/code/send', { phone })
      setCountdown(60)
      setError('')
    } catch {
      setError('验证码发送失败，请稍后重试')
    }
  }, [phone, countdown])

  // Backend: POST /users/login_sms  body: { phone, code }
  const handleLogin = useCallback(async () => {
    if (!phone || !code) return
    setLoading(true)
    setError('')
    try {
      await login(phone, code)
      navigate('/', { replace: true })
    } catch {
      setError('登录失败，请检查手机号和验证码')
    } finally {
      setLoading(false)
    }
  }, [phone, code, login, navigate])

  return (
    <div className="mx-auto max-w-[430px] min-h-screen bg-white flex flex-col">
      <div className="flex-1 px-8 pt-20 pb-8 flex flex-col">
        {/* Logo */}
        <div className="text-center mb-12">
          <div className="w-16 h-16 bg-blue-500 rounded-full flex items-center justify-center mx-auto mb-3">
            <span className="text-white text-2xl font-bold">W</span>
          </div>
          <h1 className="text-3xl font-bold text-blue-500">WeBook</h1>
          <p className="text-gray-400 text-sm mt-2">发现好内容，分享你的故事</p>
        </div>

        {/* Form */}
        <div className="space-y-4">
          <TextField fullWidth>
            <Label>手机号</Label>
            <div className="relative">
              <Smartphone className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
              <Input
                placeholder="请输入手机号"
                value={phone}
                onChange={(e) => setPhone(e.target.value)}
                type="tel"
                maxLength={11}
                className="pl-10"
              />
            </div>
          </TextField>

          <div className="flex gap-3 items-end">
            <TextField fullWidth className="flex-1">
              <Label>验证码</Label>
              <div className="relative">
                <Lock className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
                <Input
                  placeholder="请输入验证码"
                  value={code}
                  onChange={(e) => setCode(e.target.value)}
                  maxLength={6}
                  className="pl-10"
                />
              </div>
            </TextField>
            <Button
              variant="secondary"
              onPress={sendCode}
              isDisabled={!phone || countdown > 0}
              className="shrink-0 h-10"
            >
              {countdown > 0 ? `${countdown}s` : '获取验证码'}
            </Button>
          </div>

          {error && (
            <p className="text-red-500 text-sm text-center">{error}</p>
          )}

          <Button
            variant="primary"
            size="lg"
            fullWidth
            onPress={handleLogin}
            isDisabled={!phone || !code || loading}
            className="mt-4"
          >
            {loading ? '登录中...' : '登录'}
          </Button>
        </div>

        {/* Other login methods */}
        <div className="mt-10">
          <div className="flex items-center gap-3 mb-6">
            <div className="flex-1 h-px bg-gray-200" />
            <span className="text-gray-400 text-xs">其他登录方式</span>
            <div className="flex-1 h-px bg-gray-200" />
          </div>

          <div className="flex justify-center gap-8">
            <button className="w-12 h-12 rounded-full bg-green-50 flex items-center justify-center">
              <MessageSquare className="w-6 h-6 text-green-500" />
            </button>
            <button className="w-12 h-12 rounded-full bg-blue-50 flex items-center justify-center">
              <Mail className="w-6 h-6 text-blue-500" />
            </button>
          </div>
        </div>

        {/* Register link */}
        <div className="mt-auto pt-8 text-center">
          <span className="text-gray-400 text-sm">还没有账号？</span>
          <Link href="/register" className="text-blue-500 text-sm ml-1">
            立即注册
          </Link>
        </div>

        {/* Agreement footer */}
        <p className="text-xs text-gray-400 text-center mt-4">登录即表示同意<span className="text-blue-500">《用户协议》</span>和<span className="text-blue-500">《隐私政策》</span></p>
      </div>
    </div>
  )
}
