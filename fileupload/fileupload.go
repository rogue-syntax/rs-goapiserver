package fileupload

import (
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/gabriel-vasile/mimetype"
	"github.com/rogue-syntax/rs-goapiserver/apireturn/apierrorkeys"
)

const (
	MIME_TYPES_COMMON = "common"
	MIME_TYPES_IMAGE  = "image"
	MIME_TYPES_PDF    = "pdf"
	MIME_TYPES_EXCEL  = "excel"
	MIME_TYPES_WORD   = "word"
	MIME_TYPES_VIDEO  = "video"
	MIME_TYPES_AUDIO  = "audio"
)

var FileExtensionLookups = map[string]map[string]string{
	"image": {
		"image/png":  "png",
		"image/jpeg": "jpg",
		"image/bmp":  "bmp",
		"image/gif":  "gif",
		"image/tiff": "tiff",
	},
	"common": {
		"application/pdf":               "pdf",
		"application/msword":            "doc",
		"application/vnd.ms-excel":      "xls",
		"application/vnd.ms-powerpoint": "ppt",
		"application/zip":               "zip",
		"application/x-rar-compressed":  "rar",
		"application/x-tar":             "tar",
		"application/x-gzip":            "gzip",
		"application/x-bzip2":           "bzip2",
		"application/x-7z-compressed":   "7z",
		"application/x-compressed":      "gz",
		"application/x-zip-compressed":  "zip",
		"application/octet-stream":      "bin",
		"application/json":              "json",
		"application/xml":               "xml",
		"image/png":                     "png",
		"image/jpeg":                    "jpg",
		"image/bmp":                     "bmp",
		"image/gif":                     "gif",
		"image/tiff":                    "tiff",
		"text/plain":                    "txt",
		"text/html":                     "html",
		"text/css":                      "css",
		"text/javascript":               "js",
		"audio/mpeg":                    "mp3",
		"audio/wav":                     "wav",
		"video/mp4":                     "mp4",
		"video/quicktime":               "mov",
		"video/x-msvideo":               "avi",
		"video/x-flv":                   "flv",
		"video/x-matroska":              "mkv",
		"video/webm":                    "webm",
	},
	"pdf": {
		"application/pdf": "pdf",
	},
	"excel": {
		"application/vnd.ms-excel": "xls",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": "xlsx",
		"application/vnd.ms-excel.sheet.macroEnabled.12":                    "xlsm",
	},
	"word": {
		"application/msword": "doc",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": "docx",
	},
	"video": {
		"video/mp4":        "mp4",
		"video/quicktime":  "mov",
		"video/x-msvideo":  "avi",
		"video/x-flv":      "flv",
		"video/x-matroska": "mkv",
		"video/webm":       "webm",
	},
	"audio": {
		"audio/mpeg": "mp3",
		"audio/wav":  "wav",
	},
}

type FileUploadHandlerOptions struct {
	// MaxUploadMB is the maximum upload size in MB
	MaxUploadMB int64
	// FormFieldName is the name of the form field that contains the file ie.e r.FormFile(FormFieldName)
	FormFieldName string
	//Should be like MIME_TYPES_IMAGE ,MIME_TYPES_PDF for FileExtensionLookups[MIME_TYPES_IMAGE], FileExtensionLookups[MIME_TYPES_PDF], etc
	ExpectedMimeTypeList string
}

type FileUploadRaw struct {
	RawData      []byte
	FileName     string
	fileSize     int64
	FileNameBase string
	Extension    string
	FileHeader   *multipart.FileHeader
}

