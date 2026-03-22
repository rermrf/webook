import { useState, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Avatar, Button, Spinner } from '@heroui/react'
import { ArrowLeft, Camera } from 'lucide-react'
import { api } from '../services/api'
import type { User } from '../types'

export default function EditProfile() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const fileInputRef = useRef<HTMLInputElement>(null)

  const [nickname, setNickname] = useState('')
  const [aboutMe, setAboutMe] = useState('')
  const [birthday, setBirthday] = useState('')
  const [avatarUrl, setAvatarUrl] = useState('')
  const [saveMsg, setSaveMsg] = useState('')
  const [loaded, setLoaded] = useState(false)

  // Load profile
  const { isLoading } = useQuery({
    queryKey: ['profile', 'me'],
    queryFn: async () => {
      const res = await api.get<User>('/users/profile')
      if (res.data && !loaded) {
        setNickname(res.data.nickname || '')
        setAboutMe(res.data.aboutMe || '')
        setBirthday(res.data.birthday || '')
        setAvatarUrl(res.data.avatar || '')
        setLoaded(true)
      }
      return res.data
    },
  })

  const { data: profile } = useQuery({
    queryKey: ['profile', 'me'],
    queryFn: async () => {
      const res = await api.get<User>('/users/profile')
      return res.data
    },
  })

  // Save mutation
  const saveMutation = useMutation({
    mutationFn: async () => {
      await api.post('/users/edit', {
        nickname: nickname.trim(),
        aboutMe: aboutMe.trim(),
        birthday: birthday.trim(),
      })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['profile'] })
      setSaveMsg('保存成功')
      setTimeout(() => setSaveMsg(''), 2000)
    },
    onError: () => {
      setSaveMsg('保存失败')
      setTimeout(() => setSaveMsg(''), 2000)
    },
  })

  // Avatar upload mutation
  const uploadMutation = useMutation({
    mutationFn: async (file: File) => {
      const formData = new FormData()
      formData.append('avatar', file)
      const res = await api.post<{ url: string }>('/upload/avatar', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
      return res.data?.url
    },
    onSuccess: (url) => {
      if (url) {
        setAvatarUrl(url)
        queryClient.invalidateQueries({ queryKey: ['profile'] })
      }
    },
  })

  const handleAvatarClick = () => {
    fileInputRef.current?.click()
  }

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) {
      // Preview immediately
      const reader = new FileReader()
      reader.onloadend = () => {
        setAvatarUrl(reader.result as string)
      }
      reader.readAsDataURL(file)
      uploadMutation.mutate(file)
    }
  }

  if (isLoading) {
    return (
      <div className="flex flex-col min-h-screen">
        <header className="sticky top-0 z-40 bg-white border-b border-gray-100">
          <div className="flex items-center justify-between px-4 pt-[env(safe-area-inset-top)] h-14">
            <Button
              isIconOnly
              variant="ghost"
              onPress={() => navigate(-1)}
              size="sm"
            >
              <ArrowLeft className="w-5 h-5" />
            </Button>
            <h1 className="text-base font-medium text-gray-900">编辑资料</h1>
            <div className="w-9" />
          </div>
        </header>
        <div className="flex-1 flex items-center justify-center">
          <Spinner size="lg" />
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col min-h-screen bg-gray-50">
      {/* Header */}
      <header className="sticky top-0 z-40 bg-white border-b border-gray-100">
        <div className="flex items-center justify-between px-4 pt-[env(safe-area-inset-top)] h-14">
          <Button
            isIconOnly
            variant="ghost"
            onPress={() => navigate(-1)}
            size="sm"
          >
            <ArrowLeft className="w-5 h-5" />
          </Button>
          <h1 className="text-base font-medium text-gray-900">编辑资料</h1>
          <Button
            size="sm"
            variant="primary"
            onPress={() => saveMutation.mutate()}
            isDisabled={saveMutation.isPending}
          >
            保存
          </Button>
        </div>
        {saveMsg && (
          <div
            className={`text-center py-1 text-xs ${
              saveMsg.includes('成功')
                ? 'text-green-600 bg-green-50'
                : 'text-red-600 bg-red-50'
            }`}
          >
            {saveMsg}
          </div>
        )}
      </header>

      {/* Avatar Section */}
      <div className="bg-white mt-2 py-6 flex flex-col items-center">
        <div className="relative" onClick={handleAvatarClick}>
          <Avatar size="lg" className="w-20 h-20 text-2xl cursor-pointer">
            {avatarUrl && <Avatar.Image src={avatarUrl} />}
            <Avatar.Fallback>
              {(nickname || '用户').charAt(0)}
            </Avatar.Fallback>
          </Avatar>
          <div className="absolute bottom-0 right-0 w-7 h-7 rounded-full bg-blue-500 text-white flex items-center justify-center shadow-md">
            <Camera className="w-3.5 h-3.5" />
          </div>
        </div>
        <button
          className="text-sm text-blue-500 mt-2"
          onClick={handleAvatarClick}
        >
          更换头像
        </button>
        <input
          ref={fileInputRef}
          type="file"
          accept="image/*"
          className="hidden"
          onChange={handleFileChange}
        />
      </div>

      {/* Form Fields */}
      <div className="bg-white mt-2">
        {/* Nickname */}
        <div className="flex items-center px-4 py-3 border-b border-gray-50">
          <label className="w-20 text-sm text-gray-500 shrink-0">昵称</label>
          <input
            type="text"
            value={nickname}
            onChange={(e) => setNickname(e.target.value)}
            placeholder="请输入昵称"
            className="flex-1 text-sm text-gray-900 text-right outline-none"
            maxLength={20}
          />
        </div>

        {/* About Me */}
        <div className="flex items-start px-4 py-3 border-b border-gray-50">
          <label className="w-20 text-sm text-gray-500 shrink-0 pt-0.5">
            个人简介
          </label>
          <textarea
            value={aboutMe}
            onChange={(e) => setAboutMe(e.target.value)}
            placeholder="介绍一下自己吧"
            className="flex-1 text-sm text-gray-900 text-right outline-none resize-none min-h-[60px]"
            maxLength={200}
          />
        </div>

        {/* Birthday */}
        <div className="flex items-center px-4 py-3 border-b border-gray-50">
          <label className="w-20 text-sm text-gray-500 shrink-0">生日</label>
          <input
            type="date"
            value={birthday}
            onChange={(e) => setBirthday(e.target.value)}
            className="flex-1 text-sm text-gray-900 text-right outline-none"
          />
        </div>

        {/* Email (readonly) */}
        <div className="flex items-center px-4 py-3 border-b border-gray-50">
          <label className="w-20 text-sm text-gray-500 shrink-0">邮箱</label>
          <span className="flex-1 text-sm text-gray-400 text-right">
            {profile?.email || '未绑定'}
          </span>
        </div>

        {/* Phone (readonly) */}
        <div className="flex items-center px-4 py-3">
          <label className="w-20 text-sm text-gray-500 shrink-0">手机号</label>
          <span className="flex-1 text-sm text-gray-400 text-right">
            {profile?.phone
              ? profile.phone.replace(/(\d{3})\d{4}(\d{4})/, '$1****$2')
              : '未绑定'}
          </span>
        </div>
      </div>
    </div>
  )
}
