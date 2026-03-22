import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Avatar, Button, Spinner } from '@heroui/react'
import { ArrowLeft } from 'lucide-react'
import { api } from '../services/api'
import type { FollowUser, FollowStats } from '../types'

type TabKey = 'followee' | 'follower'

function formatStatCount(count: number): string {
  if (count >= 10000) return (count / 10000).toFixed(1) + 'w'
  if (count >= 1000) return (count / 1000).toFixed(1) + 'k'
  return String(count)
}

// Backend FollowUserVO: id, nickname, about_me, followed
function UserRow({
  user,
  onToggleFollow,
  isPending,
}: {
  user: FollowUser
  onToggleFollow: () => void
  isPending: boolean
}) {
  const navigate = useNavigate()

  return (
    <div className="flex items-center gap-3 px-4 py-3 border-b border-gray-50">
      <div
        className="shrink-0 cursor-pointer"
        onClick={() => navigate(`/user/${user.id}`)}
      >
        <Avatar size="md">
          <Avatar.Fallback>
            {(user.nickname || '用户').charAt(0)}
          </Avatar.Fallback>
        </Avatar>
      </div>
      <div
        className="flex-1 min-w-0 cursor-pointer"
        onClick={() => navigate(`/user/${user.id}`)}
      >
        <p className="text-sm font-medium text-gray-900 truncate">
          {user.nickname || '匿名用户'}
        </p>
        {user.about_me && (
          <p className="text-xs text-gray-500 truncate mt-0.5">
            {user.about_me}
          </p>
        )}
      </div>
      <Button
        size="sm"
        variant={user.followed ? 'outline' : 'primary'}
        onPress={onToggleFollow}
        isDisabled={isPending}
        className="min-w-[64px] shrink-0"
      >
        {user.followed ? '已关注' : '关注'}
      </Button>
    </div>
  )
}

export default function FollowList() {
  const { id: userId } = useParams()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [activeTab, setActiveTab] = useState<TabKey>('followee')

  const uid = Number(userId) || 0

  // Backend: GET /follow/static?followee=ID  returns { followees, followers }
  const { data: stats } = useQuery({
    queryKey: ['follow-stats', uid],
    queryFn: async () => {
      const res = await api.get<FollowStats>(`/follow/static?followee=${uid}`)
      return res.data
    },
    enabled: uid > 0,
  })

  // Backend: GET /follow/followee?follower=ID  returns FollowUserVO[]
  // FollowUserVO: id, nickname, about_me, followed
  const { data: followees, isLoading: followeesLoading } = useQuery({
    queryKey: ['followees', uid],
    queryFn: async () => {
      const res = await api.get<FollowUser[]>(
        `/follow/followee?follower=${uid}`
      )
      return res.data || []
    },
    enabled: activeTab === 'followee' && uid > 0,
  })

  // Backend: GET /follow/follower?follower=ID  returns FollowUserVO[]
  const { data: followers, isLoading: followersLoading } = useQuery({
    queryKey: ['followers', uid],
    queryFn: async () => {
      const res = await api.get<FollowUser[]>(
        `/follow/follower?follower=${uid}`
      )
      return res.data || []
    },
    enabled: activeTab === 'follower' && uid > 0,
  })

  // Toggle follow mutation
  // Backend: POST /follow/follow  body: { followee }
  // Backend: POST /follow/cancel  body: { followee }
  const toggleFollowMutation = useMutation({
    mutationFn: async ({
      targetUid,
      isCurrentlyFollowing,
    }: {
      targetUid: number
      isCurrentlyFollowing: boolean
    }) => {
      if (isCurrentlyFollowing) {
        await api.post('/follow/cancel', { followee: targetUid })
      } else {
        await api.post('/follow/follow', { followee: targetUid })
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['followees', uid] })
      queryClient.invalidateQueries({ queryKey: ['followers', uid] })
      queryClient.invalidateQueries({ queryKey: ['follow-stats', uid] })
    },
  })

  const users = activeTab === 'followee' ? followees : followers
  const isLoading =
    activeTab === 'followee' ? followeesLoading : followersLoading

  return (
    <div className="flex flex-col min-h-screen">
      {/* Header */}
      <header className="sticky top-0 z-40 bg-white">
        <div className="flex items-center gap-3 px-4 pt-[env(safe-area-inset-top)] h-14">
          <Button
            isIconOnly
            variant="ghost"
            onPress={() => navigate(-1)}
            size="sm"
          >
            <ArrowLeft className="w-5 h-5" />
          </Button>
          <h1 className="flex-1 text-base font-medium text-gray-900">
            关注与粉丝
          </h1>
        </div>

        {/* Tabs */}
        <div className="flex px-4 border-b border-gray-100">
          <button
            onClick={() => setActiveTab('followee')}
            className={`flex-1 py-3 text-center text-sm font-medium relative ${
              activeTab === 'followee' ? 'text-gray-900' : 'text-gray-400'
            }`}
          >
            关注{' '}
            <span className="text-xs">
              {formatStatCount(stats?.followees ?? 0)}
            </span>
            {activeTab === 'followee' && (
              <div className="absolute bottom-0 left-1/2 -translate-x-1/2 w-6 h-0.5 bg-blue-500 rounded-full" />
            )}
          </button>
          <button
            onClick={() => setActiveTab('follower')}
            className={`flex-1 py-3 text-center text-sm font-medium relative ${
              activeTab === 'follower' ? 'text-gray-900' : 'text-gray-400'
            }`}
          >
            粉丝{' '}
            <span className="text-xs">
              {formatStatCount(stats?.followers ?? 0)}
            </span>
            {activeTab === 'follower' && (
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
        ) : users && users.length > 0 ? (
          <div>
            {users.map((user) => (
              <UserRow
                key={user.id}
                user={user}
                onToggleFollow={() =>
                  toggleFollowMutation.mutate({
                    targetUid: user.id,
                    isCurrentlyFollowing: user.followed,
                  })
                }
                isPending={toggleFollowMutation.isPending}
              />
            ))}
          </div>
        ) : (
          <div className="flex flex-col items-center justify-center py-20 text-gray-400">
            <p className="text-sm">
              {activeTab === 'followee'
                ? '还没有关注任何人'
                : '还没有粉丝'}
            </p>
          </div>
        )}
      </div>
    </div>
  )
}
