import { useState, useEffect, useRef } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Avatar, Button, Spinner } from '@heroui/react'
import { ArrowLeft, Send } from 'lucide-react'
import { api } from '../services/api'

// Backend MessageVO: id(string), sender_id, receiver_id, msg_type, content, status, ctime
interface MessageItem {
  id: string
  sender_id: number
  receiver_id: number
  content: string
  msg_type: number
  status: number
  ctime: number
}

interface PeerUser {
  user_id: number
  nickname: string
  avatar: string
}

interface ConversationInfo {
  conversation_id: string
  peer_user: PeerUser
}

function formatMsgTime(timestamp: number): string {
  const ts = timestamp < 1e12 ? timestamp * 1000 : timestamp
  const date = new Date(ts)
  const now = new Date()
  const isToday =
    date.getDate() === now.getDate() &&
    date.getMonth() === now.getMonth() &&
    date.getFullYear() === now.getFullYear()
  const hours = date.getHours().toString().padStart(2, '0')
  const mins = date.getMinutes().toString().padStart(2, '0')
  if (isToday) return `${hours}:${mins}`
  const yesterday = new Date(now)
  yesterday.setDate(yesterday.getDate() - 1)
  const isYesterday =
    date.getDate() === yesterday.getDate() &&
    date.getMonth() === yesterday.getMonth() &&
    date.getFullYear() === yesterday.getFullYear()
  if (isYesterday) return `昨天 ${hours}:${mins}`
  return `${date.getMonth() + 1}/${date.getDate()} ${hours}:${mins}`
}

function shouldShowTimestamp(
  current: MessageItem,
  previous: MessageItem | undefined
): boolean {
  if (!previous) return true
  const curTs = current.ctime < 1e12 ? current.ctime * 1000 : current.ctime
  const prevTs =
    previous.ctime < 1e12 ? previous.ctime * 1000 : previous.ctime
  return curTs - prevTs > 5 * 60 * 1000 // 5 minutes gap
}

function getMyUserId(): number {
  try {
    const token = localStorage.getItem('accessToken')
    if (!token) return 0
    const payload = JSON.parse(atob(token.split('.')[1]))
    return payload.uid || payload.user_id || payload.sub || 0
  } catch {
    return 0
  }
}

