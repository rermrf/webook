import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  ArrowLeft,
  Gift,
  CheckCircle,
  Loader2,
  Coins,
} from 'lucide-react'
import { api } from '../services/api'

const PRESET_AMOUNTS = [10, 20, 50, 100, 200, 500]

export default function Reward() {
  const { articleId } = useParams()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [selectedAmount, setSelectedAmount] = useState<number | null>(null)
  const [customAmount, setCustomAmount] = useState('')
  const [isCustom, setIsCustom] = useState(false)
  const [success, setSuccess] = useState(false)

  // Fetch balance
  const { data: balance } = useQuery({
    queryKey: ['credit-balance'],
    queryFn: async () => {
      const res = await api.get<{ balance: number }>('/credit/balance')
      return res.data?.balance ?? 0
    },
  })

  // Fetch article info (for author display)
  const { data: article } = useQuery({
    queryKey: ['article', articleId],
    queryFn: async () => {
      const res = await api.get<{
        id: number
        title: string
        author: { id: number; nickname: string; avatar?: string }
      }>(`/articles/pub/${articleId}`)
      return res.data
    },
    enabled: !!articleId,
  })

  // Reward mutation
  const rewardMutation = useMutation({
    mutationFn: async (amount: number) => {
      await api.post('/credit/reward', {
        target_uid: article?.author?.id,
        biz: 'article',
        biz_id: Number(articleId),
        amount,
      })
    },
    onSuccess: () => {
      setSuccess(true)
      queryClient.invalidateQueries({ queryKey: ['credit-balance'] })
    },
  })

  const finalAmount = isCustom
    ? Number(customAmount) || 0
    : selectedAmount || 0

  const handleSelectPreset = (amount: number) => {
    setIsCustom(false)
    setCustomAmount('')
    setSelectedAmount(amount)
  }

  const handleCustomFocus = () => {
    setIsCustom(true)
    setSelectedAmount(null)
  }

  const handleConfirm = () => {
    if (finalAmount <= 0) return
    rewardMutation.mutate(finalAmount)
  }

  const insufficientBalance = balance != null && finalAmount > balance

  // Success state
  if (success) {
    return (
      <div className="flex flex-col min-h-screen bg-white">
        <header className="sticky top-0 z-40 bg-white border-b border-gray-100">
          <div className="flex items-center px-4 pt-[env(safe-area-inset-top)] h-14">
            <button
              onClick={() => navigate(-1)}
              className="w-9 h-9 flex items-center justify-center -ml-2 rounded-full active:bg-gray-100"
            >
              <ArrowLeft className="w-5 h-5 text-gray-700" />
            </button>
            <h1 className="flex-1 text-center text-base font-medium text-gray-900 pr-5">
              打赏作者
            </h1>
          </div>
        </header>

        <div className="flex-1 flex flex-col items-center justify-center px-6">
          <div className="w-16 h-16 rounded-full bg-green-50 flex items-center justify-center mb-4">
            <CheckCircle className="w-8 h-8 text-green-500" />
          </div>
          <h2 className="text-lg font-semibold text-gray-900 mb-2">
            打赏成功
          </h2>
          <p className="text-sm text-gray-500 text-center mb-2">
            已向 {article?.author?.nickname || '作者'} 打赏 {finalAmount} 积分
          </p>
          <p className="text-xs text-gray-400 mb-8">感谢你对创作者的支持！</p>
          <button
            onClick={() => navigate(-1)}
            className="w-full max-w-xs h-11 rounded-full bg-blue-500 text-white text-sm font-medium active:bg-blue-600"
          >
            返回文章
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col min-h-screen bg-gray-50">
      {/* Header */}
      <header className="sticky top-0 z-40 bg-white border-b border-gray-100">
        <div className="flex items-center px-4 pt-[env(safe-area-inset-top)] h-14">
          <button
            onClick={() => navigate(-1)}
            className="w-9 h-9 flex items-center justify-center -ml-2 rounded-full active:bg-gray-100"
          >
            <ArrowLeft className="w-5 h-5 text-gray-700" />
          </button>
          <h1 className="flex-1 text-center text-base font-medium text-gray-900 pr-5">
            打赏作者
          </h1>
        </div>
      </header>

      <div className="flex-1 px-4 pt-6">
        {/* Author card */}
        <div className="flex flex-col items-center mb-8">
          <div className="w-16 h-16 rounded-full bg-blue-100 text-blue-600 flex items-center justify-center text-xl font-semibold mb-3">
            {(article?.author?.nickname || '作者').charAt(0)}
          </div>
          <h2 className="text-base font-semibold text-gray-900">
            {article?.author?.nickname || '加载中...'}
          </h2>
          <p className="text-xs text-gray-400 mt-1 line-clamp-1 max-w-[240px] text-center">
            {article?.title}
          </p>
        </div>

        {/* Preset amounts */}
        <div className="mb-6">
          <h3 className="text-sm font-medium text-gray-700 mb-3">
            选择打赏金额
          </h3>
          <div className="grid grid-cols-3 gap-3">
            {PRESET_AMOUNTS.map((amount) => (
              <button
                key={amount}
                onClick={() => handleSelectPreset(amount)}
                className={`h-12 rounded-xl text-sm font-medium transition-colors ${
                  !isCustom && selectedAmount === amount
                    ? 'bg-blue-500 text-white shadow-sm'
                    : 'bg-white text-gray-700 border border-gray-200 active:border-blue-300'
                }`}
              >
                {amount}
                <span className="text-xs ml-0.5">积分</span>
              </button>
            ))}
          </div>
        </div>

        {/* Custom amount */}
        <div className="mb-6">
          <h3 className="text-sm font-medium text-gray-700 mb-3">
            自定义金额
          </h3>
          <div
            className={`flex items-center h-12 px-4 rounded-xl bg-white border transition-colors ${
              isCustom ? 'border-blue-500 ring-1 ring-blue-200' : 'border-gray-200'
            }`}
          >
            <Coins className="w-4 h-4 text-gray-400 mr-2" />
            <input
              type="number"
              value={customAmount}
              onChange={(e) => setCustomAmount(e.target.value)}
              onFocus={handleCustomFocus}
              placeholder="输入自定义金额"
              className="flex-1 bg-transparent text-sm outline-none placeholder:text-gray-400"
              min={1}
            />
            <span className="text-xs text-gray-400">积分</span>
          </div>
        </div>

        {/* Balance display */}
        <div className="flex items-center justify-between px-4 py-3 rounded-xl bg-white mb-6">
          <span className="text-sm text-gray-500">当前余额</span>
          <span className="text-sm font-semibold text-gray-900">
            {balance?.toLocaleString() ?? '---'} 积分
          </span>
        </div>

        {/* Insufficient balance warning */}
        {insufficientBalance && (
          <p className="text-xs text-red-500 text-center mb-4">
            余额不足，请先充值
          </p>
        )}

        {/* Error message */}
        {rewardMutation.isError && (
          <p className="text-xs text-red-500 text-center mb-4">
            打赏失败，请稍后重试
          </p>
        )}
      </div>

      {/* Confirm button */}
      <div className="sticky bottom-0 bg-white border-t border-gray-100 px-4 py-3 pb-[max(12px,env(safe-area-inset-bottom))]">
        <button
          onClick={handleConfirm}
          disabled={
            finalAmount <= 0 ||
            insufficientBalance ||
            rewardMutation.isPending
          }
          className="w-full h-12 rounded-full bg-gradient-to-r from-orange-400 to-orange-500 text-white text-base font-medium disabled:opacity-40 active:from-orange-500 active:to-orange-600 flex items-center justify-center gap-2"
        >
          {rewardMutation.isPending ? (
            <Loader2 className="w-5 h-5 animate-spin" />
          ) : (
            <>
              <Gift className="w-5 h-5" />
              确认打赏{finalAmount > 0 ? ` ${finalAmount} 积分` : ''}
            </>
          )}
        </button>
      </div>
    </div>
  )
}
