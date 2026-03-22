import { useState, useMemo } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  ArrowLeft,
  Clock,
  Trash2,
  Loader2,
} from 'lucide-react'
import { api } from '../services/api'
import { formatTime, formatDate } from '../utils/formatTime'

interface HistoryItem {
  id: number
  biz: string
  biz_id: number
  biz_title: string
  author_name?: string
  ctime: number
  utime: number
}

interface HistoryResponse {
  items: HistoryItem[]
  has_more: boolean
}

interface GroupedHistory {
  date: string
  items: HistoryItem[]
}

export default function BrowseHistory() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [showClearConfirm, setShowClearConfirm] = useState(false)

  // Fetch history
  const {
    data: historyData,
    isLoading,
  } = useQuery({
    queryKey: ['browse-history'],
    queryFn: async () => {
      const res = await api.get<HistoryResponse>(
        '/history/list?cursor=0&limit=50'
      )
      return res.data
    },
  })

  // Clear history mutation
  const clearMutation = useMutation({
    mutationFn: async () => {
      await api.delete('/history')
    },
    onSuccess: () => {
      setShowClearConfirm(false)
      queryClient.invalidateQueries({ queryKey: ['browse-history'] })
    },
  })

  // Group items by date
  const grouped: GroupedHistory[] = useMemo(() => {
    if (!historyData?.items?.length) return []

    const groups: Record<string, HistoryItem[]> = {}
    for (const item of historyData.items) {
      const dateKey = formatDate(item.ctime)
      if (!groups[dateKey]) groups[dateKey] = []
      groups[dateKey].push(item)
    }

    return Object.entries(groups).map(([date, items]) => ({
      date,
      items,
    }))
  }, [historyData])

  const hasItems = grouped.length > 0

  return (
    <div className="flex flex-col min-h-screen bg-gray-50">
      {/* Header */}
      <header className="sticky top-0 z-40 bg-white border-b border-gray-100">
        <div className="flex items-center justify-between px-4 pt-[env(safe-area-inset-top)] h-14">
          <button
            onClick={() => navigate(-1)}
            className="w-9 h-9 flex items-center justify-center -ml-2 rounded-full active:bg-gray-100"
          >
            <ArrowLeft className="w-5 h-5 text-gray-700" />
          </button>
          <h1 className="text-base font-medium text-gray-900">浏览历史</h1>
          {hasItems ? (
            <button
              onClick={() => setShowClearConfirm(true)}
              className="text-sm text-red-500 font-medium active:text-red-600"
            >
              清空
            </button>
          ) : (
            <div className="w-9" />
          )}
        </div>
      </header>

      {/* Content */}
      <div className="flex-1">
        {isLoading ? (
          <div className="flex items-center justify-center py-20">
            <Loader2 className="w-6 h-6 text-gray-300 animate-spin" />
          </div>
        ) : hasItems ? (
          <div>
            {grouped.map((group) => (
              <div key={group.date}>
                {/* Date header */}
                <div className="sticky top-14 z-30 px-4 py-2 bg-gray-50">
                  <span className="text-xs font-medium text-gray-500">
                    {group.date}
                  </span>
                </div>

                {/* Items */}
                <div className="bg-white">
                  {group.items.map((item) => (
                    <div
                      key={item.id}
                      className="flex items-start gap-3 px-4 py-3 border-b border-gray-50 cursor-pointer active:bg-gray-50"
                      onClick={() =>
                        navigate(`/article/${item.biz_id}`)
                      }
                    >
                      <div className="flex-1 min-w-0">
                        <h3 className="text-sm font-medium text-gray-900 line-clamp-2 mb-1">
                          {item.biz_title || `文章 #${item.biz_id}`}
                        </h3>
                        <div className="flex items-center gap-2 text-xs text-gray-400">
                          {item.author_name && (
                            <span className="truncate max-w-[120px]">
                              {item.author_name}
                            </span>
                          )}
                          <span>{formatTime(item.ctime)}</span>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="flex flex-col items-center justify-center py-20 text-gray-400">
            <Clock className="w-10 h-10 mb-2 text-gray-200" />
            <p className="text-sm">暂无浏览记录</p>
          </div>
        )}
      </div>

      {/* Clear confirmation modal */}
      {showClearConfirm && (
        <div className="fixed inset-0 z-[100] flex items-end justify-center">
          {/* Backdrop */}
          <div
            className="absolute inset-0 bg-black/40"
            onClick={() => setShowClearConfirm(false)}
          />
          {/* Dialog */}
          <div className="relative w-full max-w-[430px] bg-white rounded-t-2xl p-6 pb-[max(24px,env(safe-area-inset-bottom))] animate-[slideUp_0.2s_ease-out]">
            <div className="flex flex-col items-center mb-6">
              <div className="w-12 h-12 rounded-full bg-red-50 flex items-center justify-center mb-3">
                <Trash2 className="w-6 h-6 text-red-500" />
              </div>
              <h3 className="text-base font-semibold text-gray-900 mb-1">
                确认清空浏览历史？
              </h3>
              <p className="text-sm text-gray-500 text-center">
                清空后将无法恢复，确定要继续吗？
              </p>
            </div>
            <div className="flex gap-3">
              <button
                onClick={() => setShowClearConfirm(false)}
                className="flex-1 h-11 rounded-full border border-gray-200 text-sm font-medium text-gray-700 active:bg-gray-50"
              >
                取消
              </button>
              <button
                onClick={() => clearMutation.mutate()}
                disabled={clearMutation.isPending}
                className="flex-1 h-11 rounded-full bg-red-500 text-white text-sm font-medium active:bg-red-600 disabled:opacity-40 flex items-center justify-center"
              >
                {clearMutation.isPending ? (
                  <Loader2 className="w-4 h-4 animate-spin" />
                ) : (
                  '确认清空'
                )}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
