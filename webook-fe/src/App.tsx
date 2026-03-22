import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { MobileLayout } from './components/MobileLayout'
import Login from './pages/Login'
import Register from './pages/Register'
import Home from './pages/Home'
import ArticleDetail from './pages/ArticleDetail'
import Profile from './pages/Profile'
import Messages from './pages/Messages'
import Search from './pages/Search'

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
          <Route path="/messages" element={<Messages />} />
          <Route path="/search" element={<Search />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}
