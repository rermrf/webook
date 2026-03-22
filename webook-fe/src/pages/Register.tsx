import { useState, useEffect, useCallback } from 'react'
import { TextField, Input, Button, Label, Link } from '@heroui/react'
import { useNavigate } from 'react-router-dom'
import { api } from '../services/api'

export default function Register() {
  const navigate = useNavigate()
  const [phone, setPhone] = useState('')
  const [code, setCode] = useState('')
  const [countdown, setCountdown] = useState(0)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState(false)

  useEffect(() => {
    if (countdown <= 0) return
    const timer = setTimeout(() => setCountdown((c) => c - 1), 1000)
    return () => clearTimeout(timer)
  }, [countdown])

  const sendCode = useCallback(async () => {
    if (!phone || countdown > 0) return
    try {
      await api.post('/code/send', { biz: 'register', phone })
      setCountdown(60)
      setError('')
    } catch {
      setError('验证码发送失败，请稍后重试')
    }
  }, [phone, countdown])

  const handleRegister = useCallback(async () => {
    if (!phone || !code) return
    setLoading(true)
    setError('')
    try {
      await api.post('/users/register', { phone, code })
      setSuccess(true)
      setTimeout(() => navigate('/login', { replace: true }), 1500)
    } catch {
      setError('注册失败，请稍后重试')
    } finally {
      setLoading(false)
    }
  }, [phone, code, navigate])

  return (
    <div className="mx-auto max-w-[430px] min-h-screen bg-white flex flex-col">
      <div className="flex-1 px-8 pt-20 pb-8 flex flex-col">
        {/* Logo */}
        <div className="text-center mb-12">
          <h1 className="text-3xl font-bold text-blue-500">WeBook</h1>
          <p className="text-gray-400 text-sm mt-2">创建你的账号</p>
        </div>

        {/* Form */}
        <div className="space-y-4">
          <TextField fullWidth>
            <Label>手机号</Label>
            <Input
              placeholder="请输入手机号"
              value={phone}
              onChange={(e) => setPhone(e.target.value)}
              type="tel"
              maxLength={11}
            />
          </TextField>

          <div className="flex gap-3 items-end">
            <TextField fullWidth className="flex-1">
              <Label>验证码</Label>
              <Input
                placeholder="请输入验证码"
                value={code}
                onChange={(e) => setCode(e.target.value)}
                maxLength={6}
              />
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

          {success && (
            <p className="text-green-500 text-sm text-center">注册成功，正在跳转登录...</p>
          )}

          <Button
            variant="primary"
            size="lg"
            fullWidth
            onPress={handleRegister}
            isDisabled={!phone || !code || loading}
            className="mt-4"
          >
            {loading ? '注册中...' : '注册'}
          </Button>
        </div>

        {/* Login link */}
        <div className="mt-auto pt-8 text-center">
          <span className="text-gray-400 text-sm">已有账号？</span>
          <Link href="/login" className="text-blue-500 text-sm ml-1">
            立即登录
          </Link>
        </div>
      </div>
    </div>
  )
}
