package service

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"webook/article/domain"
	"webook/article/repository"
	repomocks "webook/article/repository/mocks"
	"webook/pkg/logger"
)

func Test_articleService_Publish(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) (repository.ArticleAuthorRepository, repository.ArticleReaderRepository)
		art     domain.Article
		wantErr error
		wantId  int64
	}{
		{
			name: "新建发表成功",
			mock: func(ctrl *gomock.Controller) (repository.ArticleAuthorRepository, repository.ArticleReaderRepository) {
				author := repomocks.NewMockArticleAuthorRepository(ctrl)
				author.EXPECT().Create(gomock.Any(), domain.Article{
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(1), nil)
				reader := repomocks.NewMockArticleReaderRepository(ctrl)
				reader.EXPECT().Save(gomock.Any(), domain.Article{
					// 确保使用制作库id
					Id:      1,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(1), nil)
				return author, reader
			},

			art: domain.Article{
				// 新建帖子并发表没有id
				Title:   "我的标题",
				Content: "我的内容",
				Author: domain.Author{
					Id: 123,
				},
			},

			wantErr: nil,
			wantId:  1,
		},
		{
			name: "修改并发表成功",
			mock: func(ctrl *gomock.Controller) (repository.ArticleAuthorRepository, repository.ArticleReaderRepository) {
				author := repomocks.NewMockArticleAuthorRepository(ctrl)
				author.EXPECT().Update(gomock.Any(), domain.Article{
					Id:      2,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(nil)
				reader := repomocks.NewMockArticleReaderRepository(ctrl)
				reader.EXPECT().Save(gomock.Any(), domain.Article{
					// 确保使用制作库id
					Id:      2,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(2), nil)
				return author, reader
			},

			art: domain.Article{
				Id:      2,
				Title:   "我的标题",
				Content: "我的内容",
				Author: domain.Author{
					Id: 123,
				},
			},

			wantErr: nil,
			wantId:  2,
		},
		{
			name: "新建保存到制作库失败",
			mock: func(ctrl *gomock.Controller) (repository.ArticleAuthorRepository, repository.ArticleReaderRepository) {
				author := repomocks.NewMockArticleAuthorRepository(ctrl)
				author.EXPECT().Create(gomock.Any(), domain.Article{
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(0), errors.New("保存到制作库失败"))
				reader := repomocks.NewMockArticleReaderRepository(ctrl)
				return author, reader
			},

			art: domain.Article{
				// 新建帖子并发表没有id
				Title:   "我的标题",
				Content: "我的内容",
				Author: domain.Author{
					Id: 123,
				},
			},

			wantErr: errors.New("保存到制作库失败"),
			wantId:  0,
		},
		{
			name: "修改保存到制作库失败",
			mock: func(ctrl *gomock.Controller) (repository.ArticleAuthorRepository, repository.ArticleReaderRepository) {
				author := repomocks.NewMockArticleAuthorRepository(ctrl)
				author.EXPECT().Update(gomock.Any(), domain.Article{
					Id:      2,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(errors.New("保存到制作库失败"))
				reader := repomocks.NewMockArticleReaderRepository(ctrl)
				return author, reader
			},

			art: domain.Article{
				Id:      2,
				Title:   "我的标题",
				Content: "我的内容",
				Author: domain.Author{
					Id: 123,
				},
			},

			wantErr: errors.New("保存到制作库失败"),
			wantId:  0,
		},
		{
			name: "新建保存到制作库成功，保存到线上库重试成功",
			mock: func(ctrl *gomock.Controller) (repository.ArticleAuthorRepository, repository.ArticleReaderRepository) {
				author := repomocks.NewMockArticleAuthorRepository(ctrl)
				author.EXPECT().Create(gomock.Any(), domain.Article{
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(1), nil)
				reader := repomocks.NewMockArticleReaderRepository(ctrl)
				reader.EXPECT().Save(gomock.Any(), domain.Article{
					// 确保使用制作库id
					Id:      1,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(0), errors.New("保存到线上库失败"))
				reader.EXPECT().Save(gomock.Any(), domain.Article{
					// 确保使用制作库id
					Id:      1,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(1), nil)
				return author, reader
			},

			art: domain.Article{
				// 新建帖子并发表没有id
				Title:   "我的标题",
				Content: "我的内容",
				Author: domain.Author{
					Id: 123,
				},
			},

			wantErr: nil,
			wantId:  1,
		},
		{
			// 部分失败
			// 制作库和线上库的事务问题
			// 1. 如果使用了关系型数据库，也分：同库不同表、不同库
			// 2. 使用了非关系型数据库
			// 3. service 这一层既不适合开事物，也不一定就能开事物
			name: "修改保存到制作库成功，保存到线上库重试成功",
			mock: func(ctrl *gomock.Controller) (repository.ArticleAuthorRepository, repository.ArticleReaderRepository) {
				author := repomocks.NewMockArticleAuthorRepository(ctrl)
				author.EXPECT().Update(gomock.Any(), domain.Article{
					Id:      2,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(nil)
				reader := repomocks.NewMockArticleReaderRepository(ctrl)
				reader.EXPECT().Save(gomock.Any(), domain.Article{
					// 确保使用制作库id
					Id:      2,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(0), errors.New("线上库保存失败"))
				reader.EXPECT().Save(gomock.Any(), domain.Article{
					// 确保使用制作库id
					Id:      2,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(2), nil)
				return author, reader
			},

			art: domain.Article{
				Id:      2,
				Title:   "我的标题",
				Content: "我的内容",
				Author: domain.Author{
					Id: 123,
				},
			},

			wantErr: nil,
			wantId:  2,
		},
		{
			name: "新建保存到制作库成功，保存到线上库重试全部失败",
			mock: func(ctrl *gomock.Controller) (repository.ArticleAuthorRepository, repository.ArticleReaderRepository) {
				author := repomocks.NewMockArticleAuthorRepository(ctrl)
				author.EXPECT().Create(gomock.Any(), domain.Article{
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(1), nil)
				reader := repomocks.NewMockArticleReaderRepository(ctrl)
				reader.EXPECT().Save(gomock.Any(), domain.Article{
					// 确保使用制作库id
					Id:      1,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Times(3).Return(int64(0), errors.New("保存到线上库失败"))
				return author, reader
			},

			art: domain.Article{
				// 新建帖子并发表没有id
				Title:   "我的标题",
				Content: "我的内容",
				Author: domain.Author{
					Id: 123,
				},
			},

			wantErr: errors.New("保存到线上库失败"),
			wantId:  0,
		},
		{
			// 制作库和线上库的事务问题
			// 1. 如果使用了关系型数据库，也分：同库不同表、不同库
			// 2. 使用了非关系型数据库
			// 3. service 这一层既不适合开事物，也不一定就能开事物
			name: "修改保存到制作库成功，但是保存到线上库重试失败",
			mock: func(ctrl *gomock.Controller) (repository.ArticleAuthorRepository, repository.ArticleReaderRepository) {
				author := repomocks.NewMockArticleAuthorRepository(ctrl)
				author.EXPECT().Update(gomock.Any(), domain.Article{
					Id:      2,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(nil)
				reader := repomocks.NewMockArticleReaderRepository(ctrl)
				reader.EXPECT().Save(gomock.Any(), domain.Article{
					// 确保使用制作库id
					Id:      2,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Times(3).Return(int64(0), errors.New("线上库保存失败"))
				return author, reader
			},

			art: domain.Article{
				Id:      2,
				Title:   "我的标题",
				Content: "我的内容",
				Author: domain.Author{
					Id: 123,
				},
			},

			wantErr: errors.New("线上库保存失败"),
			wantId:  0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			author, reader := tc.mock(ctrl)
			svc := NewArticleServiceV1(author, reader, logger.NopLogger{})
			id, err := svc.PublishV1(context.Background(), tc.art)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantId, id)
		})
	}
}
