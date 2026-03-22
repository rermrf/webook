import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Avatar, Button, Chip, Spinner } from '@heroui/react'
import {
  ArrowLeft,
  Heart,
  Bookmark,
  Share2,
  MessageCircle,
  Send,
} from 'lucide-react'
import { api } from '../services/api'
import type { ArticleDetail as ArticleDetailType, Comment } from '../types'

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

function formatCount(count: number): string {
  if (count >= 10000) return (count / 10000).toFixed(1) + 'w'
  if (count >= 1000) return (count / 1000).toFixed(1) + 'k'
  return String(count)
}

function CommentItem({ comment }: { comment: Comment }) {
  const [showReplies, setShowReplies] = useState(false)

  return (
    <div className="px-4 py-3">
      <div className="flex gap-3">
        <Avatar size="sm" className="shrink-0">
          <Avatar.Fallback>
            {(comment.user?.nickname || '用户').charAt(0)}
          </Avatar.Fallback>
        </Avatar>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 mb-1">
            <span className="text-sm font-medium text-gray-900">
              {comment.user?.nickname || '匿名用户'}
            </span>
            <span className="text-xs text-gray-400">
              {formatTime(comment.ctime)}
            </span>
          </div>
          <p className="text-sm text-gray-700 leading-relaxed">
            {comment.content}
          </p>
          {comment.replyCnt > 0 && (
            <button
              className="mt-2 text-xs text-blue-500 font-medium"
              onClick={() => setShowReplies(!showReplies)}
            >
              {showReplies ? '收起回复' : `查看 ${comment.replyCnt} 条回复`}
            </button>
          )}
          {showReplies && comment.replies && comment.replies.length > 0 && (
            <div className="mt-2 pl-2 border-l-2 border-gray-100 space-y-3">
              {comment.replies.map((reply) => (
                <div key={reply.id} className="flex gap-2">
                  <Avatar size="sm" className="shrink-0 w-6 h-6">
                    <Avatar.Fallback>
                      {(reply.user?.nickname || '用户').charAt(0)}
                    </Avatar.Fallback>
                  </Avatar>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-0.5">
                      <span className="text-xs font-medium text-gray-900">
                        {reply.user?.nickname || '匿名用户'}
                      </span>
                      <span className="text-xs text-gray-400">
                        {formatTime(reply.ctime)}
                      </span>
                    </div>
                    <p className="text-xs text-gray-700">{reply.content}</p>
                  </div>
                </div>
              ))}
            </div>
          )}
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

  // Fetch article detail
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

  // Fetch follow status
  const { data: followStatus } = useQuery({
    queryKey: ['follow-check', article?.author?.id],
    queryFn: async () => {
      const res = await api.get<{ follow: boolean }>(
        `/follow/check?uid=${article!.author.id}`
      )
      return res.data
    },
    enabled: !!article?.author?.id,
  })

  // Fetch comments
  const { data: comments } = useQuery({
    queryKey: ['comments', id],
    queryFn: async () => {
      const res = await api.get<Comment[]>(
        `/articles/pub/comment?biz=article&biz_id=${id}&min_id=0&limit=20`
      )
      return res.data || []
    },
    enabled: !!id,
  })

  // Like mutation
  const likeMutation = useMutation({
    mutationFn: async () => {
      await api.post('/articles/pub/like', {
        id: 0,
        biz: 'article',
        biz_id: Number(id),
        like: !article?.interactive?.liked,
      })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['article', id] })
    },
  })

  // Collect mutation
  const collectMutation = useMutation({
    mutationFn: async () => {
      await api.post('/articles/pub/collect', {
        id: 0,
        biz: 'article',
        biz_id: Number(id),
        collect: !article?.interactive?.collected,
      })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['article', id] })
    },
  })

  // Follow mutation
  const followMutation = useMutation({
    mutationFn: async () => {
      const isFollowing = followStatus?.follow
      if (isFollowing) {
        await api.post('/follow/cancel', { followee: article!.author.id })
      } else {
        await api.post('/follow/follow', { followee: article!.author.id })
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['follow-check', article?.author?.id],
      })
    },
  })

  // Comment mutation
  const commentMutation = useMutation({
    mutationFn: async (content: string) => {
      await api.post('/articles/pub/comment', {
        biz: 'article',
        biz_id: Number(id),
        content,
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

  const isFollowing = followStatus?.follow ?? false
  const interactive = article.interactive

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
          <h1 className="flex-1 text-center text-base font-medium text-gray-900 pr-10">
            文章详情
          </h1>
        </div>
      </header>

      {/* Article Content */}
      <div className="flex-1">
        {/* Author Row */}
        <div className="flex items-center gap-3 px-4 py-3">
          <div
            className="shrink-0 cursor-pointer"
            onClick={() => navigate(`/user/${article.author?.id}`)}
          >
            <Avatar size="md">
              {article.author?.avatar && (
                <Avatar.Image src={article.author.avatar} />
              )}
              <Avatar.Fallback>
                {(article.author?.nickname || '作者').charAt(0)}
              </Avatar.Fallback>
            </Avatar>
          </div>
          <div className="flex-1 min-w-0">
            <p
              className="text-sm font-medium text-gray-900 cursor-pointer"
              onClick={() => navigate(`/user/${article.author?.id}`)}
            >
              {article.author?.nickname || '匿名作者'}
            </p>
            <p className="text-xs text-gray-400">
              {formatTime(article.ctime)}
              {interactive && ` | 阅读 ${formatCount(interactive.readCnt)}`}
            </p>
          </div>
          <Button
            size="sm"
            variant={isFollowing ? 'outline' : 'primary'}
            onPress={() => followMutation.mutate()}
            isDisabled={followMutation.isPending}
            className="min-w-[64px]"
          >
            {isFollowing ? '已关注' : '关注'}
          </Button>
        </div>

        {/* Title */}
        <div className="px-4 pb-3">
          <h1 className="text-xl font-bold text-gray-900 leading-tight">
            {article.title}
          </h1>
        </div>

        {/* Tags */}
        {article.tags && article.tags.length > 0 && (
          <div className="flex gap-2 px-4 pb-4 flex-wrap">
            {article.tags.map((tag) => (
              <Chip key={tag.id} size="sm" variant="soft" color="accent">
                {tag.name}
              </Chip>
            ))}
          </div>
        )}

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
                    interactive?.liked
                      ? 'fill-red-500 text-red-500'
                      : 'text-gray-400'
                  }`}
                />
                <span
                  className={`text-xs ${
                    interactive?.liked ? 'text-red-500' : 'text-gray-400'
                  }`}
                >
                  {interactive ? formatCount(interactive.likeCnt) : '0'}
                </span>
              </button>
              <button
                className="flex items-center gap-1 px-3 py-2"
                onClick={() => collectMutation.mutate()}
              >
                <Bookmark
                  className={`w-5 h-5 ${
                    interactive?.collected
                      ? 'fill-yellow-500 text-yellow-500'
                      : 'text-gray-400'
                  }`}
                />
                <span
                  className={`text-xs ${
                    interactive?.collected ? 'text-yellow-500' : 'text-gray-400'
                  }`}
                >
                  {interactive ? formatCount(interactive.collectCnt) : '0'}
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
