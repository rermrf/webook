// User Types
export interface User {
  id: number;
  nickname: string;
  email?: string;
  phone?: string;
  avatar?: string;
  aboutMe?: string;
  birthday?: string;
  ctime?: number;
}

export interface UserProfile extends User {
  followeeCount: number;
  followerCount: number;
  isFollowed?: boolean;
}

// Article Types
export interface Article {
  id: number;
  title: string;
  content: string;
  abstract?: string;
  authorId: number;
  authorName?: string;
  status: ArticleStatus;
  ctime: number;
  utime: number;
}

export type ArticleStatus = 'draft' | 'published' | 'withdrawn';

export interface ArticleDetail extends Article {
  author: User;
  interactive: ArticleInteractive;
  tags?: Tag[];
}

export interface ArticleInteractive {
  readCnt: number;
  likeCnt: number;
  collectCnt: number;
  liked: boolean;
  collected: boolean;
}

// Comment Types
export interface Comment {
  id: number;
  content: string;
  userId: number;
  user: User;
  bizId: number;
  rootId?: number;
  parentId?: number;
  replies?: Comment[];
  replyCnt: number;
  ctime: number;
}

// Follow Types
export interface FollowStats {
  followeeCount: number;
  followerCount: number;
}

export interface FollowUser extends User {
  isFollowed: boolean;
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

// Credit Types
export interface CreditBalance {
  balance: number;
  totalEarned: number;
  totalSpent: number;
  monthlyEarned?: number;
  monthlySpent?: number;
}

export interface CreditFlow {
  id: number;
  amount: number;
  type: 'earn' | 'spend';
  description: string;
  bizType?: string;
  ctime: number;
}

// Feed Types
export interface FeedItem {
  id: number;
  type: 'article' | 'follow' | 'like';
  content: Article | User;
  ctime: number;
}

// Ranking Types
export interface RankingItem {
  rank: number;
  article: Article;
  score: number;
}

// API Response Types
export interface ApiResponse<T> {
  code: number;
  msg?: string;
  data?: T;
}

export interface PaginatedResponse<T> {
  list: T[];
  total: number;
  page: number;
  pageSize: number;
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
