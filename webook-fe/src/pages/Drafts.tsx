import { useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { Button, Chip, Spinner } from '@heroui/react'
import { ArrowLeft, FileText } from 'lucide-react'
import { api } from '../services/api'

// Backend ArticleVO from /articles/list: id, title, abstract, status, ctime, utime
interface DraftItem {
  id: number
  title: string
  abstract?: string
  status: number
  ctime: string
  utime: string
}

function formatTime(dateStr: string): string {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  if (isNaN(d.getTime())) return ''
  const year = d.getFullYear()
  const month = (d.getMonth() + 1).toString().padStart(2, '0')
  const day = d.getDate().toString().padStart(2, '0')
  return `${year}-${month}-${day}`
}

function getStatusInfo(status: number): { label: string; color: string } {
  switch (status) {
    case 1:
      return { label: '草稿', color: 'orange' }
    case 2:
      return { label: '已发布', color: 'green' }
    case 3:
      return { label: '已撤回', color: 'gray' }
    default:
      return { label: '草稿', color: 'orange' }
  }
}

export default function Drafts() {
  const navigate = useNavigate()

  // Backend: POST /articles/list  body: { offset, limit } (not page/page_size)
  const { data: articles, isLoading } = useQuery({
    queryKey: ['my-articles'],
    queryFn: async () => {
      const res = await api.post<DraftItem[]>('/articles/list', {
        offset: 0,
        limit: 50,
      })
      return res.data || []
    },
  })

  return (
    <div className="flex flex-col min-h-screen">
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
          <h1 className="text-base font-medium text-gray-900">草稿箱</h1>
          <button
            className="text-blue-500 text-sm font-medium"
            onClick={() => navigate('/write')}
          >
            编辑
          </button>
        </div>
      </header>

      {/* Content */}
      <div className="flex-1">
        {isLoading ? (
          <div className="flex items-center justify-center py-20">
            <Spinner size="lg" />
          </div>
        ) : articles && articles.length > 0 ? (
          <div>
            {articles.map((article) => {
              const statusInfo = getStatusInfo(article.status)
              return (
                <div
                  key={article.id}
                  className="px-4 py-3 border-b border-gray-50 active:bg-gray-50 cursor-pointer"
                  onClick={() => navigate(`/write/${article.id}`)}
                >
                  <div className="flex items-start justify-between gap-2 mb-1">
                    <h3 className="text-sm font-semibold text-gray-900 flex-1 line-clamp-2">
                      {article.title || '无标题'}
                    </h3>
                    <Chip
                      size="sm"
                      variant="soft"
                      color={
                        statusInfo.color === 'green'
                          ? 'success'
                          : statusInfo.color === 'orange'
                            ? 'warning'
                            : 'default'
                      }
                    >
                      {statusInfo.label}
                    </Chip>
                  </div>
                  {article.abstract && (
                    <p className="text-xs text-gray-500 line-clamp-2 mb-2">
                      {article.abstract}
                    </p>
                  )}
                  <div className="flex items-center gap-3 text-gray-400">
                    <span className="text-xs">
                      {formatTime(article.utime || article.ctime)}
                    </span>
                  </div>
                </div>
              )
            })}
          </div>
        ) : (
          <div className="flex flex-col items-center justify-center py-20 text-gray-400">
            <FileText className="w-12 h-12 mb-3 text-gray-200" />
            <p className="text-sm">还没有文章</p>
            <Button
              size="sm"
              variant="primary"
              className="mt-4"
              onPress={() => navigate('/write')}
            >
              开始写作
            </Button>
          </div>
        )}
      </div>
    </div>
  )
}
