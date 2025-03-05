package streamgo

import "strings"

type MimeSignature struct {
	Type       string
	Signature  string
	Extensions string
	Category   MimeCategory
}

type MimeSignatureList []MimeSignature

type MimeCategory byte

const (
	MimeCategoryImage MimeCategory = iota
	MimeCategoryVideo
	MimeCategoryAudio
	MimeCategoryDocument
)

var MimeDefaultSignatures = MimeSignatureList{
	{"image/jpeg", "\xFF\xD8\xFF", "jpg,jpeg", MimeCategoryImage},
	{"image/png", "\x89PNG", "png", MimeCategoryImage},
	{"image/gif", "GIF8", "gif", MimeCategoryImage},
	{"image/bmp", "BM", ".bmp", MimeCategoryImage},
	{"image/tiff", "\x49\x49\x2A\x00", "tiff,tif", MimeCategoryImage},
	{"image/tiff", "\x4D\x4D\x00\x2A", "tiff,tif", MimeCategoryImage},
	{"image/webp", "RIFF", "webp", MimeCategoryImage},
	{"image/x-icon", "\x00\x00\x01\x00", "ico", MimeCategoryImage},
	{"image/heic", "\x00\x00\x00\x18", "heic", MimeCategoryImage},
	{"image/heic", "ftyp", "heic", MimeCategoryImage},
	{"image/svg+xml", "<?xm", "svg", MimeCategoryImage},

	{"video/mp4", "\x00\x00\x00\x18ftyp", "mp4", MimeCategoryVideo},
	{"video/avi", "RIFF", "avi", MimeCategoryVideo},
	{"video/mpeg", "\x00\x00\x01\xBA", "mpeg,mpg", MimeCategoryVideo},
	{"video/quicktime", "\x00\x00\x00\x18ftyp", "mov", MimeCategoryVideo},
	{"video/x-msvideo", "RIFF", "avi", MimeCategoryVideo},
	{"video/x-matroska", "\x1A\x45\xDF\xA3", "mkv", MimeCategoryVideo},
	{"video/x-flv", "FLV", "flv", MimeCategoryVideo},
	{"video/webm", "\x1A\x45\xDF\xA3", "webm", MimeCategoryVideo},

	{"audio/mpeg", "\xFF\xFB", "mp3", MimeCategoryAudio},
	{"audio/wav", "RIFF", "wav", MimeCategoryAudio},
	{"audio/flac", "fLaC", "flac", MimeCategoryAudio},
	{"audio/aac", "\xFF\xF1", "aac", MimeCategoryAudio},
	{"audio/ogg", "OggS", "ogg", MimeCategoryAudio},
	{"audio/webm", "\x1A\x45\xDF\xA3", "webm", MimeCategoryAudio},

	{"application/pdf", "%PDF", "pdf", MimeCategoryDocument},
	{"application/msword", "\xD0\xCF\x11\xE0", "doc", MimeCategoryDocument},
	{"application/vnd.openxmlformats-officedocument.wordprocessingml.document", "PK\x03\x04", "docx", MimeCategoryDocument},
	{"application/vnd.ms-excel", "\xD0\xCF\x11\xE0", "xls", MimeCategoryDocument},
	{"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", "PK\x03\x04", "xlsx", MimeCategoryDocument},
	{"application/rtf", "{\\rt", "rtf", MimeCategoryDocument},
	{"text/plain", "\xEF\xBB\xBF", "txt", MimeCategoryDocument},
	{"application/zip", "PK\x03\x04", "zip", MimeCategoryDocument},
	{"application/x-rar-compressed", "Rar!", "rar", MimeCategoryDocument},
}

func (m *MimeSignatureList) GetByCategorys(name MimeCategory) *MimeSignatureList {
	list := make(MimeSignatureList, 0, len(*m)/2) // Preallocate
	for _, v := range *m {
		if v.Category == name {
			list = append(list, v)
		}
	}
	return &list
}

func (m *MimeSignatureList) GetByExtension(name string) *MimeSignatureList {
	list := make(MimeSignatureList, 0, 3) // Preallocate small initial capacity
	for _, v := range *m {
		for _, ext := range strings.Split(v.Extensions, ",") {
			if ext == name {
				list = append(list, v)
				break
			}
		}
	}
	return &list
}
