import { useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import {
  ArrowLeft,
  Eye,
  Heart,
  Flame,
  Loader2,
} from 'lucide-react'
import { api } from '../services/api'
import { formatCount } from '../utils/formatTime'

interface HotArticle {
  id: number
  title: string
  abstract?: string
  author_id?: number
  author_name?: string
  like_cnt?: number
  read_cnt?: number
  authorName?: string
  interactive?: {
    readCnt: number
    likeCnt: number
    collectCnt: number
  }
}

function getRankStyle(rank: number): {
  numberClass: string
  titleClass: string
} {
  if (rank === 1) {
    return {
      numberClass: 'text-2xl font-black text-red-500',
      titleClass: 'text-sm font-bold text-gray-900',
    }
  }
  if (rank === 2) {
    return {
      numberClass: 'text-2xl font-black text-orange-500',
      titleClass: 'text-sm font-bold text-gray-900',
    }
  }
  if (rank === 3) {
    return {
      numberClass: 'text-2xl font-black text-orange-400',
      titleClass: 'text-sm font-bold text-gray-900',
    }
  }
  return {
    numberClass: 'text-lg font-bold text-gray-300',
    titleClass: 'text-sm font-medium text-gray-900',
  }
}

export default function HotRanking() {
  const navigate = useNavigate()

  const {
    data: articles,
    isLoading,
  } = useQuery({
    queryKey: ['hot-ranking-full'],
    queryFn: async () => {
      const res = await api.get<HotArticle[]>('/ranking/hot')
      return res.data || []
    },
  })

  return (
    <div className="flex flex-col min-h-screen bg-white">
      {/* Header */}
      <header className="sticky top-0 z-40 bg-white/95 backdrop-blur-sm border-b border-gray-100">
        <div className="flex items-center px-4 pt-[env(safe-area-inset-top)] h-14">
          <button
            onClick={() => navigate(-1)}
            className="w-9 h-9 flex items-center justify-center -ml-2 rounded-full active:bg-gray-100"
          >
            <ArrowLeft className="w-5 h-5 text-gray-700" />
          </button>
          <div className="flex-1 flex items-center justify-center gap-1.5 pr-5">
            <Flame className="w-5 h-5 text-orange-500" />
            <h1 className="text-base font-semibold text-gray-900">
              热门排行
            </h1>
          </div>
        </div>
      </header>

      {/* Content */}
      <div className="flex-1">
        {isLoading ? (
          <div className="flex items-center justify-center py-20">
            <Loader2 className="w-6 h-6 text-gray-300 animate-spin" />
          </div>
        ) : articles && articles.length > 0 ? (
          <div>
            {articles.map((article, index) => {
              const rank = index + 1
              const { numberClass, titleClass } = getRankStyle(rank)
              const readCnt =
                article.read_cnt ??
                article.interactive?.readCnt ??
                0
              const likeCnt =
                article.like_cnt ??
                article.interactive?.likeCnt ??
                0
              const authorName =
                article.author_name ?? article.authorName ?? ''

              return (
                <div
                  key={article.id}
                  className={`flex items-start gap-3 px-4 py-3.5 border-b border-gray-50 cursor-pointer active:bg-gray-50 ${
                    rank <= 3 ? 'bg-orange-50/30' : ''
                  }`}
                  onClick={() => navigate(`/article/${article.id}`)}
                >
                  {/* Rank number */}
                  <div className="w-8 text-center shrink-0 pt-0.5">
                    <span className={numberClass}>{rank}</span>
                  </div>

                  {/* Article info */}
                  <div className="flex-1 min-w-0">
                    <h3 className={`${titleClass} line-clamp-2 mb-1`}>
                      {article.title}
                    </h3>
                    {article.abstract && (
                      <p className="text-xs text-gray-500 line-clamp-2 mb-2">
                        {article.abstract}
                      </p>
                    )}
                    <div className="flex items-center gap-3 text-xs text-gray-400">
                      {authorName && (
                        <span className="truncate max-w-[100px]">
                          {authorName}
                        </span>
                      )}
                      {readCnt > 0 && (
                        <span className="flex items-center gap-0.5">
                          <Eye className="w-3 h-3" />
                          {formatCount(readCnt)}
                        </span>
                      )}
                      {likeCnt > 0 && (
                        <span className="flex items-center gap-0.5">
                          <Heart className="w-3 h-3" />
                          {formatCount(likeCnt)}
                        </span>
                      )}
                    </div>
                  </div>
                </div>
              )
            })}
          </div>
        ) : (
          <div className="flex flex-col items-center justify-center py-20 text-gray-400">
            <Flame className="w-10 h-10 mb-2 text-gray-200" />
            <p className="text-sm">暂无排行数据</p>
          </div>
        )}
      </div>
    </div>
  )
}
