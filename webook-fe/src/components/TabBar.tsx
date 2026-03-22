import { useLocation, useNavigate } from 'react-router-dom'
import { Home, Compass, PlusCircle, MessageCircle, User } from 'lucide-react'

const tabs = [
  { key: '/', label: '首页', icon: Home },
  { key: '/search', label: '发现', icon: Compass },
  { key: '/create', label: '创作', icon: PlusCircle, isCenter: true },
  { key: '/messages', label: '消息', icon: MessageCircle },
  { key: '/profile', label: '我的', icon: User },
]

export function TabBar() {
  const location = useLocation()
  const navigate = useNavigate()

  return (
    <nav className="fixed bottom-0 left-1/2 -translate-x-1/2 w-full max-w-[430px] bg-white border-t border-gray-100 z-50">
      <div className="flex items-center justify-around h-16 pb-[env(safe-area-inset-bottom)]">
        {tabs.map((tab) => {
          const isActive = location.pathname === tab.key
          const Icon = tab.icon

          if (tab.isCenter) {
            return (
              <button
                key={tab.key}
                onClick={() => navigate(tab.key)}
                className="flex flex-col items-center justify-center -mt-4"
              >
                <div className="w-12 h-12 rounded-full bg-blue-500 flex items-center justify-center shadow-lg">
                  <Icon className="w-6 h-6 text-white" />
                </div>
              </button>
            )
          }

          return (
            <button
              key={tab.key}
              onClick={() => navigate(tab.key)}
              className="flex flex-col items-center justify-center gap-0.5 flex-1"
            >
              <Icon
                className={`w-5 h-5 ${isActive ? 'text-blue-500' : 'text-gray-400'}`}
              />
              <span
                className={`text-[10px] ${isActive ? 'text-blue-500 font-medium' : 'text-gray-400'}`}
              >
                {tab.label}
              </span>
            </button>
          )
        })}
      </div>
    </nav>
  )
}
