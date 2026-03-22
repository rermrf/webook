import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation } from '@tanstack/react-query'
import { Button, Spinner } from '@heroui/react'
import {
  X,
  Heading,
  Bold,
  Italic,
  List,
  Image,
  Link,
  Code,
  Quote,
} from 'lucide-react'
import { api } from '../services/api'

interface ArticleData {
  id: number
  title: string
  content: string
  abstract?: string
  status: string
  ctime: number
  utime: number
}

export default function ArticleEditor() {
  const { id } = useParams()
  const navigate = useNavigate()
  const [title, setTitle] = useState('')
  const [content, setContent] = useState('')
  const [articleId, setArticleId] = useState<number | undefined>(
    id ? Number(id) : undefined
  )
  const [saveMsg, setSaveMsg] = useState('')

  // Load existing article for editing
  const { isLoading } = useQuery({
    queryKey: ['article-edit', id],
    queryFn: async () => {
      const res = await api.get<ArticleData>(`/articles/detail/${id}`)
      if (res.data) {
        setTitle(res.data.title)
        setContent(res.data.content)
        setArticleId(res.data.id)
      }
      return res.data
    },
    enabled: !!id,
  })

  // Save draft mutation
  const saveDraftMutation = useMutation({
    mutationFn: async () => {
      const res = await api.post<{ id: number }>('/articles/edit', {
        id: articleId,
        title: title.trim(),
        content: content.trim(),
      })
      return res.data
    },
    onSuccess: (data) => {
      if (data?.id) setArticleId(data.id)
      setSaveMsg('草稿已保存')
      setTimeout(() => setSaveMsg(''), 2000)
    },
    onError: () => {
      setSaveMsg('保存失败')
      setTimeout(() => setSaveMsg(''), 2000)
    },
  })

  // Publish mutation
  const publishMutation = useMutation({
    mutationFn: async () => {
      const res = await api.post<{ id: number }>('/articles/publish', {
        id: articleId,
        title: title.trim(),
        content: content.trim(),
      })
      return res.data
    },
    onSuccess: () => {
      navigate(-1)
    },
    onError: () => {
      setSaveMsg('发布失败')
      setTimeout(() => setSaveMsg(''), 2000)
    },
  })

  const canSubmit = title.trim().length > 0 && content.trim().length > 0

  // Auto-save draft every 30 seconds
  useEffect(() => {
    if (!title.trim() && !content.trim()) return
    const timer = setInterval(() => {
      if (title.trim() || content.trim()) {
        saveDraftMutation.mutate()
      }
    }, 30000)
    return () => clearInterval(timer)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [title, content])

  if (id && isLoading) {
    return (
      <div className="flex flex-col min-h-screen">
        <header className="sticky top-0 z-40 bg-white border-b border-gray-100">
          <div className="flex items-center px-4 pt-[env(safe-area-inset-top)] h-14">
            <Button
              isIconOnly
              variant="ghost"
              onPress={() => navigate(-1)}
              size="sm"
            >
              <X className="w-5 h-5" />
            </Button>
          </div>
        </header>
        <div className="flex-1 flex items-center justify-center">
          <Spinner size="lg" />
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col h-screen">
      {/* Header */}
      <header className="sticky top-0 z-40 bg-white border-b border-gray-100 shrink-0">
        <div className="flex items-center justify-between px-4 pt-[env(safe-area-inset-top)] h-14">
          <Button
            isIconOnly
            variant="ghost"
            onPress={() => navigate(-1)}
            size="sm"
          >
            <X className="w-5 h-5" />
          </Button>
          <h1 className="text-base font-medium text-gray-900">写文章</h1>
          <div className="flex items-center gap-2">
            <Button
              size="sm"
              variant="ghost"
              onPress={() => saveDraftMutation.mutate()}
              isDisabled={saveDraftMutation.isPending || (!title.trim() && !content.trim())}
            >
              存草稿
            </Button>
            <Button
              size="sm"
              variant="primary"
              onPress={() => publishMutation.mutate()}
              isDisabled={!canSubmit || publishMutation.isPending}
            >
              发布
            </Button>
          </div>
        </div>
        {/* Save message */}
        {saveMsg && (
          <div className="text-center py-1 text-xs text-green-600 bg-green-50">
            {saveMsg}
          </div>
        )}
      </header>

      {/* Editor */}
      <div className="flex-1 overflow-y-auto px-4 py-4">
        {/* Title Input */}
        <input
          type="text"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="在这里输入标题..."
          className="w-full text-xl font-bold text-gray-900 placeholder-gray-300 outline-none border-none mb-4 leading-tight"
          maxLength={100}
        />

        {/* Content Textarea */}
        <textarea
          value={content}
          onChange={(e) => setContent(e.target.value)}
          placeholder="开始写作..."
          className="w-full flex-1 min-h-[400px] text-sm text-gray-700 placeholder-gray-300 outline-none border-none resize-none leading-relaxed"
        />
        <p className="text-sm text-gray-300 px-4">支持 Markdown 语法，可以使用 **加粗**、*斜体*、# 标题等格式。</p>
      </div>

      {/* Bottom Toolbar */}
      <div className="shrink-0 bg-white border-t border-gray-100">
        <div className="flex items-center gap-1 px-4 py-2 pb-[max(8px,env(safe-area-inset-bottom))] overflow-x-auto">
          {[
            { icon: <Heading className="w-5 h-5" />, label: '标题' },
            { icon: <Bold className="w-5 h-5" />, label: '加粗' },
            { icon: <Italic className="w-5 h-5" />, label: '斜体' },
            { icon: <List className="w-5 h-5" />, label: '列表' },
            { icon: <Image className="w-5 h-5" />, label: '图片' },
            { icon: <Link className="w-5 h-5" />, label: '链接' },
            { icon: <Code className="w-5 h-5" />, label: '代码' },
            { icon: <Quote className="w-5 h-5" />, label: '引用' },
          ].map((item) => (
            <button
              key={item.label}
              className="p-2 text-gray-400 hover:text-gray-600 active:bg-gray-100 rounded-lg shrink-0"
              title={item.label}
            >
              {item.icon}
            </button>
          ))}
        </div>
      </div>
    </div>
  )
}
