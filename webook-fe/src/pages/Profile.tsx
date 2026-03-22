import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Avatar, Button, Spinner, Tabs } from '@heroui/react'
import {
  ArrowLeft,
  Settings,
  ChevronRight,
  FileText,
  Wallet,
  Heart,
} from 'lucide-react'
import { api } from '../services/api'
import type { User, Article } from '../types'

function formatTime(timestamp: number): string {
  const now = Date.now()
  const ts = timestamp < 1e12 ? timestamp * 1000 : timestamp
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

function formatStatCount(count: number): string {
  if (count >= 10000) return (count / 10000).toFixed(1) + 'w'
  if (count >= 1000) return (count / 1000).toFixed(1) + 'k'
  return String(count)
}

interface FollowStatsData {
  followee_count: number
  follower_count: number
}

interface ArticleListItem extends Article {
  authorName?: string
  tags?: { id: number; name: string }[]
  interactive?: {
    likeCnt: number
    collectCnt: number
    readCnt: number
  }
}

function ArticleListCard({ article, onClick }: { article: ArticleListItem; onClick: () => void }) {
  return (
    <div
      className="px-4 py-3 border-b border-gray-50 active:bg-gray-50 cursor-pointer"
      onClick={onClick}
    >
      <h3 className="text-sm font-semibold text-gray-900 mb-1 line-clamp-2">
        {article.title}
      </h3>
      {article.abstract && (
        <p className="text-xs text-gray-500 line-clamp-2 mb-2">
          {article.abstract}
        </p>
      )}
      <div className="flex items-center gap-3 text-gray-400">
        <span className="text-xs">{formatTime(article.utime || article.ctime)}</span>
        {article.interactive && (
          <span className="text-xs">
            <Heart className="w-3 h-3 inline mr-0.5" />
            {article.interactive.likeCnt}
          </span>
        )}
      </div>
    </div>
  )
}

function EmptyState({ text }: { text: string }) {
  return (
    <div className="flex flex-col items-center justify-center py-16 text-gray-400">
      <p className="text-sm">{text}</p>
    </div>
  )
}

export default function Profile() {
  const { id: routeId } = useParams()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [activeTab, setActiveTab] = useState<string | number>('articles')

  // Determine if viewing own profile
  const isOwnProfile = !routeId

  // Fetch profile
  const { data: profile, isLoading: profileLoading } = useQuery({
    queryKey: ['profile', routeId || 'me'],
    queryFn: async () => {
      const url = isOwnProfile ? '/users/profile' : `/users/profile/${routeId}`
      const res = await api.get<User>(url)
      return res.data
    },
  })

  // Fetch follow stats
  const { data: followStats } = useQuery({
    queryKey: ['follow-stats', profile?.id],
    queryFn: async () => {
      const res = await api.post<FollowStatsData>('/follow/static', {
        uid: profile!.id,
      })
      return res.data
    },
    enabled: !!profile?.id,
  })

  // Check if following (for other's profile)
  const { data: followStatus } = useQuery({
    queryKey: ['follow-check', profile?.id],
    queryFn: async () => {
      const res = await api.get<{ follow: boolean }>(
        `/follow/check?uid=${profile!.id}`
      )
      return res.data
    },
    enabled: !isOwnProfile && !!profile?.id,
  })

  // Follow/unfollow mutation
  const followMutation = useMutation({
    mutationFn: async () => {
      const isFollowing = followStatus?.follow
      if (isFollowing) {
        await api.post('/follow/cancel', { followee: profile!.id })
      } else {
        await api.post('/follow/follow', { followee: profile!.id })
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['follow-check', profile?.id],
      })
      queryClient.invalidateQueries({
        queryKey: ['follow-stats', profile?.id],
      })
    },
  })

  // Fetch articles (own)
  const { data: articles, isLoading: articlesLoading } = useQuery({
    queryKey: ['profile-articles', profile?.id],
    queryFn: async () => {
      const res = await api.post<ArticleListItem[]>('/articles/list', {
        page: 1,
        page_size: 20,
      })
      return res.data || []
    },
    enabled: activeTab === 'articles' && isOwnProfile,
  })

  // Fetch liked article IDs
  const { data: likedIds, isLoading: likedLoading } = useQuery({
    queryKey: ['profile-liked'],
    queryFn: async () => {
      const res = await api.get<number[]>(
        '/articles/pub/liked?offset=0&limit=20'
      )
      return res.data || []
    },
    enabled: activeTab === 'liked' && isOwnProfile,
  })

  // Fetch collected article IDs
  const { data: collectedIds, isLoading: collectedLoading } = useQuery({
    queryKey: ['profile-collected'],
    queryFn: async () => {
      const res = await api.get<number[]>(
        '/articles/pub/collected?offset=0&limit=20'
      )
      return res.data || []
    },
    enabled: activeTab === 'collected' && isOwnProfile,
  })

  // Fetch history
  const { data: historyIds, isLoading: historyLoading } = useQuery({
    queryKey: ['profile-history'],
    queryFn: async () => {
      const res = await api.get<number[]>(
        '/history/list?cursor=0&limit=20'
      )
      return res.data || []
    },
    enabled: activeTab === 'history' && isOwnProfile,
  })

  if (profileLoading) {
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
            <h1 className="text-base font-medium text-gray-900">个人主页</h1>
            <div className="w-9" />
          </div>
        </header>
        <div className="flex-1 flex items-center justify-center">
          <Spinner size="lg" />
        </div>
      </div>
    )
  }

  const isFollowing = followStatus?.follow ?? false

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
          <h1 className="text-base font-medium text-gray-900">个人主页</h1>
          {isOwnProfile ? (
            <Button
              isIconOnly
              variant="ghost"
              size="sm"
              onPress={() => navigate('/settings')}
            >
              <Settings className="w-5 h-5 text-gray-500" />
            </Button>
          ) : (
            <div className="w-9" />
          )}
        </div>
      </header>

      {/* Profile Info */}
      <div className="flex flex-col items-center pt-6 pb-4 px-4">
        <Avatar size="lg" className="w-20 h-20 text-2xl mb-3">
          {profile?.avatar && <Avatar.Image src={profile.avatar} />}
          <Avatar.Fallback>
            {(profile?.nickname || '用户').charAt(0)}
          </Avatar.Fallback>
        </Avatar>
        <h2 className="text-lg font-bold text-gray-900 mb-1">
          {profile?.nickname || '匿名用户'}
        </h2>
        {profile?.aboutMe && (
          <p className="text-sm text-gray-500 text-center mb-3 px-8 line-clamp-2">
            {profile.aboutMe}
          </p>
        )}

        {/* Stats Row */}
        <div className="flex items-center justify-center gap-8 py-3 w-full">
          <div className="flex flex-col items-center">
            <span className="text-lg font-bold text-gray-900">
              {formatStatCount(followStats?.followee_count ?? 0)}
            </span>
            <span className="text-xs text-gray-400">关注</span>
          </div>
          <div className="w-px h-8 bg-gray-200" />
          <div className="flex flex-col items-center">
            <span className="text-lg font-bold text-gray-900">
              {formatStatCount(followStats?.follower_count ?? 0)}
            </span>
            <span className="text-xs text-gray-400">粉丝</span>
          </div>
          <div className="w-px h-8 bg-gray-200" />
          <div className="flex flex-col items-center">
            <span className="text-lg font-bold text-gray-900">
              {articles?.length ?? 0}
            </span>
            <span className="text-xs text-gray-400">文章</span>
          </div>
        </div>

        {/* Action Button */}
        {isOwnProfile ? (
          <Button
            variant="outline"
            fullWidth
            className="mt-2"
            onPress={() => navigate('/profile/edit')}
          >
            编辑资料
          </Button>
        ) : (
          <Button
            variant={isFollowing ? 'outline' : 'primary'}
            fullWidth
            className="mt-2"
            onPress={() => followMutation.mutate()}
            isDisabled={followMutation.isPending}
          >
            {isFollowing ? '已关注' : '关注'}
          </Button>
        )}
      </div>

      {/* Quick Links (own profile only) */}
      {isOwnProfile && (
        <div className="border-t border-b border-gray-100 bg-white mb-2">
          <button
            className="flex items-center justify-between w-full px-4 py-3 active:bg-gray-50"
            onClick={() => navigate('/drafts')}
          >
            <div className="flex items-center gap-3">
              <FileText className="w-5 h-5 text-gray-400" />
              <span className="text-sm text-gray-700">我的草稿</span>
            </div>
            <ChevronRight className="w-4 h-4 text-gray-300" />
          </button>
          <div className="h-px bg-gray-50 ml-12" />
          <button
            className="flex items-center justify-between w-full px-4 py-3 active:bg-gray-50"
            onClick={() => navigate('/credit')}
          >
            <div className="flex items-center gap-3">
              <Wallet className="w-5 h-5 text-gray-400" />
              <span className="text-sm text-gray-700">积分钱包</span>
            </div>
            <div className="flex items-center gap-1">
              <span className="text-sm text-orange-500 font-medium">580</span>
              <ChevronRight className="w-4 h-4 text-gray-300" />
            </div>
          </button>
          <div className="h-px bg-gray-50 ml-12" />
          <button
            className="flex items-center justify-between w-full px-4 py-3 active:bg-gray-50"
            onClick={() => navigate('/settings')}
          >
            <div className="flex items-center gap-3">
              <Settings className="w-5 h-5 text-gray-400" />
              <span className="text-sm text-gray-700">设置</span>
            </div>
            <ChevronRight className="w-4 h-4 text-gray-300" />
          </button>
        </div>
      )}

      {/* Tabs */}
      <div className="flex-1">
        <Tabs
          selectedKey={activeTab as string}
          onSelectionChange={setActiveTab}
        >
          <Tabs.ListContainer>
            <Tabs.List className="w-full grid grid-cols-4">
              <Tabs.Tab id="articles">文章</Tabs.Tab>
              <Tabs.Tab id="collected">收藏</Tabs.Tab>
              <Tabs.Tab id="liked">点赞</Tabs.Tab>
              <Tabs.Tab id="history">历史</Tabs.Tab>
            </Tabs.List>
            <Tabs.Indicator />
          </Tabs.ListContainer>

          <Tabs.Panel id="articles">
            {articlesLoading ? (
              <div className="flex items-center justify-center py-10">
                <Spinner size="md" />
              </div>
            ) : articles && articles.length > 0 ? (
              <div>
                {articles.map((article) => (
                  <ArticleListCard
                    key={article.id}
                    article={article}
                    onClick={() => navigate(`/article/${article.id}`)}
                  />
                ))}
              </div>
            ) : (
              <EmptyState text="还没有发布文章" />
            )}
          </Tabs.Panel>

          <Tabs.Panel id="collected">
            {collectedLoading ? (
              <div className="flex items-center justify-center py-10">
                <Spinner size="md" />
              </div>
            ) : collectedIds && collectedIds.length > 0 ? (
              <div>
                {collectedIds.map((bizId) => (
                  <div
                    key={bizId}
                    className="px-4 py-3 border-b border-gray-50 active:bg-gray-50 cursor-pointer"
                    onClick={() => navigate(`/article/${bizId}`)}
                  >
                    <p className="text-sm text-gray-700">文章 #{bizId}</p>
                  </div>
                ))}
              </div>
            ) : (
              <EmptyState text="还没有收藏文章" />
            )}
          </Tabs.Panel>

          <Tabs.Panel id="liked">
            {likedLoading ? (
              <div className="flex items-center justify-center py-10">
                <Spinner size="md" />
              </div>
            ) : likedIds && likedIds.length > 0 ? (
              <div>
                {likedIds.map((bizId) => (
                  <div
                    key={bizId}
                    className="px-4 py-3 border-b border-gray-50 active:bg-gray-50 cursor-pointer"
                    onClick={() => navigate(`/article/${bizId}`)}
                  >
                    <p className="text-sm text-gray-700">文章 #{bizId}</p>
                  </div>
                ))}
              </div>
            ) : (
              <EmptyState text="还没有点赞文章" />
            )}
          </Tabs.Panel>

          <Tabs.Panel id="history">
            {historyLoading ? (
              <div className="flex items-center justify-center py-10">
                <Spinner size="md" />
              </div>
            ) : historyIds && historyIds.length > 0 ? (
              <div>
                {historyIds.map((bizId) => (
                  <div
                    key={bizId}
                    className="px-4 py-3 border-b border-gray-50 active:bg-gray-50 cursor-pointer"
                    onClick={() => navigate(`/article/${bizId}`)}
                  >
                    <p className="text-sm text-gray-700">文章 #{bizId}</p>
                  </div>
                ))}
              </div>
            ) : (
              <EmptyState text="暂无浏览记录" />
            )}
          </Tabs.Panel>
        </Tabs>
      </div>
    </div>
  )
}
