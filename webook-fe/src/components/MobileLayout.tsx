import { Outlet } from 'react-router-dom'
import { TabBar } from './TabBar'

export function MobileLayout() {
  return (
    <div className="mx-auto max-w-[430px] min-h-screen bg-white relative">
      <main className="pb-20 min-h-screen">
        <Outlet />
      </main>
      <TabBar />
    </div>
  )
}