export default function ChatDetail() {
  const { id: conversationId } = useParams()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [inputText, setInputText] = useState('')
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const myUserId = getMyUserId()

  // Fetch conversation info
  const { data: convInfo } = useQuery({
    queryKey: ['conversation-info', conversationId],
    queryFn: async () => {
      const res = await api.get<ConversationInfo[]>('/im/conversations')
      const list = res.data || []
      return list.find((c) => c.conversation_id === conversationId) || null
    },
    enabled: !!conversationId,
  })

  // Fetch online status
  const { data: onlineStatus } = useQuery({
    queryKey: ['online-status', convInfo?.peer_user?.user_id],
    queryFn: async () => {
      const res = await api.get<{ online: boolean }>(
        `/im/online/${convInfo!.peer_user.user_id}`
      )
      return res.data?.online ?? false
    },
    enabled: !!convInfo?.peer_user?.user_id,
    refetchInterval: 30000,
  })

  // Fetch messages — backend returns flat MessageVO[] (not wrapped in {messages, has_more})
  const {
    data: messagesData,
    isLoading,
  } = useQuery({
    queryKey: ['messages', conversationId],
    queryFn: async () => {
      const res = await api.get<MessageItem[]>(
        `/im/conversations/${conversationId}/messages?cursor=0&limit=50`
      )
      return res.data || []
    },
    enabled: !!conversationId,
    refetchInterval: 10000,
  })

  const messages = messagesData || []

  // Mark as read
  useEffect(() => {
    if (conversationId) {
      api.post(`/im/conversations/${conversationId}/read`).catch(() => {})
      queryClient.invalidateQueries({ queryKey: ['im-unread'] })
      queryClient.invalidateQueries({ queryKey: ['conversations'] })
    }
  }, [conversationId, queryClient])

  // Scroll to bottom on new messages
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages.length])

  // Send message mutation — uses POST /im/send REST endpoint
  const sendMutation = useMutation({
    mutationFn: async (content: string) => {
      const peerId = convInfo?.peer_user?.user_id
      if (!peerId) throw new Error('No peer user')
      await api.post('/im/send', {
        receiver_id: peerId,
        msg_type: 1,
        content,
      })
    },
    onSuccess: () => {
      setInputText('')
      queryClient.invalidateQueries({
        queryKey: ['messages', conversationId],
      })
    },
  })

  const handleSend = () => {
    const text = inputText.trim()
    if (!text) return
    sendMutation.mutate(text)
  }

  const peer = convInfo?.peer_user

  return (
    <div className="flex flex-col h-screen">
      {/* Header */}
      <header className="sticky top-0 z-40 bg-white border-b border-gray-100 shrink-0">
        <div className="flex items-center gap-3 px-4 pt-[env(safe-area-inset-top)] h-14">
          <Button
            isIconOnly
            variant="ghost"
            onPress={() => navigate(-1)}
            size="sm"
          >
            <ArrowLeft className="w-5 h-5" />
          </Button>
          <div className="flex-1 min-w-0">
            <h1 className="text-base font-medium text-gray-900 truncate">
              {peer?.nickname || '聊天'}
            </h1>
            {peer && (
              <p className="text-[11px] text-gray-400">
                {onlineStatus ? (
                  <span className="text-green-500">在线</span>
                ) : (
                  '离线'
                )}
              </p>
            )}
          </div>
        </div>
      </header>

      {/* Messages Area */}
      <div className="flex-1 overflow-y-auto px-4 py-3 bg-gray-50">
        {isLoading ? (
          <div className="flex items-center justify-center py-20">
            <Spinner size="lg" />
          </div>
        ) : messages.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-20 text-gray-400">
            <p className="text-sm">暂无消息，发送一条开始聊天吧</p>
          </div>
        ) : (
          <div className="space-y-1">
            {messages.map((msg, idx) => {
              const isMine = msg.sender_id === myUserId
              const showTime = shouldShowTimestamp(msg, messages[idx - 1])
              return (
                <div key={msg.id}>
                  {showTime && (
                    <div className="text-center my-3">
                      <span className="text-[11px] text-gray-400 bg-gray-200/60 px-2 py-0.5 rounded-full">
                        {formatMsgTime(msg.ctime)}
                      </span>
                    </div>
                  )}
                  <div
                    className={`flex items-end gap-2 mb-2 ${
                      isMine ? 'flex-row-reverse' : 'flex-row'
                    }`}
                  >
                    {!isMine && (
                      <Avatar size="sm" className="shrink-0 w-8 h-8">
                        {peer?.avatar && (
                          <Avatar.Image src={peer.avatar} />
                        )}
                        <Avatar.Fallback>
                          {(peer?.nickname || '用').charAt(0)}
                        </Avatar.Fallback>
                      </Avatar>
                    )}
                    <div
                      className={`max-w-[70%] px-3 py-2 rounded-2xl text-sm leading-relaxed break-words ${
                        isMine
                          ? 'bg-blue-500 text-white rounded-br-sm'
                          : 'bg-white text-gray-800 rounded-bl-sm shadow-sm'
                      }`}
                    >
                      {msg.content}
                    </div>
                  </div>
                </div>
              )
            })}
            <div ref={messagesEndRef} />
          </div>
        )}
      </div>

      {/* Input Bar */}
      <div className="shrink-0 bg-white border-t border-gray-100">
        <div className="flex items-center gap-2 px-4 py-2 pb-[max(8px,env(safe-area-inset-bottom))]">
          <input
            type="text"
            value={inputText}
            onChange={(e) => setInputText(e.target.value)}
            placeholder="输入消息..."
            className="flex-1 h-10 px-4 rounded-full bg-gray-100 text-sm outline-none focus:bg-gray-50 focus:ring-1 focus:ring-blue-300"
            onKeyDown={(e) => {
              if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault()
                handleSend()
              }
            }}
          />
          <Button
            isIconOnly
            size="sm"
            variant="primary"
            onPress={handleSend}
            isDisabled={!inputText.trim() || sendMutation.isPending}
            className="shrink-0"
          >
            <Send className="w-4 h-4" />
          </Button>
        </div>
      </div>
    </div>
  )
}
