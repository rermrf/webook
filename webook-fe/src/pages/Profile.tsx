import { User } from 'lucide-react'

export default function Profile() {
  return (
    <div className="flex flex-col min-h-screen">
      <header className="sticky top-0 z-40 bg-white border-b border-gray-100">
        <div className="flex items-center justify-center px-4 pt-[env(safe-area-inset-top)] h-14">
          <h1 className="text-base font-medium text-gray-900">我的</h1>
        </div>
      </header>

      <div className="flex-1 flex flex-col items-center justify-center text-gray-400 gap-3">
        <User className="w-12 h-12 text-gray-300" />
        <p className="text-sm">个人主页开发中...</p>
      </div>
    </div>
  )
}
