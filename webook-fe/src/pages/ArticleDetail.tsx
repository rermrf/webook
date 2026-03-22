import { useParams, useNavigate } from 'react-router-dom'
import { Button } from '@heroui/react'
import { ArrowLeft } from 'lucide-react'

export default function ArticleDetail() {
  const { id } = useParams()
  const navigate = useNavigate()

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
          <h1 className="flex-1 text-center text-base font-medium text-gray-900 pr-10">
            文章详情
          </h1>
        </div>
      </header>

      <div className="flex-1 flex items-center justify-center text-gray-400">
        <p className="text-sm">文章 #{id} 详情页开发中...</p>
      </div>
    </div>
  )
}
