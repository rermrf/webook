import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, Spinner } from '@heroui/react'
import {
  ArrowLeft,
  Wallet,
  ArrowUpCircle,
  ArrowDownCircle,
  CalendarCheck,
  Gift,
  ShoppingCart,
  Star,
} from 'lucide-react'
import { api } from '../services/api'
import type { CreditFlow, DailyStatus } from '../types'

function formatTime(timestamp: number): string {
  const ts = timestamp < 1e12 ? timestamp * 1000 : timestamp
  const date = new Date(ts)
  const month = (date.getMonth() + 1).toString().padStart(2, '0')
  const day = date.getDate().toString().padStart(2, '0')
  const hours = date.getHours().toString().padStart(2, '0')
  const mins = date.getMinutes().toString().padStart(2, '0')
  return `${month}-${day} ${hours}:${mins}`
}

function getFlowIcon(description: string) {
  if (description.includes('签到')) return <CalendarCheck className="w-5 h-5" />
  if (description.includes('奖励') || description.includes('赠送'))
    return <Gift className="w-5 h-5" />
  if (description.includes('购买') || description.includes('消费'))
    return <ShoppingCart className="w-5 h-5" />
  if (description.includes('打赏')) return <Star className="w-5 h-5" />
  return <Wallet className="w-5 h-5" />
}

export default function CreditWallet() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  // Backend: GET /credit/balance  returns raw number in data (not { balance: N })
  const { data: balance, isLoading: balanceLoading } = useQuery({
    queryKey: ['credit-balance'],
    queryFn: async () => {
      const res = await api.get<number>('/credit/balance')
      // Backend returns raw number: { data: 580 }
      return res.data ?? 0
    },
  })

  // Backend: POST /credit/flows  body: { offset, limit } (not page/page_size)
  // Returns CreditFlowVO[]: id, biz, biz_id, change_amt, balance, description, ctime
  const { data: flows, isLoading: flowsLoading } = useQuery({
    queryKey: ['credit-flows'],
    queryFn: async () => {
      const res = await api.post<CreditFlow[]>('/credit/flows', {
        offset: 0,
        limit: 50,
      })
      return res.data || []
    },
  })

  // Backend: GET /credit/daily-status  returns DailyStatusVO[]
  // Each: biz, earned_count, earned_amt, daily_limit, remaining
  const { data: dailyStatuses } = useQuery({
    queryKey: ['credit-daily-status'],
    queryFn: async () => {
      const res = await api.get<DailyStatus[]>('/credit/daily-status')
      return res.data || []
    },
  })

  // Check if user has "signed in" today (check if there's a sign_in biz with earned_count > 0)
  const hasSignedToday = dailyStatuses?.some(
    (s) => s.biz === 'sign_in' && s.earned_count > 0
  ) ?? false

  const signInMutation = useMutation({
    mutationFn: async () => {
      await api.post('/credit/sign-in')
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['credit-balance'] })
      queryClient.invalidateQueries({ queryKey: ['credit-daily-status'] })
      queryClient.invalidateQueries({ queryKey: ['credit-flows'] })
    },
  })

  const isLoading = balanceLoading || flowsLoading

  return (
    <div className="flex flex-col min-h-screen bg-gray-50">
      {/* Header */}
      <header className="sticky top-0 z-40 bg-gradient-to-r from-blue-500 to-blue-600">
        <div className="flex items-center justify-between px-4 pt-[env(safe-area-inset-top)] h-14">
          <Button
            isIconOnly
            variant="ghost"
            onPress={() => navigate(-1)}
            size="sm"
            className="text-white"
          >
            <ArrowLeft className="w-5 h-5" />
          </Button>
          <h1 className="text-base font-medium text-white">积分钱包</h1>
          <div className="w-9" />
        </div>
      </header>

      {/* Balance Card */}
      <div className="px-4 -mt-0">
        <div className="bg-gradient-to-br from-blue-500 via-blue-600 to-indigo-700 rounded-2xl p-6 text-white shadow-lg">
          <p className="text-sm text-blue-100 mb-1">当前积分</p>
          {balanceLoading ? (
            <div className="py-2">
              <Spinner size="sm" />
            </div>
          ) : (
            <p className="text-4xl font-bold tracking-tight mb-1">
              {(typeof balance === 'number' ? balance : 0).toLocaleString()}
              <span className="text-base font-normal text-blue-200 ml-1">
                积分
              </span>
            </p>
          )}

          {/* Action Buttons */}
          <div className="flex items-center gap-3 mt-5">
            <button className="flex-1 flex flex-col items-center gap-1 py-2 rounded-xl bg-white/15 active:bg-white/25">
              <ArrowUpCircle className="w-5 h-5" />
              <span className="text-xs">充值</span>
            </button>
            <button className="flex-1 flex flex-col items-center gap-1 py-2 rounded-xl bg-white/15 active:bg-white/25">
              <ArrowDownCircle className="w-5 h-5" />
              <span className="text-xs">提现</span>
            </button>
            <button
              className={`flex-1 flex flex-col items-center gap-1 py-2 rounded-xl ${
                hasSignedToday
                  ? 'bg-white/10 opacity-60'
                  : 'bg-white/15 active:bg-white/25'
              }`}
              onClick={() => !hasSignedToday && signInMutation.mutate()}
              disabled={hasSignedToday || signInMutation.isPending}
            >
              <CalendarCheck className="w-5 h-5" />
              <span className="text-xs">
                {hasSignedToday ? '已签到' : '签到'}
              </span>
            </button>
          </div>
        </div>
      </div>

      {/* Flow List */}
      <div className="mt-4 bg-white rounded-t-2xl flex-1">
        <div className="px-4 py-3 border-b border-gray-100">
          <h2 className="text-base font-semibold text-gray-900">积分流水</h2>
        </div>

        {isLoading ? (
          <div className="flex items-center justify-center py-20">
            <Spinner size="lg" />
          </div>
        ) : flows && flows.length > 0 ? (
          <div>
            {flows.map((flow) => {
              // Backend field: change_amt (not amount)
              const isEarn = flow.change_amt > 0
              return (
                <div
                  key={flow.id}
                  className="flex items-center gap-3 px-4 py-3 border-b border-gray-50"
                >
                  <div
                    className={`w-10 h-10 rounded-full flex items-center justify-center shrink-0 ${
                      isEarn
                        ? 'bg-green-50 text-green-500'
                        : 'bg-red-50 text-red-500'
                    }`}
                  >
                    {getFlowIcon(flow.description)}
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="text-sm text-gray-900 truncate">
                      {flow.description}
                    </p>
                    <p className="text-xs text-gray-400 mt-0.5">
                      {formatTime(flow.ctime)}
                    </p>
                  </div>
                  <div className="shrink-0 text-right">
                    <span
                      className={`text-base font-semibold ${
                        isEarn ? 'text-green-500' : 'text-red-500'
                      }`}
                    >
                      {isEarn ? '+' : ''}
                      {flow.change_amt}
                    </span>
                  </div>
                </div>
              )
            })}
          </div>
        ) : (
          <div className="flex flex-col items-center justify-center py-20 text-gray-400">
            <Wallet className="w-12 h-12 mb-3 text-gray-200" />
            <p className="text-sm">暂无积分记录</p>
          </div>
        )}
      </div>
    </div>
  )
}
