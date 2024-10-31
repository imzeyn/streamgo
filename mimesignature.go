package streamgo

type MimeSignature struct {
	Type      string
	Signature []byte
	Category  MimeCategory
	Extensions []string
}

type MimeSignatureList []MimeSignature

type MimeCategory string

const (
	MimeCategoryImage    MimeCategory = "image"
	MimeCategoryVideo    MimeCategory = "video"
	MimeCategoryAudio    MimeCategory = "audio"
	MimeCategoryDocument MimeCategory = "document"
)

var MimeDefaultSignatures = MimeSignatureList{
	{"image/jpeg", []byte{0xFF, 0xD8, 0xFF}, MimeCategoryImage, []string{"jpg", "jpeg"}},
	{"image/png", []byte{0x89, 0x50, 0x4E, 0x47}, MimeCategoryImage, []string{"png"}},
	{"image/gif", []byte{0x47, 0x49, 0x46, 0x38}, MimeCategoryImage, []string{"gif"}},
	{"image/bmp", []byte{0x42, 0x4D}, MimeCategoryImage, []string{".bmp"}},
	{"image/tiff", []byte{0x49, 0x49, 0x2A, 0x00}, MimeCategoryImage, []string{"tiff", "tif"}},
	{"image/tiff", []byte{0x4D, 0x4D, 0x00, 0x2A}, MimeCategoryImage, []string{"tiff", "tif"}},
	{"image/webp", []byte{0x52, 0x49, 0x46, 0x46}, MimeCategoryImage, []string{"webp"}},
	{"image/x-icon", []byte{0x00, 0x00, 0x01, 0x00}, MimeCategoryImage, []string{"ico"}},
	{"image/heic", []byte{0x00, 0x00, 0x00, 0x18}, MimeCategoryImage, []string{"heic"}},
	{"image/heic", []byte{0x66, 0x74, 0x79, 0x70}, MimeCategoryImage, []string{"heic"}},
	{"image/svg+xml", []byte{0x3C, 0x3F, 0x78, 0x6D}, MimeCategoryImage, []string{"svg"}}, 

	{"video/mp4", []byte{0x00, 0x00, 0x00, 0x18, 0x66, 0x74, 0x79, 0x70}, MimeCategoryVideo, []string{"mp4"}},
	{"video/avi", []byte{'R', 'I', 'F', 'F'}, MimeCategoryVideo, []string{"avi"}},
	{"video/mpeg", []byte{0x00, 0x00, 0x01, 0xBA}, MimeCategoryVideo, []string{"mpeg", "mpg"}},
	{"video/quicktime", []byte{0x00, 0x00, 0x00, 0x18, 0x66, 0x74, 0x79, 0x70}, MimeCategoryVideo, []string{"mov"}},
	{"video/x-msvideo", []byte{'R', 'I', 'F', 'F'}, MimeCategoryVideo, []string{"avi"}},
	{"video/x-matroska", []byte{0x1A, 0x45, 0xDF, 0xA3}, MimeCategoryVideo, []string{"mkv"}},
	{"video/x-flv", []byte{0x46, 0x4C, 0x56}, MimeCategoryVideo, []string{"flv"}},
	{"video/webm", []byte{0x1A, 0x45, 0xDF, 0xA3}, MimeCategoryVideo, []string{"webm"}},

	{"audio/mpeg", []byte{0xFF, 0xFB}, MimeCategoryAudio, []string{"mp3"}},
	{"audio/wav", []byte{'R', 'I', 'F', 'F'}, MimeCategoryAudio, []string{"wav"}},
	{"audio/flac", []byte{'f', 'L', 'A', 'C'}, MimeCategoryAudio, []string{"flac"}},
	{"audio/aac", []byte{0xFF, 0xF1}, MimeCategoryAudio, []string{"aac"}},
	{"audio/ogg", []byte{0x4F, 0x67, 0x67, 0x53}, MimeCategoryAudio, []string{"ogg"}},
	{"audio/webm", []byte{0x1A, 0x45, 0xDF, 0xA3}, MimeCategoryAudio, []string{"webm"}},

	{"application/pdf", []byte{0x25, 0x50, 0x44, 0x46}, MimeCategoryDocument, []string{"pdf"}},
	{"application/msword", []byte{0xD0, 0xCF, 0x11, 0xE0}, MimeCategoryDocument, []string{"doc"}},
	{"application/vnd.openxmlformats-officedocument.wordprocessingml.document", []byte{0x50, 0x4B, 0x03, 0x04}, MimeCategoryDocument, []string{"docx"}},
	{"application/vnd.ms-excel", []byte{0xD0, 0xCF, 0x11, 0xE0}, MimeCategoryDocument, []string{"xls"}},
	{"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", []byte{0x50, 0x4B, 0x03, 0x04}, MimeCategoryDocument, []string{"xlsx"}},
	{"application/rtf", []byte{0x7B, 0x5C, 0x72, 0x74}, MimeCategoryDocument, []string{"rtf"}},
	{"text/plain", []byte{0xEF, 0xBB, 0xBF}, MimeCategoryDocument, []string{"txt"}},

	{"application/zip", []byte{0x50, 0x4B, 0x03, 0x04}, MimeCategoryDocument, []string{"zip"}},
	{"application/x-rar-compressed", []byte{0x52, 0x61, 0x72, 0x21}, MimeCategoryDocument, []string{"rar"}},

}

func (m *MimeSignatureList) GetByCategorys(name MimeCategory) MimeSignatureList{
	list := MimeSignatureList{}
	for _, v := range *m {
		if v.Category == name {
			list = append(list, v)
		}
	}
	return list
}

func (m *MimeSignatureList) GetByExtension(name string) MimeSignatureList{
	list := MimeSignatureList{}
	for _, v := range *m {
		for _, ext := range v.Extensions {
			if ext == name {
				list = append(list, v)
			}
		}
	}

	return list
}