func FileUploadHandler(r *http.Request, opts *FileUploadHandlerOptions) (*FileUploadRaw, error) {

	var FileUploadRaw FileUploadRaw
	// Check options for nil values, etc
	if opts == nil {
		return &FileUploadRaw, errors.New("FileUploadHandlerOptions is nil")
	}
	if opts.ExpectedMimeTypeList == "" {
		return &FileUploadRaw, errors.New("ExpectedMimeTypes is empty")
	}
	if opts.MaxUploadMB == 0 {
		return &FileUploadRaw, errors.New("MaxUploadMB is 0")
	}
	if opts.FormFieldName == "" {
		return &FileUploadRaw, errors.New("FormFieldName is empty")
	}

	maxUploadSize := opts.MaxUploadMB * 1024 * 1024 // MaxUploadMB MB

	err := r.ParseMultipartForm(int64(maxUploadSize))
	if err != nil {
		return &FileUploadRaw, err
	}
	file, headerPtr, err := r.FormFile(opts.FormFieldName) // "myFile" is the key of the input field of your form
	if err != nil {
		return &FileUploadRaw, err
	}
	defer file.Close()
	FileUploadRaw.RawData, err = io.ReadAll(file)
	if err != nil {
		return &FileUploadRaw, err
	}

	FileUploadRaw.FileNameBase = filepath.Base(headerPtr.Filename)
	FileUploadRaw.Extension = filepath.Ext(headerPtr.Filename)
	FileUploadRaw.FileName = headerPtr.Filename
	FileUploadRaw.fileSize = headerPtr.Size
	FileUploadRaw.FileHeader = headerPtr

	_, err = VerifyExtension(&file, FileUploadRaw.Extension, opts.ExpectedMimeTypeList)
	if err != nil {
		return &FileUploadRaw, err
	}

	return &FileUploadRaw, nil

}

// VerifyExtension checks the file extension against the expected MIME types
// It returns the file extension if it matches an entry in FileExtensionsLookup[mimeTypeList], otherwise it returns an error
// The file is rewound to the beginning after the check
//   - file: the file to check
//   - ext: the file extension to check against
//   - mimeTypeList: the list of MIME types to check against, e.g. MIME_TYPES_IMAGE, MIME_TYPES_PDF, etc
func VerifyExtension(file *multipart.File, ext string, mimeTypeList string) (string, error) {
	//check if mime type list exists
	if _, exists := FileExtensionLookups[mimeTypeList]; !exists {
		return "", errors.New(apierrorkeys.MapKeyNotFound + ": File type list not found: " + mimeTypeList)
	}
	(*file).Seek(0, 0)
	mtype, err := mimetype.DetectReader((*file))
	if err != nil {
		return "", errors.New(apierrorkeys.UnauthorizedFileType + ": Could not detect MIME type")
	}

	contentType := mtype.String()
	//remove first char "."
	ext = ext[1:]
	// check if content type is allowed
	if _, exists := FileExtensionLookups[mimeTypeList][contentType]; !exists {
		return "", errors.New(apierrorkeys.UnauthorizedFileType + ": Requested file type is not allowed: " + contentType + " vs " + ext + " " + FileExtensionLookups[mimeTypeList][contentType])
	}

	// check if extension matches expected types
	if ext != FileExtensionLookups[mimeTypeList][contentType] {
		return "", errors.New(apierrorkeys.MismatchedFileType + ": Requested file does not match detected type: " + contentType + " vs " + ext + " " + FileExtensionLookups[mimeTypeList][contentType])
	}

	(*file).Seek(0, 0)
	return FileExtensionLookups[mimeTypeList][contentType], nil

}

