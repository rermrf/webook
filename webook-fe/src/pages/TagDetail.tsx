import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  ArrowLeft,
  Hash,
  Users,
  FileText,
  Loader2,
  ChevronRight,
} from 'lucide-react'
import { api } from '../services/api'
import { formatCount } from '../utils/formatTime'

type SortTab = 'newest' | 'hottest' | 'featured'

interface TagInfo {
  id: number
  name: string
  description?: string
  follower_count?: number
}

// Backend CountBizByTag returns raw number (resp.GetCount())
// Backend CheckTagFollow returns raw boolean (resp.GetFollowed())

export default function TagDetail() {
  const { id: tagId } = useParams()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [activeSort, setActiveSort] = useState<SortTab>('newest')

  const numTagId = Number(tagId) || 0

  // Fetch tag detail
  const {
    data: tagInfo,
    isLoading: tagLoading,
  } = useQuery({
    queryKey: ['tag-detail', tagId],
    queryFn: async () => {
      const res = await api.get<TagInfo>(`/tags/detail?id=${tagId}`)
      return res.data
    },
    enabled: numTagId > 0,
  })

  // Fetch article count for tag — backend returns raw number
  const { data: bizCount } = useQuery({
    queryKey: ['tag-biz-count', tagId],
    queryFn: async () => {
      const res = await api.get<number>(
        `/tags/biz_count?biz=article&tag_id=${tagId}`
      )
      return res.data ?? 0
    },
    enabled: numTagId > 0,
  })

  // Fetch follow status — backend returns raw boolean
  const { data: followStatus } = useQuery({
    queryKey: ['tag-follow-status', tagId],
    queryFn: async () => {
      const res = await api.get<boolean>(`/tags/${tagId}/follow`)
      return res.data ?? false
    },
    enabled: numTagId > 0,
  })

  // Fetch article IDs for tag
  const {
    data: articleIds,
    isLoading: articlesLoading,
  } = useQuery({
    queryKey: ['tag-articles', tagId, activeSort],
    queryFn: async () => {
      const res = await api.get<number[]>(
        `/tags/biz_ids?biz=article&tag_id=${tagId}&sort_by=${activeSort}&offset=0&limit=20`
      )
      return res.data || []
    },
    enabled: numTagId > 0,
  })

  // Follow mutation
  const followMutation = useMutation({
    mutationFn: async () => {
      if (followStatus) {
        await api.delete(`/tags/${tagId}/follow`)
      } else {
        await api.post(`/tags/${tagId}/follow`)
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['tag-follow-status', tagId],
      })
      queryClient.invalidateQueries({
        queryKey: ['tag-detail', tagId],
      })
    },
  })

  const sortTabs: { key: SortTab; label: string }[] = [
    { key: 'newest', label: '最新' },
    { key: 'hottest', label: '最热' },
    { key: 'featured', label: '精选' },
  ]

  const followerCount = tagInfo?.follower_count ?? 0

  return (
    <div className="flex flex-col min-h-screen bg-gray-50">
      {/* Header */}
      <header className="sticky top-0 z-40 bg-white border-b border-gray-100">
        <div className="flex items-center px-4 pt-[env(safe-area-inset-top)] h-14">
          <button
            onClick={() => navigate(-1)}
            className="w-9 h-9 flex items-center justify-center -ml-2 rounded-full active:bg-gray-100"
          >
            <ArrowLeft className="w-5 h-5 text-gray-700" />
          </button>
          <h1 className="flex-1 text-center text-base font-medium text-gray-900 pr-5 truncate">
            {tagInfo ? `#${tagInfo.name}` : '#...'}
          </h1>
        </div>
      </header>

      {/* Tag Info Card */}
      {tagLoading ? (
        <div className="flex items-center justify-center py-10">
          <Loader2 className="w-6 h-6 text-gray-300 animate-spin" />
        </div>
      ) : tagInfo ? (
        <div className="bg-white px-4 py-5 mb-2">
          {/* Tag chip */}
          <div className="flex items-center justify-center mb-4">
            <span className="inline-flex items-center gap-1.5 px-4 py-2 rounded-full bg-blue-50 text-blue-600 text-base font-semibold">
              <Hash className="w-4 h-4" />
              {tagInfo.name}
            </span>
          </div>

          {/* Description */}
          {tagInfo.description && (
            <p className="text-sm text-gray-500 text-center mb-4 px-4">
              {tagInfo.description}
            </p>
          )}

          {/* Stats */}
          <div className="flex items-center justify-center gap-4 text-sm text-gray-500 mb-5">
            <span className="flex items-center gap-1">
              <FileText className="w-3.5 h-3.5" />
              {bizCount != null ? formatCount(bizCount) : '0'} 篇文章
            </span>
            <span className="text-gray-300">|</span>
            <span className="flex items-center gap-1">
              <Users className="w-3.5 h-3.5" />
              {formatCount(followerCount)} 关注
            </span>
          </div>

          {/* Follow button */}
          <div className="flex justify-center">
            <button
              onClick={() => followMutation.mutate()}
              disabled={followMutation.isPending}
              className={`px-8 h-10 rounded-full text-sm font-medium transition-colors ${
                followStatus
                  ? 'bg-gray-100 text-gray-600 active:bg-gray-200'
                  : 'bg-blue-500 text-white active:bg-blue-600'
              } disabled:opacity-50`}
            >
              {followMutation.isPending ? (
                <Loader2 className="w-4 h-4 animate-spin mx-auto" />
              ) : followStatus ? (
                '已关注'
              ) : (
                '关注话题'
              )}
            </button>
          </div>
        </div>
      ) : null}

      {/* Sort Tabs */}
      <div className="bg-white border-b border-gray-100 mb-2">
        <div className="flex">
          {sortTabs.map((tab) => (
            <button
              key={tab.key}
              onClick={() => setActiveSort(tab.key)}
              className={`flex-1 py-3 text-center text-sm font-medium relative ${
                activeSort === tab.key ? 'text-gray-900' : 'text-gray-400'
              }`}
            >
              {tab.label}
              {activeSort === tab.key && (
                <div className="absolute bottom-0 left-1/2 -translate-x-1/2 w-6 h-0.5 bg-blue-500 rounded-full" />
              )}
            </button>
          ))}
        </div>
      </div>

      {/* Article List */}
      <div className="flex-1">
        {articlesLoading ? (
          <div className="flex items-center justify-center py-20">
            <Loader2 className="w-6 h-6 text-gray-300 animate-spin" />
          </div>
        ) : articleIds && articleIds.length > 0 ? (
          <div className="bg-white">
            {articleIds.map((artId) => (
              <div
                key={artId}
                className="flex items-center justify-between px-4 py-3.5 border-b border-gray-50 cursor-pointer active:bg-gray-50"
                onClick={() => navigate(`/article/${artId}`)}
              >
                <div className="flex items-center gap-3 flex-1 min-w-0">
                  <div className="w-8 h-8 rounded-lg bg-blue-50 flex items-center justify-center shrink-0">
                    <FileText className="w-4 h-4 text-blue-400" />
                  </div>
                  <span className="text-sm text-gray-700 truncate">
                    查看文章详情
                  </span>
                </div>
                <ChevronRight className="w-4 h-4 text-gray-300 shrink-0" />
              </div>
            ))}
          </div>
        ) : (
          <div className="flex flex-col items-center justify-center py-20 text-gray-400">
            <FileText className="w-10 h-10 mb-2 text-gray-200" />
            <p className="text-sm">暂无相关文章</p>
          </div>
        )}
      </div>
    </div>
  )
}
