import { useState } from 'react'
import { Spinner } from '@heroui/react'
import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { Bell, User, Heart, MessageCircle } from 'lucide-react'
import { api } from '../services/api'
import type { Article, FeedEvent, Tag } from '../types'

type TabKey = 'following' | 'recommended'

function parseDateTime(s: string): number {
  if (!s) return 0
  const d = new Date(s)
  return isNaN(d.getTime()) ? 0 : d.getTime()
}

function formatTime(dateStr: string): string {
  const ts = parseDateTime(dateStr)
  if (!ts) return ''
  const now = Date.now()
  const diff = now - ts
  const minutes = Math.floor(diff / 60000)
  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes}分钟前`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}小时前`
  const days = Math.floor(hours / 24)
  if (days < 30) return `${days}天前`
  const date = new Date(ts)
  return `${date.getMonth() + 1}月${date.getDate()}日`
}

// Backend ArticleVO: flat structure with snake_case fields
function ArticleCard({ article }: { article: Article }) {
  const navigate = useNavigate()

  return (
    <div
      className="px-4 py-3 border-b border-gray-50 active:bg-gray-50 cursor-pointer"
      onClick={() => navigate(`/article/${article.id}`)}
    >
      {/* Author info */}
      <div className="flex items-center gap-2 mb-2">
        <div className="w-8 h-8 rounded-full bg-gradient-to-br from-blue-400 to-purple-400 flex items-center justify-center text-white text-xs font-medium shrink-0">
          {(article.author_name || '匿名').charAt(0)}
        </div>
        <span className="text-sm text-gray-700 font-medium">
          {article.author_name || '匿名用户'}
        </span>
        <span className="text-xs text-gray-400">
          {formatTime(article.ctime)}
        </span>
      </div>

      {/* Title */}
      <h3 className="text-base font-semibold text-gray-900 mb-1 line-clamp-2">
        {article.title}
      </h3>

      {/* Tags */}
      {(article as Article & { tags?: Tag[] }).tags && (article as Article & { tags?: Tag[] }).tags!.length > 0 && (
        <div className="flex flex-wrap gap-1 mb-2">
          {(article as Article & { tags?: Tag[] }).tags!.map((t) => (
            <span key={t.id} className="text-xs text-blue-500 bg-blue-50 px-2 py-0.5 rounded-full">#{t.name}</span>
          ))}
        </div>
      )}

      {/* Abstract */}
      {article.abstract && (
        <p className="text-sm text-gray-500 line-clamp-2 mb-2">
          {article.abstract}
        </p>
      )}

      {/* Stats — flat fields: like_cnt, collect_cnt from backend ArticleVO */}
      <div className="flex items-center gap-4 text-gray-400">
        <div className="flex items-center gap-1">
          <Heart className="w-3.5 h-3.5" />
          <span className="text-xs">{(article as Article & { like_cnt?: number }).like_cnt || 0}</span>
        </div>
        <div className="flex items-center gap-1">
          <MessageCircle className="w-3.5 h-3.5" />
          <span className="text-xs">{(article as Article & { collect_cnt?: number }).collect_cnt || 0}</span>
        </div>
      </div>
    </div>
  )
}

// Feed articles use FeedArticleVO shape from backend
function FeedArticleCard({ event }: { event: FeedEvent }) {
  const navigate = useNavigate()
  const article = event.article
  if (!article) return null

  return (
    <div
      className="px-4 py-3 border-b border-gray-50 active:bg-gray-50 cursor-pointer"
      onClick={() => navigate(`/article/${article.id}`)}
    >
      <div className="flex items-center gap-2 mb-2">
        <div className="w-8 h-8 rounded-full bg-gradient-to-br from-blue-400 to-purple-400 flex items-center justify-center text-white text-xs font-medium shrink-0">
          {(article.author_name || '匿名').charAt(0)}
        </div>
        <span className="text-sm text-gray-700 font-medium">
          {article.author_name || '匿名用户'}
        </span>
        <span className="text-xs text-gray-400">
          {formatTime(article.ctime)}
        </span>
      </div>
      <h3 className="text-base font-semibold text-gray-900 mb-1 line-clamp-2">
        {article.title}
      </h3>
      {/* Tags */}
      {article.tags && article.tags.length > 0 && (
        <div className="flex flex-wrap gap-1 mb-2">
          {article.tags.map((tag, idx) => (
            <span key={idx} className="text-xs text-blue-500 bg-blue-50 px-2 py-0.5 rounded-full">#{tag}</span>
          ))}
        </div>
      )}
      {article.abstract && (
        <p className="text-sm text-gray-500 line-clamp-2 mb-2">
          {article.abstract}
        </p>
      )}
      <div className="flex items-center gap-4 text-gray-400">
        <div className="flex items-center gap-1">
          <Heart className="w-3.5 h-3.5" />
          <span className="text-xs">{article.like_cnt || 0}</span>
        </div>
        <div className="flex items-center gap-1">
          <MessageCircle className="w-3.5 h-3.5" />
          <span className="text-xs">{article.comment_cnt || 0}</span>
        </div>
      </div>
    </div>
  )
}

