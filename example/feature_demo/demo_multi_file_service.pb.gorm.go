package example

import (
	context "context"
	gorm "github.com/jinzhu/gorm"
)

type BlogPostServiceDefaultServer struct {
	DB *gorm.DB
}

// Read ...
func (m *BlogPostServiceDefaultServer) Read(ctx context.Context, in *ReadAccountRequest) (*ReadBlogPostsResponse, error) {
	out := &ReadBlogPostsResponse{}
	return out, nil
}
