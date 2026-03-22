import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { MobileLayout } from './components/MobileLayout'
import Login from './pages/Login'
import Register from './pages/Register'
import Home from './pages/Home'
import ArticleDetail from './pages/ArticleDetail'
import Profile from './pages/Profile'
import Messages from './pages/Messages'
import Search from './pages/Search'
import ChatDetail from './pages/ChatDetail'
import CreditWallet from './pages/CreditWallet'
import ArticleEditor from './pages/ArticleEditor'
import Drafts from './pages/Drafts'
import EditProfile from './pages/EditProfile'
import FollowList from './pages/FollowList'
import CommentDetail from './pages/CommentDetail'
import Reward from './pages/Reward'
import HotRanking from './pages/HotRanking'
import BrowseHistory from './pages/BrowseHistory'
import TagDetail from './pages/TagDetail'

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
        <Route element={<MobileLayout />}>
          <Route path="/" element={<Home />} />
          <Route path="/article/:id" element={<ArticleDetail />} />
          <Route path="/profile" element={<Profile />} />
          <Route path="/user/:id" element={<Profile />} />
          <Route path="/messages" element={<Messages />} />
          <Route path="/search" element={<Search />} />
          <Route path="/chat/:id" element={<ChatDetail />} />
          <Route path="/credit" element={<CreditWallet />} />
          <Route path="/write" element={<ArticleEditor />} />
          <Route path="/write/:id" element={<ArticleEditor />} />
          <Route path="/drafts" element={<Drafts />} />
          <Route path="/edit-profile" element={<EditProfile />} />
          <Route path="/follow/:id" element={<FollowList />} />
          <Route path="/comments/:bizType/:bizId" element={<CommentDetail />} />
          <Route path="/reward/:articleId" element={<Reward />} />
          <Route path="/ranking" element={<HotRanking />} />
          <Route path="/history" element={<BrowseHistory />} />
          <Route path="/tag/:id" element={<TagDetail />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}