/*


func route_uploadBPTFile(w http.ResponseWriter, r *http.Request, usr *user.UserExternal) {
	r.ParseMultipartForm(250 << 20)

	_id, convErr := strconv.Atoi(r.FormValue("_a_id"))
	if convErr != nil {
		fmt.Fprintf(w, `{"msg":"error", "route":"uploadReport", "action":"nil", "displayMsg":"`+convErr.Error()+`" }`)
		return
	}

	_batch_id, convErr := strconv.Atoi(r.FormValue("_batch_id"))
	if convErr != nil {
		fmt.Fprintf(w, `{"msg":"error", "route":"uploadReport", "action":"nil", "displayMsg":"`+convErr.Error()+`" }`)
		return
	}
	_c_id := (*usr)._c_id

	file, handler, fErr := r.FormFile("file")
	if fErr != nil {
		fmt.Fprintf(w, `{"msg":"error", "route":"uploadReport", "action":"nil", "displayMsg":"`+fErr.Error()+`" }`)
		return
	}
	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Fprintf(w, `{"msg":"error", "route":"uploadReport", "action":"nil", "displayMsg":"`+err.Error()+`" }`)
		return
	}

	fLoc := "/var/www/uploads/" + strconv.Itoa(_c_id) + "/"
	if _, err := os.Stat(fLoc); os.IsNotExist(err) {
		err := os.Mkdir(fLoc, 0777)
		if err != nil {
			log.Print(err.Error())
		}
	}

	fLoc = "/var/www/uploads/" + strconv.Itoa(_c_id) + "/uploaded/"
	if _, err := os.Stat(fLoc); os.IsNotExist(err) {
		err := os.Mkdir(fLoc, 0777)
		if err != nil {
			log.Print(err.Error())
		}
	}

	timeStamp := time.Now().Unix()
	timeStampString := fmt.Sprint(timeStamp)

	fName := handler.Filename
	fName = strings.ReplaceAll(fName, " ", "")
	fName = timeStampString + "_" + fName

		fileNameSli := strings.Split(fName, ".")
		fileNameSli = fileNameSli[:len(fName)-1]
		fNameJoin := strings.Join(fileNameSli, ".")

			fileNameSli := strings.Split(fName, ".")
			fileExt := fileNameSli[len(fileNameSli)-1]

			if len(fileNameSli) > 0 {
				fileNameSli = fileNameSli[:len(fName)-1]
			}

			fNameJoin := strings.Join(fileNameSli, ".")
			fNameJoin = strings.ReplaceAll(fNameJoin, " ", "")

	//fName = strings.ReplaceAll(fName, " ", "")

	//fn := fLoc + fNameJoin + "_" + timeStampString + "." + fileExt
	fn := fLoc + fName

	wErr := ioutil.WriteFile(fn, fileBytes, 0755)
	if wErr != nil {
		fmt.Fprintf(w, `{"msg":"error", "route":"uploadReport", "action":"nil", "displayMsg":"`+wErr.Error()+`" }`)
		return
	}

	contentType, err := mimetype.DetectFile(fn)
	if err != nil {
		fmt.Fprintf(w, `{"msg":"error", "route":"uploadReport", "action":"nil", "displayMsg":"`+err.Error()+`" }`)
		return
	}

	if (*contentType).Extension() == ".pdf" || (*contentType).Extension() == ".png" || (*contentType).Extension() == ".bmp" || (*contentType).Extension() == ".jpg" || (*contentType).Extension() == ".gif" || (*contentType).Extension() == ".tiff" {

		fLoc = fLoc + fName

		qRows, upErr := DB.Query("call saveBPTFile( ?, ?, ?, ?, ?, ?, ?)", _id, fName, fLoc, timeStamp, usr.uID, _c_id, _batch_id)
		if upErr != nil {
			fmt.Fprintf(w, `{"msg":"error", "route":"uploadReport", "action":"nil", "displayMsg":"`+upErr.Error()+`" }`)
			qRows.Close()
			return
		}
		defer qRows.Close()

		fmt.Fprintf(w, `{"msg":"success", "route":"uploadFile", "action":"nil", "displayMsg":"Upload Successful" }`)
		return
	} else {
		e := os.Remove(fLoc)
		if e != nil {
			fmt.Fprintf(w, `{"msg":"error", "route":"uploadW9", "action":"nil", "displayMsg":"`+e.Error()+`" }`)
			return
		}
		fmt.Fprintf(w, `{"msg":"error", "route":"uploadW9", "action":"nil", "displayMsg":"Incorrect file type - file type should be image or PDF" }`)
		return
	}

}
*/
