package gongular

import (
	"mime/multipart"
)

// UploadedFile packs the file and the header of the file to one struct
type UploadedFile struct {
	File   multipart.File
	Header *multipart.FileHeader
}
