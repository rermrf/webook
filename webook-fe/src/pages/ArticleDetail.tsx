import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Avatar, Button, Spinner } from '@heroui/react'
import {
  ArrowLeft,
  Heart,
  Bookmark,
  Share2,
  MessageCircle,
  Send,
  MoreHorizontal,
} from 'lucide-react'
import { api } from '../services/api'
import type { ArticleDetail as ArticleDetailType, Comment, GetCommentResp } from '../types'

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

function formatTimestamp(timestamp: number): string {
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

function formatCount(count: number): string {
  if (count >= 10000) return (count / 10000).toFixed(1) + 'w'
  if (count >= 1000) return (count / 1000).toFixed(1) + 'k'
  return String(count)
}

// Backend Comment: { id, content, uid, user_name, user_avatar_url, parent_id, root_id, ctime }
function CommentItem({ comment }: { comment: Comment }) {
  return (
    <div className="px-4 py-3">
      <div className="flex gap-3">
        <Avatar size="sm" className="shrink-0">
          {comment.user_avatar_url && (
            <Avatar.Image src={comment.user_avatar_url} />
          )}
          <Avatar.Fallback>
            {(comment.user_name || '用户').charAt(0)}
          </Avatar.Fallback>
        </Avatar>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 mb-1">
            <span className="text-sm font-medium text-gray-900">
              {comment.user_name || '匿名用户'}
            </span>
            <span className="text-xs text-gray-400">
              {formatTimestamp(comment.ctime)}
            </span>
          </div>
          <p className="text-sm text-gray-700 leading-relaxed">
            {comment.content}
          </p>
        </div>
      </div>
    </div>
  )
}

export default function ArticleDetail() {
  const { id } = useParams()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [commentText, setCommentText] = useState('')
  const [isCommentExpanded, setIsCommentExpanded] = useState(false)

  // Backend: GET /articles/pub/:id  returns flat ArticleVO
  // Fields: id, title, content, author_id, author_name, author_avatar_url, author_followed,
  //         read_cnt, like_cnt, collect_cnt, liked, collected, ctime, utime
  const {
    data: article,
    isLoading,
    error,
  } = useQuery({
    queryKey: ['article', id],
    queryFn: async () => {
      const res = await api.get<ArticleDetailType>(`/articles/pub/${id}`)
      return res.data
    },
    enabled: !!id,
  })

  // Backend: GET /articles/pub/comment?id=X&min_id=0&limit=20
  // Returns { Comments: Comment[] } where Comment has uid, user_name
  const { data: comments } = useQuery({
    queryKey: ['comments', id],
    queryFn: async () => {
      const res = await api.get<GetCommentResp>(
        `/articles/pub/comment?id=${id}&min_id=0&limit=20`
      )
      return res.data?.Comments || []
    },
    enabled: !!id,
  })

  // Backend: POST /articles/pub/like  body: { id: articleId, like: bool }
  const likeMutation = useMutation({
    mutationFn: async () => {
      await api.post('/articles/pub/like', {
        id: Number(id),
        like: !article?.liked,
      })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['article', id] })
    },
  })

  // Backend: POST /articles/pub/collect  body: { id: articleId, collect: bool }
  const collectMutation = useMutation({
    mutationFn: async () => {
      await api.post('/articles/pub/collect', {
        id: Number(id),
        collect: !article?.collected,
      })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['article', id] })
    },
  })

  // Backend: POST /follow/follow  body: { followee: userId }
  // Backend: POST /follow/cancel  body: { followee: userId }
  const followMutation = useMutation({
    mutationFn: async () => {
      if (article?.author_followed) {
        await api.post('/follow/cancel', { followee: article!.author_id })
      } else {
        await api.post('/follow/follow', { followee: article!.author_id })
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['article', id] })
    },
  })

  // Backend: POST /articles/pub/comment  body: { id: bizId, content, parent_id, root_id }
  const commentMutation = useMutation({
    mutationFn: async (content: string) => {
      await api.post('/articles/pub/comment', {
        id: Number(id),
        content,
        parent_id: 0,
        root_id: 0,
      })
    },
    onSuccess: () => {
      setCommentText('')
      setIsCommentExpanded(false)
      queryClient.invalidateQueries({ queryKey: ['comments', id] })
    },
  })

  const handleSubmitComment = () => {
    if (!commentText.trim()) return
    commentMutation.mutate(commentText.trim())
  }

  const handleShare = async () => {
    if (navigator.share) {
      try {
        await navigator.share({
          title: article?.title,
          url: window.location.href,
        })
      } catch {
        // User cancelled or not supported
      }
    }
  }

  if (isLoading) {
    return (
      <div className="flex flex-col min-h-screen">
        <header className="sticky top-0 z-40 bg-white border-b border-gray-100">
          <div className="flex items-center px-4 pt-[env(safe-area-inset-top)] h-14">
            <Button
              isIconOnly
              variant="ghost"
              onPress={() => navigate(-1)}
              size="sm"
            >
              <ArrowLeft className="w-5 h-5" />
            </Button>
          </div>
        </header>
        <div className="flex-1 flex items-center justify-center">
          <Spinner size="lg" />
        </div>
      </div>
    )
  }

  if (error || !article) {
    return (
      <div className="flex flex-col min-h-screen">
        <header className="sticky top-0 z-40 bg-white border-b border-gray-100">
          <div className="flex items-center px-4 pt-[env(safe-area-inset-top)] h-14">
            <Button
              isIconOnly
              variant="ghost"
              onPress={() => navigate(-1)}
              size="sm"
            >
              <ArrowLeft className="w-5 h-5" />
            </Button>
          </div>
        </header>
        <div className="flex-1 flex items-center justify-center text-gray-400">
          <p className="text-sm">文章加载失败</p>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col min-h-screen pb-16">
      {/* Header */}
      <header className="sticky top-0 z-40 bg-white/95 backdrop-blur-sm border-b border-gray-100">
        <div className="flex items-center px-4 pt-[env(safe-area-inset-top)] h-14">
          <Button
            isIconOnly
            variant="ghost"
            onPress={() => navigate(-1)}
            size="sm"
          >
            <ArrowLeft className="w-5 h-5" />
          </Button>
          <h1 className="flex-1 text-center text-base font-medium text-gray-900">
            文章详情
          </h1>
          <Button
            isIconOnly
            variant="ghost"
            size="sm"
          >
            <MoreHorizontal className="w-5 h-5" />
          </Button>
        </div>
      </header>

      {/* Article Content */}
      <div className="flex-1">
        {/* Author Row — flat fields: author_id, author_name, author_avatar_url, author_followed */}
        <div className="flex items-center gap-3 px-4 py-3">
          <div
            className="shrink-0 cursor-pointer"
            onClick={() => navigate(`/user/${article.author_id}`)}
          >
            <Avatar size="md">
              {article.author_avatar_url && (
                <Avatar.Image src={article.author_avatar_url} />
              )}
              <Avatar.Fallback>
                {(article.author_name || '作者').charAt(0)}
              </Avatar.Fallback>
            </Avatar>
          </div>
          <div className="flex-1 min-w-0">
            <p
              className="text-sm font-medium text-gray-900 cursor-pointer"
              onClick={() => navigate(`/user/${article.author_id}`)}
            >
              {article.author_name || '匿名作者'}
            </p>
            <p className="text-xs text-gray-400">
              {formatTime(article.ctime)}
              {` | 阅读 ${formatCount(article.read_cnt)}`}
            </p>
          </div>
          <Button
            size="sm"
            variant={article.author_followed ? 'outline' : 'primary'}
            onPress={() => followMutation.mutate()}
            isDisabled={followMutation.isPending}
            className="min-w-[64px]"
          >
            {article.author_followed ? '已关注' : '关注'}
          </Button>
        </div>

        {/* Title */}
        <div className="px-4 pb-3">
          <h1 className="text-xl font-bold text-gray-900 leading-tight">
            {article.title}
          </h1>
          {/* Tag chips */}
          {(article as ArticleDetailType & { tags?: { id: number; name: string }[] }).tags?.map((t) => (
            <span key={t.id} className="inline-block text-sm text-blue-500 bg-blue-50 px-3 py-1 rounded-full mr-2 mb-2">#{t.name}</span>
          ))}
        </div>

        {/* Content */}
        <div className="px-4 pb-6">
          <div className="text-sm text-gray-700 leading-relaxed whitespace-pre-wrap">
            {article.content}
          </div>
        </div>

        {/* Divider */}
        <div className="h-2 bg-gray-50" />

        {/* Comments Section */}
        <div className="px-4 py-3">
          <h3 className="text-base font-semibold text-gray-900">
            评论 {comments && comments.length > 0 ? `(${comments.length})` : ''}
          </h3>
        </div>

        {comments && comments.length > 0 ? (
          <div className="divide-y divide-gray-50">
            {comments.map((comment) => (
              <CommentItem key={comment.id} comment={comment} />
            ))}
          </div>
        ) : (
          <div className="flex flex-col items-center justify-center py-10 text-gray-400">
            <MessageCircle className="w-10 h-10 mb-2 text-gray-200" />
            <p className="text-sm">暂无评论，来说点什么吧</p>
          </div>
        )}
      </div>

      {/* Bottom Action Bar */}
      <div className="fixed bottom-0 left-1/2 -translate-x-1/2 w-full max-w-[430px] bg-white border-t border-gray-100 z-50">
        <div className="flex items-center gap-2 px-4 py-2 pb-[max(8px,env(safe-area-inset-bottom))]">
          {isCommentExpanded ? (
            <div className="flex-1 flex items-center gap-2">
              <input
                type="text"
                value={commentText}
                onChange={(e) => setCommentText(e.target.value)}
                placeholder="写评论..."
                className="flex-1 h-9 px-3 rounded-full bg-gray-100 text-sm outline-none focus:bg-gray-50 focus:ring-1 focus:ring-blue-300"
                autoFocus
                onKeyDown={(e) => {
                  if (e.key === 'Enter') handleSubmitComment()
                }}
              />
              <Button
                isIconOnly
                size="sm"
                variant="primary"
                onPress={handleSubmitComment}
                isDisabled={commentMutation.isPending || !commentText.trim()}
              >
                <Send className="w-4 h-4" />
              </Button>
              <Button
                isIconOnly
                size="sm"
                variant="ghost"
                onPress={() => {
                  setIsCommentExpanded(false)
                  setCommentText('')
                }}
              >
                <span className="text-xs text-gray-400">取消</span>
              </Button>
            </div>
          ) : (
            <>
              <button
                className="flex-1 h-9 px-4 rounded-full bg-gray-100 text-left text-sm text-gray-400"
                onClick={() => setIsCommentExpanded(true)}
              >
                写评论...
              </button>
              <button
                className="flex items-center gap-1 px-3 py-2"
                onClick={() => likeMutation.mutate()}
              >
                <Heart
                  className={`w-5 h-5 ${
                    article.liked
                      ? 'fill-red-500 text-red-500'
                      : 'text-gray-400'
                  }`}
                />
                <span
                  className={`text-xs ${
                    article.liked ? 'text-red-500' : 'text-gray-400'
                  }`}
                >
                  {formatCount(article.like_cnt)}
                </span>
              </button>
              <button
                className="flex items-center gap-1 px-3 py-2"
                onClick={() => collectMutation.mutate()}
              >
                <Bookmark
                  className={`w-5 h-5 ${
                    article.collected
                      ? 'fill-yellow-500 text-yellow-500'
                      : 'text-gray-400'
                  }`}
                />
                <span
                  className={`text-xs ${
                    article.collected ? 'text-yellow-500' : 'text-gray-400'
                  }`}
                >
                  {formatCount(article.collect_cnt)}
                </span>
              </button>
              <button
                className="flex items-center px-3 py-2"
                onClick={handleShare}
              >
                <Share2 className="w-5 h-5 text-gray-400" />
              </button>
            </>
          )}
        </div>
      </div>
    </div>
  )
}
