// User Types — matches backend Profile struct (user.go)
// Backend json tags: id, email, phone, nickname, aboutMe, birthday, avatar_url, ctime
export interface User {
  id: number;
  nickname: string;
  email?: string;
  phone?: string;
  avatar_url?: string;
  aboutMe?: string;
  birthday?: string;
  ctime?: string;
}

// Backend PublicProfile struct (user.go)
// json tags: id, nickname, aboutMe, follower_count, following_count, article_count, avatar_url, ctime
export interface UserProfile extends User {
  follower_count: number;
  following_count: number;
  article_count: number;
}

// Article Types — matches backend ArticleVO struct (article_vo.go)
// Backend json tags: id, title, abstract, content, author_id, author_name, cover_url, author_avatar_url, status, read_cnt, like_cnt, collect_cnt, liked, collected, author_followed, ctime, utime
export interface Article {
  id: number;
  title: string;
  content: string;
  abstract?: string;
  author_id: number;
  author_name?: string;
  author_avatar_url?: string;
  cover_url?: string;
  status: number;
  ctime: string;
  utime: string;
}

export interface ArticleDetail extends Article {
  // Flat fields from backend ArticleVO
  read_cnt: number;
  like_cnt: number;
  collect_cnt: number;
  liked: boolean;
  collected: boolean;
  author_followed: boolean;
}

// Comment Types — matches backend Comment struct (article_vo.go)
// Backend json tags: id, content, uid, user_name, user_avatar_url, parent_id, root_id, ctime
export interface Comment {
  id: number;
  content: string;
  uid: number;
  user_name: string;
  user_avatar_url?: string;
  parent_id: number;
  root_id: number;
  ctime: number;
}

// Backend GetCommentResp: { Comments: Comment[] }
export interface GetCommentResp {
  Comments: Comment[];
}

// Follow Types — matches backend GetFollowStatic response (follow.go)
// Backend json tags: followees, followers
export interface FollowStats {
  followees: number;
  followers: number;
}

// FollowUserVO from backend (follow.go): id, nickname, about_me, followed
export interface FollowUser {
  id: number;
  nickname: string;
  about_me: string;
  followed: boolean;
}

// Tag Types
export interface Tag {
  id: number;
  name: string;
  uid: number;
}

// Notification Types
export interface Notification {
  id: number;
  type: NotificationType;
  title: string;
  content: string;
  senderId?: number;
  sender?: User;
  bizId?: number;
  bizType?: string;
  isRead: boolean;
  ctime: number;
}

export type NotificationType = 'system' | 'like' | 'comment' | 'follow' | 'reward';

export interface UnreadCount {
  total: number;
  byType: Record<NotificationType, number>;
}

// Credit Types — matches backend CreditFlowVO (credit.go)
// Backend json tags: id, biz, biz_id, change_amt, balance, description, ctime
export interface CreditFlow {
  id: number;
  biz: string;
  biz_id: number;
  change_amt: number;
  balance: number;
  description: string;
  ctime: number;
}

// Backend DailyStatusVO (credit.go): biz, earned_count, earned_amt, daily_limit, remaining
export interface DailyStatus {
  biz: string;
  earned_count: number;
  earned_amt: number;
  daily_limit: number;
  remaining: number;
}

// Feed Types — matches backend FeedEventVO (feed.go)
// Backend json tags: id, user_id, type, content, article, ctime
export interface FeedEvent {
  id: number;
  user_id: number;
  type: string;
  content: string;
  article?: FeedArticle;
  ctime: number;
}

// Backend FeedArticleVO (feed.go): id, title, abstract, author_id, author_name, tags, like_cnt, comment_cnt, read_cnt, ctime
export interface FeedArticle {
  id: number;
  title: string;
  abstract: string;
  author_id: number;
  author_name: string;
  tags: string[];
  like_cnt: number;
  comment_cnt: number;
  read_cnt: number;
  ctime: string;
}

// Ranking Types — matches backend HotArticleVO (ranking.go)
// Backend json tags: id, title, author_id, author_name, like_cnt, comment_cnt, read_cnt, ctime
export interface HotArticle {
  id: number;
  title: string;
  author_id: number;
  author_name: string;
  like_cnt: number;
  comment_cnt: number;
  read_cnt: number;
  ctime: string;
}