export default function Home() {
  const [activeTab, setActiveTab] = useState<TabKey>('recommended')
  const navigate = useNavigate()

  // Backend: GET /feed  returns FeedEventVO[] with nested article
  const { data: feedEvents, isLoading: feedLoading } = useQuery({
    queryKey: ['feed'],
    queryFn: async () => {
      const res = await api.get<FeedEvent[]>('/feed')
      return res.data || []
    },
    enabled: activeTab === 'following',
  })

  // Backend: GET /articles/pub/articles  returns ArticleVO[] (flat, snake_case)
  const { data: recommendedArticles, isLoading: recommendedLoading } = useQuery({
    queryKey: ['articles', 'recommended'],
    queryFn: async () => {
      const res = await api.get<Article[]>('/articles/pub/articles')
      return res.data || []
    },
    enabled: activeTab === 'recommended',
  })

  const isLoading = activeTab === 'following' ? feedLoading : recommendedLoading

  return (
    <div className="flex flex-col min-h-screen">
      {/* Header */}
      <header className="sticky top-0 z-40 bg-white">
        <div className="flex items-center justify-between px-4 pt-[env(safe-area-inset-top)] h-14">
          <h1 className="text-xl font-bold text-blue-500">WeBook</h1>
          <div className="flex items-center gap-3">
            <button
              onClick={() => navigate('/messages')}
              className="p-1.5 text-gray-500"
            >
              <Bell className="w-5 h-5" />
            </button>
            <button
              onClick={() => navigate('/profile')}
              className="p-1.5 text-gray-500"
            >
              <User className="w-5 h-5" />
            </button>
          </div>
        </div>

        {/* Tabs */}
        <div className="flex px-4 border-b border-gray-100">
          <button
            onClick={() => setActiveTab('following')}
            className={`flex-1 py-3 text-center text-sm font-medium relative ${
              activeTab === 'following' ? 'text-gray-900' : 'text-gray-400'
            }`}
          >
            关注
            {activeTab === 'following' && (
              <div className="absolute bottom-0 left-1/2 -translate-x-1/2 w-6 h-0.5 bg-blue-500 rounded-full" />
            )}
          </button>
          <button
            onClick={() => setActiveTab('recommended')}
            className={`flex-1 py-3 text-center text-sm font-medium relative ${
              activeTab === 'recommended' ? 'text-gray-900' : 'text-gray-400'
            }`}
          >
            推荐
            {activeTab === 'recommended' && (
              <div className="absolute bottom-0 left-1/2 -translate-x-1/2 w-6 h-0.5 bg-blue-500 rounded-full" />
            )}
          </button>
        </div>
      </header>

      {/* Content */}
      <div className="flex-1">
        {isLoading ? (
          <div className="flex items-center justify-center py-20">
            <Spinner size="lg" />
          </div>
        ) : activeTab === 'following' ? (
          feedEvents && feedEvents.length > 0 ? (
            <div>
              {feedEvents.map((event) => (
                <FeedArticleCard key={event.id} event={event} />
              ))}
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center py-20 text-gray-400">
              <p className="text-sm">关注一些作者，这里会显示他们的最新文章</p>
            </div>
          )
        ) : recommendedArticles && recommendedArticles.length > 0 ? (
          <div>
            {recommendedArticles.map((article) => (
              <ArticleCard key={article.id} article={article} />
            ))}
          </div>
        ) : (
          <div className="flex flex-col items-center justify-center py-20 text-gray-400">
            <p className="text-sm">暂无推荐内容</p>
          </div>
        )}
      </div>
    </div>
  )
}
