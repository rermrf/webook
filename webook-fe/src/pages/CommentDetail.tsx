import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  ArrowLeft,
  Heart,
  MessageCircle,
  Send,
  ArrowDownUp,
  Loader2,
} from 'lucide-react'
import { api } from '../services/api'
import { formatTime } from '../utils/formatTime'
import type { Comment, GetCommentResp } from '../types'

interface ReplyTarget {
  id: number
  rootId: number
  nickname: string
}

function CommentItem({
  comment,
  onReply,
}: {
  comment: Comment
  onReply: (target: ReplyTarget) => void
}) {
  const [liked, setLiked] = useState(false)

  return (
    <div className="px-4 py-3">
      <div className="flex gap-3">
        {/* Avatar */}
        <div className="w-8 h-8 rounded-full bg-blue-100 text-blue-600 flex items-center justify-center text-sm font-medium shrink-0">
          {(comment.user_name || '用户').charAt(0)}
        </div>
        <div className="flex-1 min-w-0">
          {/* Name + Time */}
          <div className="flex items-center gap-2 mb-1">
            <span className="text-sm font-medium text-gray-900">
              {comment.user_name || '匿名用户'}
            </span>
            <span className="text-xs text-gray-400">
              {formatTime(comment.ctime)}
            </span>
          </div>

          {/* Content */}
          <p className="text-sm text-gray-700 leading-relaxed">
            {comment.content}
          </p>

          {/* Actions: like + reply */}
          <div className="flex items-center gap-4 mt-2">
            <button
              className="flex items-center gap-1 text-xs text-gray-400 active:text-red-500"
              onClick={() => setLiked(!liked)}
            >
              <Heart
                className={`w-3.5 h-3.5 ${liked ? 'fill-red-500 text-red-500' : ''}`}
              />
              <span className={liked ? 'text-red-500' : ''}>
                {liked ? '1' : '点赞'}
              </span>
            </button>
            <button
              className="flex items-center gap-1 text-xs text-gray-400 active:text-blue-500"
              onClick={() =>
                onReply({
                  id: comment.id,
                  rootId: comment.root_id || comment.id,
                  nickname: comment.user_name || '匿名用户',
                })
              }
            >
              <MessageCircle className="w-3.5 h-3.5" />
              <span>回复</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}

export default function CommentDetail() {
  const { bizType, bizId } = useParams()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [commentText, setCommentText] = useState('')
  const [replyTarget, setReplyTarget] = useState<ReplyTarget | null>(null)
  const [sortBy, setSortBy] = useState<'newest' | 'hottest'>('newest')

  // Backend: GET /articles/pub/comment?id=X&min_id=0&limit=20
  // Returns { Comments: Comment[] } where Comment has uid, user_name
  const {
    data: comments,
    isLoading,
  } = useQuery({
    queryKey: ['comments-detail', bizType, bizId],
    queryFn: async () => {
      const res = await api.get<GetCommentResp>(
        `/articles/pub/comment?id=${bizId}&min_id=0&limit=100`
      )
      return res.data?.Comments || []
    },
    enabled: !!bizId,
  })

  // Backend: POST /articles/pub/comment  body: { id: bizId, content, parent_id, root_id }
  const commentMutation = useMutation({
    mutationFn: async (content: string) => {
      const body: Record<string, unknown> = {
        id: Number(bizId),
        content,
        parent_id: replyTarget?.id ?? 0,
        root_id: replyTarget?.rootId ?? 0,
      }
      await api.post('/articles/pub/comment', body)
    },
    onSuccess: () => {
      setCommentText('')
      setReplyTarget(null)
      queryClient.invalidateQueries({
        queryKey: ['comments-detail', bizType, bizId],
      })
    },
  })

  const handleSubmitComment = () => {
    if (!commentText.trim()) return
    commentMutation.mutate(commentText.trim())
  }

  const handleReply = (target: ReplyTarget) => {
    setReplyTarget(target)
    const input = document.getElementById('comment-input') as HTMLInputElement
    input?.focus()
  }

  const cancelReply = () => {
    setReplyTarget(null)
  }

  const sortedComments = comments
    ? sortBy === 'newest'
      ? [...comments].sort((a, b) => b.ctime - a.ctime)
      : [...comments]
    : []

  return (
    <div className="flex flex-col min-h-screen bg-white">
      {/* Header */}
      <header className="sticky top-0 z-40 bg-white/95 backdrop-blur-sm border-b border-gray-100">
        <div className="flex items-center justify-between px-4 pt-[env(safe-area-inset-top)] h-14">
          <button
            onClick={() => navigate(-1)}
            className="w-9 h-9 flex items-center justify-center -ml-2 rounded-full active:bg-gray-100"
          >
            <ArrowLeft className="w-5 h-5 text-gray-700" />
          </button>
          <h1 className="text-base font-medium text-gray-900">
            评论{comments && comments.length > 0 ? ` (${comments.length})` : ''}
          </h1>
          <button
            onClick={() =>
              setSortBy(sortBy === 'newest' ? 'hottest' : 'newest')
            }
            className="flex items-center gap-1 text-sm text-gray-500 active:text-blue-500"
          >
            <ArrowDownUp className="w-4 h-4" />
            {sortBy === 'newest' ? '最新' : '最热'}
          </button>
        </div>
      </header>

      {/* Comment list */}
      <div className="flex-1 pb-20">
        {isLoading ? (
          <div className="flex items-center justify-center py-20">
            <Loader2 className="w-6 h-6 text-gray-300 animate-spin" />
          </div>
        ) : sortedComments.length > 0 ? (
          <div className="divide-y divide-gray-50">
            {sortedComments.map((comment) => (
              <CommentItem
                key={comment.id}
                comment={comment}
                onReply={handleReply}
              />
            ))}
          </div>
        ) : (
          <div className="flex flex-col items-center justify-center py-20 text-gray-400">
            <MessageCircle className="w-10 h-10 mb-2 text-gray-200" />
            <p className="text-sm">暂无评论，来说点什么吧</p>
          </div>
        )}
      </div>

      {/* Bottom input bar */}
      <div className="fixed bottom-0 left-1/2 -translate-x-1/2 w-full max-w-[430px] bg-white border-t border-gray-100 z-50">
        {/* Reply indicator */}
        {replyTarget && (
          <div className="flex items-center justify-between px-4 py-1.5 bg-gray-50 text-xs text-gray-500">
            <span>
              回复 @{replyTarget.nickname}
            </span>
            <button
              onClick={cancelReply}
              className="text-gray-400 active:text-gray-600"
            >
              取消
            </button>
          </div>
        )}
        <div className="flex items-center gap-2 px-4 py-2 pb-[max(8px,env(safe-area-inset-bottom))]">
          <input
            id="comment-input"
            type="text"
            value={commentText}
            onChange={(e) => setCommentText(e.target.value)}
            placeholder={
              replyTarget
                ? `回复 @${replyTarget.nickname}...`
                : '发表评论...'
            }
            className="flex-1 h-9 px-4 rounded-full bg-gray-100 text-sm outline-none focus:bg-gray-50 focus:ring-1 focus:ring-blue-300 placeholder:text-gray-400"
            onKeyDown={(e) => {
              if (e.key === 'Enter') handleSubmitComment()
            }}
          />
          <button
            onClick={handleSubmitComment}
            disabled={commentMutation.isPending || !commentText.trim()}
            className="w-9 h-9 flex items-center justify-center rounded-full bg-blue-500 text-white disabled:opacity-40 active:bg-blue-600"
          >
            {commentMutation.isPending ? (
              <Loader2 className="w-4 h-4 animate-spin" />
            ) : (
              <Send className="w-4 h-4" />
            )}
          </button>
        </div>
      </div>
    </div>
  )
}
