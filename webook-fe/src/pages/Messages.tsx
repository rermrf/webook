import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { Avatar, Spinner } from '@heroui/react'
import { PenLine, Bell, MessageCircle } from 'lucide-react'
import { api } from '../services/api'

interface PeerUser {
  user_id: number
  nickname: string
  avatar: string
}

interface LastMsg {
  content: string
  msg_type: number
  ctime: number
}

interface Conversation {
  conversation_id: string
  members: number[]
  peer_user: PeerUser
  last_msg: LastMsg
  unread_count: number
  utime: number
}

interface NotificationItem {
  id: number
  group_type: string
  source_name: string
  content: string
  is_read: boolean
  ctime: number
}

type TabKey = 'chat' | 'notify'

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
  if (days < 7) return `${days}天前`
  const date = new Date(ts)
  return `${date.getMonth() + 1}月${date.getDate()}日`
}

function ConversationItem({
  conversation,
  onClick,
}: {
  conversation: Conversation
  onClick: () => void
}) {
  const peer = conversation.peer_user
  return (
    <div
      className="flex items-center gap-3 px-4 py-3 border-b border-gray-50 active:bg-gray-50 cursor-pointer"
      onClick={onClick}
    >
      <div className="relative shrink-0">
        <Avatar size="md">
          {peer.avatar && <Avatar.Image src={peer.avatar} />}
          <Avatar.Fallback>
            {(peer.nickname || '用户').charAt(0)}
          </Avatar.Fallback>
        </Avatar>
        {conversation.unread_count > 0 && (
          <span className="absolute -top-1 -right-1 min-w-[18px] h-[18px] px-1 rounded-full bg-red-500 text-white text-[10px] font-bold flex items-center justify-center">
            {conversation.unread_count > 99
              ? '99+'
              : conversation.unread_count}
          </span>
        )}
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex items-center justify-between mb-0.5">
          <span className="text-sm font-medium text-gray-900 truncate">
            {peer.nickname || '未知用户'}
          </span>
          <span className="text-[11px] text-gray-400 shrink-0 ml-2">
            {formatTime(conversation.last_msg?.ctime || conversation.utime)}
          </span>
        </div>
        <p className="text-xs text-gray-500 truncate">
          {conversation.last_msg?.content || '暂无消息'}
        </p>
      </div>
    </div>
  )
}

function NotificationRow({ item }: { item: NotificationItem }) {
  return (
    <div
      className={`flex items-start gap-3 px-4 py-3 border-b border-gray-50 ${
        item.is_read ? '' : 'bg-blue-50/30'
      }`}
    >
      <div className="w-9 h-9 rounded-full bg-blue-100 text-blue-500 flex items-center justify-center shrink-0 mt-0.5">
        <Bell className="w-4 h-4" />
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex items-center justify-between mb-0.5">
          <span className="text-sm font-medium text-gray-900">
            {item.source_name || '系统通知'}
          </span>
          <span className="text-[11px] text-gray-400 shrink-0 ml-2">
            {formatTime(item.ctime)}
          </span>
        </div>
        <p className="text-xs text-gray-600 line-clamp-2">{item.content}</p>
      </div>
      {!item.is_read && (
        <span className="w-2 h-2 rounded-full bg-red-500 shrink-0 mt-2" />
      )}
    </div>
  )
}

export default function Messages() {
  const [activeTab, setActiveTab] = useState<TabKey>('chat')
  const navigate = useNavigate()

  const { data: conversations, isLoading: convLoading } = useQuery({
    queryKey: ['conversations'],
    queryFn: async () => {
      const res = await api.get<Conversation[]>('/im/conversations')
      return res.data || []
    },
    enabled: activeTab === 'chat',
  })

  const { data: notifications, isLoading: notifyLoading } = useQuery({
    queryKey: ['notifications'],
    queryFn: async () => {
      const res = await api.get<NotificationItem[]>(
        '/notifications/list?offset=0&limit=20'
      )
      return res.data || []
    },
    enabled: activeTab === 'notify',
  })

  const { data: imUnread } = useQuery({
    queryKey: ['im-unread'],
    queryFn: async () => {
      const res = await api.get<{ total: number }>('/im/unread-count')
      return res.data?.total || 0
    },
  })

  const { data: notifyUnread } = useQuery({
    queryKey: ['notify-unread'],
    queryFn: async () => {
      const res = await api.get<{ total: number }>(
        '/notifications/unread-count'
      )
      return res.data?.total || 0
    },
  })

  const isLoading = activeTab === 'chat' ? convLoading : notifyLoading

  return (
    <div className="flex flex-col min-h-screen">
      {/* Header */}
      <header className="sticky top-0 z-40 bg-white">
        <div className="flex items-center justify-between px-4 pt-[env(safe-area-inset-top)] h-14">
          <h1 className="text-xl font-bold text-gray-900">消息</h1>
          <button
            onClick={() => navigate('/chat/new')}
            className="p-1.5 text-gray-500 hover:text-blue-500"
          >
            <PenLine className="w-5 h-5" />
          </button>
        </div>

        {/* Tabs */}
        <div className="flex px-4 border-b border-gray-100">
          <button
            onClick={() => setActiveTab('chat')}
            className={`flex-1 py-3 text-center text-sm font-medium relative ${
              activeTab === 'chat' ? 'text-gray-900' : 'text-gray-400'
            }`}
          >
            <span className="relative">
              聊天
              {(imUnread ?? 0) > 0 && (
                <span className="absolute -top-2 -right-5 min-w-[16px] h-[16px] px-1 rounded-full bg-red-500 text-white text-[10px] font-bold flex items-center justify-center">
                  {imUnread}
                </span>
              )}
            </span>
            {activeTab === 'chat' && (
              <div className="absolute bottom-0 left-1/2 -translate-x-1/2 w-6 h-0.5 bg-blue-500 rounded-full" />
            )}
          </button>
          <button
            onClick={() => setActiveTab('notify')}
            className={`flex-1 py-3 text-center text-sm font-medium relative ${
              activeTab === 'notify' ? 'text-gray-900' : 'text-gray-400'
            }`}
          >
            <span className="relative">
              通知
              {(notifyUnread ?? 0) > 0 && (
                <span className="absolute -top-2 -right-5 min-w-[16px] h-[16px] px-1 rounded-full bg-red-500 text-white text-[10px] font-bold flex items-center justify-center">
                  {notifyUnread}
                </span>
              )}
            </span>
            {activeTab === 'notify' && (
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
        ) : activeTab === 'chat' ? (
          conversations && conversations.length > 0 ? (
            <div>
              {conversations.map((conv) => (
                <ConversationItem
                  key={conv.conversation_id}
                  conversation={conv}
                  onClick={() =>
                    navigate(`/chat/${conv.conversation_id}`)
                  }
                />
              ))}
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center py-20 text-gray-400">
              <MessageCircle className="w-12 h-12 mb-3 text-gray-200" />
              <p className="text-sm">暂无会话</p>
              <p className="text-xs mt-1">去关注的人那里发起聊天吧</p>
            </div>
          )
        ) : notifications && notifications.length > 0 ? (
          <div>
            {notifications.map((item) => (
              <NotificationRow key={item.id} item={item} />
            ))}
          </div>
        ) : (
          <div className="flex flex-col items-center justify-center py-20 text-gray-400">
            <Bell className="w-12 h-12 mb-3 text-gray-200" />
            <p className="text-sm">暂无通知</p>
          </div>
        )}
      </div>
    </div>
  )
}
