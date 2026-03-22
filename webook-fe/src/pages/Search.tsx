import { useState, useEffect, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Avatar, Button, Chip, Spinner } from '@heroui/react'
import {
  Search as SearchIcon,
  ArrowLeft,
  TrendingUp,
  ChevronRight,
  Eye,
  X,
} from 'lucide-react'
import { api } from '../services/api'
import type { Article, User } from '../types'

interface SearchResults {
  articles: SearchArticle[]
  users: SearchUser[]
}

interface SearchArticle extends Article {
  author?: { id: number; nickname: string; avatar?: string }
  interactive?: {
    readCnt: number
    likeCnt: number
    collectCnt: number
  }
}

interface SearchUser extends User {
  isFollowed?: boolean
}

function formatCount(count: number): string {
  if (count >= 10000) return (count / 10000).toFixed(1) + 'w'
  if (count >= 1000) return (count / 1000).toFixed(1) + 'k'
  return String(count)
}

function useDebounce<T>(value: T, delay: number): T {
  const [debouncedValue, setDebouncedValue] = useState<T>(value)

  useEffect(() => {
    const timer = setTimeout(() => setDebouncedValue(value), delay)
    return () => clearTimeout(timer)
  }, [value, delay])

  return debouncedValue
}

export default function Search() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [query, setQuery] = useState('')
  const [isFocused, setIsFocused] = useState(false)
  const debouncedQuery = useDebounce(query, 300)

  // Hot keywords
  const { data: hotKeywords } = useQuery({
    queryKey: ['hot-keywords'],
    queryFn: async () => {
      const res = await api.get<string[]>('/search/hot-keywords')
      return res.data || []
    },
  })

  // Hot ranking
  const { data: hotArticles } = useQuery({
    queryKey: ['hot-ranking'],
    queryFn: async () => {
      const res = await api.get<SearchArticle[]>('/ranking/hot')
      return res.data || []
    },
  })

  // Recommended authors
  const { data: recommendedAuthors } = useQuery({
    queryKey: ['recommend'],
    queryFn: async () => {
      const res = await api.get<SearchUser[]>('/users/recommend')
      return res.data || []
    },
  })

  // Search results
  const {
    data: searchResults,
    isLoading: searchLoading,
  } = useQuery({
    queryKey: ['search', debouncedQuery],
    queryFn: async () => {
      const res = await api.get<SearchResults>(
        `/search?expression=${encodeURIComponent(debouncedQuery)}`
      )
      return res.data
    },
    enabled: debouncedQuery.length > 0,
  })

  // Follow mutation for recommended authors
  const followMutation = useMutation({
    mutationFn: async ({ userId, isFollowed }: { userId: number; isFollowed: boolean }) => {
      if (isFollowed) {
        await api.post('/follow/cancel', { followee: userId })
      } else {
        await api.post('/follow/follow', { followee: userId })
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['recommend'] })
      queryClient.invalidateQueries({ queryKey: ['search', debouncedQuery] })
    },
  })

  const handleKeywordClick = useCallback((keyword: string) => {
    setQuery(keyword)
  }, [])

  const hasQuery = debouncedQuery.length > 0
  const showResults = hasQuery && searchResults

  return (
    <div className="flex flex-col min-h-screen">
      {/* Search Header */}
      <header className="sticky top-0 z-40 bg-white border-b border-gray-100">
        <div className="flex items-center gap-2 px-4 pt-[env(safe-area-inset-top)] h-14">
          {isFocused && (
            <Button
              isIconOnly
              variant="ghost"
              onPress={() => {
                setIsFocused(false)
                setQuery('')
              }}
              size="sm"
            >
              <ArrowLeft className="w-5 h-5" />
            </Button>
          )}
          <div className="flex-1 flex items-center gap-2 h-9 px-3 rounded-full bg-gray-100">
            <SearchIcon className="w-4 h-4 text-gray-400 shrink-0" />
            <input
              type="text"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              onFocus={() => setIsFocused(true)}
              placeholder="搜索文章、用户..."
              className="flex-1 bg-transparent text-sm outline-none placeholder:text-gray-400"
            />
            {query && (
              <button onClick={() => setQuery('')}>
                <X className="w-4 h-4 text-gray-400" />
              </button>
            )}
          </div>
          {isFocused && (
            <button
              className="text-sm text-blue-500 font-medium shrink-0"
              onClick={() => {
                setIsFocused(false)
                if (!query) setQuery('')
              }}
            >
              取消
            </button>
          )}
        </div>
      </header>

      {/* Content */}
      <div className="flex-1">
        {searchLoading && hasQuery ? (
          <div className="flex items-center justify-center py-20">
            <Spinner size="lg" />
          </div>
        ) : showResults ? (
          /* Search Results */
          <div>
            {/* Article Results */}
            {searchResults.articles && searchResults.articles.length > 0 && (
              <div>
                <div className="px-4 py-3">
                  <h3 className="text-sm font-semibold text-gray-900">
                    文章
                  </h3>
                </div>
                {searchResults.articles.map((article) => (
                  <div
                    key={article.id}
                    className="px-4 py-3 border-b border-gray-50 active:bg-gray-50 cursor-pointer"
                    onClick={() => navigate(`/article/${article.id}`)}
                  >
                    <h3 className="text-sm font-semibold text-gray-900 mb-1 line-clamp-2">
                      {article.title}
                    </h3>
                    {article.abstract && (
                      <p className="text-xs text-gray-500 line-clamp-2 mb-1">
                        {article.abstract}
                      </p>
                    )}
                    <div className="flex items-center gap-2 text-xs text-gray-400">
                      {article.author && (
                        <span>{article.author.nickname}</span>
                      )}
                      {article.interactive && (
                        <span>
                          <Eye className="w-3 h-3 inline mr-0.5" />
                          {formatCount(article.interactive.readCnt)}
                        </span>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            )}

            {/* User Results */}
            {searchResults.users && searchResults.users.length > 0 && (
              <div>
                <div className="px-4 py-3">
                  <h3 className="text-sm font-semibold text-gray-900">
                    用户
                  </h3>
                </div>
                {searchResults.users.map((user) => (
                  <div
                    key={user.id}
                    className="flex items-center gap-3 px-4 py-3 border-b border-gray-50"
                  >
                    <div
                      className="shrink-0 cursor-pointer"
                      onClick={() => navigate(`/user/${user.id}`)}
                    >
                      <Avatar size="md">
                        {user.avatar && <Avatar.Image src={user.avatar} />}
                        <Avatar.Fallback>
                          {(user.nickname || '用户').charAt(0)}
                        </Avatar.Fallback>
                      </Avatar>
                    </div>
                    <div
                      className="flex-1 min-w-0 cursor-pointer"
                      onClick={() => navigate(`/user/${user.id}`)}
                    >
                      <p className="text-sm font-medium text-gray-900">
                        {user.nickname}
                      </p>
                      {user.aboutMe && (
                        <p className="text-xs text-gray-500 line-clamp-1">
                          {user.aboutMe}
                        </p>
                      )}
                    </div>
                    <Button
                      size="sm"
                      variant={user.isFollowed ? 'outline' : 'primary'}
                      onPress={() =>
                        followMutation.mutate({
                          userId: user.id,
                          isFollowed: !!user.isFollowed,
                        })
                      }
                      className="min-w-[64px]"
                    >
                      {user.isFollowed ? '已关注' : '关注'}
                    </Button>
                  </div>
                ))}
              </div>
            )}

            {/* No Results */}
            {(!searchResults.articles || searchResults.articles.length === 0) &&
              (!searchResults.users || searchResults.users.length === 0) && (
                <div className="flex flex-col items-center justify-center py-20 text-gray-400">
                  <SearchIcon className="w-10 h-10 mb-3 text-gray-200" />
                  <p className="text-sm">没有找到相关内容</p>
                  <p className="text-xs mt-1">换个关键词试试吧</p>
                </div>
              )}
          </div>
        ) : (
          /* Discovery Content (no query) */
          <div>
            {/* Hot Keywords */}
            {hotKeywords && hotKeywords.length > 0 && (
              <div className="px-4 pt-4 pb-2">
                <h3 className="text-base font-semibold text-gray-900 mb-3">
                  热门搜索
                </h3>
                <div className="flex flex-wrap gap-2">
                  {hotKeywords.map((keyword, index) => (
                    <Chip
                      key={index}
                      variant="soft"
                      color="default"
                      className="cursor-pointer"
                      onClick={() => handleKeywordClick(keyword)}
                    >
                      {keyword}
                    </Chip>
                  ))}
                </div>
              </div>
            )}

            {/* Hot Ranking */}
            {hotArticles && hotArticles.length > 0 && (
              <div className="px-4 pt-4 pb-2">
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center gap-2">
                    <TrendingUp className="w-4 h-4 text-orange-500" />
                    <h3 className="text-base font-semibold text-gray-900">
                      热门榜单
                    </h3>
                  </div>
                  <button
                    className="flex items-center gap-0.5 text-xs text-gray-400"
                    onClick={() => navigate('/ranking')}
                  >
                    查看全部
                    <ChevronRight className="w-3 h-3" />
                  </button>
                </div>
                <div className="space-y-0">
                  {hotArticles.slice(0, 5).map((article, index) => (
                    <div
                      key={article.id}
                      className="flex items-start gap-3 py-3 border-b border-gray-50 cursor-pointer active:bg-gray-50"
                      onClick={() => navigate(`/article/${article.id}`)}
                    >
                      <span
                        className={`text-lg font-bold w-6 text-center shrink-0 ${
                          index < 3
                            ? 'text-orange-500'
                            : 'text-gray-300'
                        }`}
                      >
                        {index + 1}
                      </span>
                      <div className="flex-1 min-w-0">
                        <h4 className="text-sm font-medium text-gray-900 line-clamp-2 mb-1">
                          {article.title}
                        </h4>
                        <div className="flex items-center gap-2 text-xs text-gray-400">
                          {article.interactive && (
                            <span>
                              <Eye className="w-3 h-3 inline mr-0.5" />
                              {formatCount(article.interactive.readCnt)}
                            </span>
                          )}
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Recommended Authors */}
            {recommendedAuthors && recommendedAuthors.length > 0 && (
              <div className="px-4 pt-4 pb-6">
                <h3 className="text-base font-semibold text-gray-900 mb-3">
                  推荐作者
                </h3>
                <div className="space-y-0">
                  {recommendedAuthors.map((author) => (
                    <div
                      key={author.id}
                      className="flex items-center gap-3 py-3 border-b border-gray-50"
                    >
                      <div
                        className="shrink-0 cursor-pointer"
                        onClick={() => navigate(`/user/${author.id}`)}
                      >
                        <Avatar size="md">
                          {author.avatar && (
                            <Avatar.Image src={author.avatar} />
                          )}
                          <Avatar.Fallback>
                            {(author.nickname || '用户').charAt(0)}
                          </Avatar.Fallback>
                        </Avatar>
                      </div>
                      <div
                        className="flex-1 min-w-0 cursor-pointer"
                        onClick={() => navigate(`/user/${author.id}`)}
                      >
                        <p className="text-sm font-medium text-gray-900">
                          {author.nickname}
                        </p>
                        {author.aboutMe && (
                          <p className="text-xs text-gray-500 line-clamp-1">
                            {author.aboutMe}
                          </p>
                        )}
                      </div>
                      <Button
                        size="sm"
                        variant={author.isFollowed ? 'outline' : 'primary'}
                        onPress={() =>
                          followMutation.mutate({
                            userId: author.id,
                            isFollowed: !!author.isFollowed,
                          })
                        }
                        className="min-w-[64px]"
                      >
                        {author.isFollowed ? '已关注' : '关注'}
                      </Button>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Fallback empty state */}
            {!hotKeywords?.length &&
              !hotArticles?.length &&
              !recommendedAuthors?.length && (
                <div className="flex flex-col items-center justify-center py-20 text-gray-400">
                  <SearchIcon className="w-10 h-10 mb-3 text-gray-200" />
                  <p className="text-sm">搜索文章、用户</p>
                </div>
              )}
          </div>
        )}
      </div>
    </div>
  )
}