// Search Types — matches backend SearchResponse (search.go)
// Backend json: { users: SearchUser[], articles: SearchArticle[] }
export interface SearchResponse {
  users: SearchUserResult[];
  articles: SearchArticleResult[];
}

// Backend SearchUser: id, nickname, about_me, follower_count
export interface SearchUserResult {
  id: number;
  nickname: string;
  about_me: string;
  follower_count: number;
}

// Backend SearchArticle: id, title, abstract, author_id, author_name, tags, like_cnt, ctime
export interface SearchArticleResult {
  id: number;
  title: string;
  abstract: string;
  author_id: number;
  author_name: string;
  tags: string[];
  like_cnt: number;
  ctime: string;
}

// Recommend user — matches backend RecommendUserVO (user.go)
// Backend json tags: id, nickname, about_me, article_count, follower_count, avatar_url, followed
export interface RecommendUser {
  id: number;
  nickname: string;
  about_me: string;
  article_count: number;
  follower_count: number;
  avatar_url: string;
  followed: boolean;
}

// Hot keyword — matches backend HotKeyword (search.go)
// Backend json: { keyword: string }
export interface HotKeyword {
  keyword: string;
}

// API Response Types — matches backend ginx.Result
// Backend json tags: code, msg, data
export interface ApiResponse<T> {
  code: number;
  msg?: string;
  data?: T;
}

// Auth Types
export interface LoginRequest {
  email: string;
  password: string;
}

export interface SignupRequest {
  email: string;
  password: string;
  confirmPassword: string;
  nickname: string;
}

export interface SmsLoginRequest {
  phone: string;
  code: string;
}

// OpenAPI Types
export type AppType = 1 | 2 | 3; // 1=OAuth2, 2=API, 3=Both
export type AppStatus = 0 | 1 | 2 | 3; // 0=Pending, 1=Approved, 2=Rejected, 3=Disabled

export interface OpenAPIApp {
  appId: string;
  name: string;
  type: AppType;
  status: AppStatus;
  statusText: string;
  description: string;
  homepageURL: string;
  logoURL: string;
  redirectURIs: string[];
  scopes: string[];
  callbackURL: string;
  ipWhitelist: string[];
  rejectReason?: string;
  createdAt: number;
  updatedAt: number;
}

export interface CreateAppRequest {
  name: string;
  type: AppType;
  description: string;
  homepageURL: string;
  logoURL?: string;
  redirectURIs: string[];
  scopes: string[];
  callbackURL?: string;
  ipWhitelist?: string[];
}

export interface CreateAppResponse {
  appId: string;
  appSecret: string;
  name: string;
}

export interface UpdateAppRequest {
  name: string;
  description: string;
  homepageURL: string;
  logoURL?: string;
  redirectURIs: string[];
  scopes: string[];
  callbackURL?: string;
  ipWhitelist?: string[];
}

export interface APILog {
  id: number;
  endpoint: string;
  method: string;
  status: number;
  ip: string;
  duration: number;
  createdAt: number;
}

export interface AppAuthorization {
  appId: string;
  appName: string;
  scopes: string[];
  createdAt: number;
  expiresAt: number;
}

// Available OAuth Scopes
export const OAUTH_SCOPES = {
  profile: { name: 'profile', label: '基本资料', description: '读取用户基本资料（昵称、头像）' },
  email: { name: 'email', label: '邮箱', description: '读取用户邮箱地址' },
  openid: { name: 'openid', label: 'OpenID', description: 'OpenID Connect 标识' },
  offline: { name: 'offline', label: '离线访问', description: '获取刷新令牌，长期访问' },
  'read:credit': { name: 'read:credit', label: '读取积分', description: '查看用户积分余额' },
  'write:credit': { name: 'write:credit', label: '扣减积分', description: '扣减用户积分（需审核）' },
  'read:article': { name: 'read:article', label: '读取文章', description: '读取用户文章列表' },
  'write:article': { name: 'write:article', label: '发布文章', description: '以用户身份发布文章' },
} as const;
