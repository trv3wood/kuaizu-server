package service

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/trv3wood/kuaizu-server/internal/oss"
)

const maxFileSize = 5 * 1024 * 1024 // 5MB

var allowedExts = map[string]bool{".jpg": true, ".jpeg": true, ".png": true}

// CommonsService handles common utilities like file upload.
type CommonsService struct {
	ossClient *oss.Client
}

// NewCommonsService creates a new CommonsService.
func NewCommonsService(ossClient *oss.Client) *CommonsService {
	return &CommonsService{ossClient: ossClient}
}

// FileUploadResult is returned on successful upload.
type FileUploadResult struct {
	URL string
}

// UploadFile validates and uploads a multipart file to OSS.
func (s *CommonsService) UploadFile(file multipart.File, header *multipart.FileHeader) (*FileUploadResult, error) {
	if header.Size > maxFileSize {
		return nil, ErrBadRequest(fmt.Sprintf("文件大小超过限制 (最大 %dMB)", maxFileSize/1024/1024))
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedExts[ext] {
		return nil, ErrBadRequest("不支持的文件类型，仅支持 JPG 和 PNG")
	}

	filename := uuid.New().String() + ext
	result, err := s.ossClient.Upload(file, filename)
	if err != nil {
		return nil, ErrInternal("文件上传失败")
	}

	return &FileUploadResult{URL: result.URL}, nil
}
