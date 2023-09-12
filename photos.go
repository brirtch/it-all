package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rwcarlsen/goexif/exif"
	"golang.org/x/image/draw"
)

type PhotoFolderRequest struct {
	Folder string `json:folder`
}

type PhotoResponse struct {
	PhotoYears []*PhotoYear `json:"years"`
}

type PhotoYear struct {
	Year   int           `json:"year"`
	Months []*PhotoMonth `json:"months"`
}

type PhotoMonth struct {
	Month string      `json:"month"`
	Days  []*PhotoDay `json:"days"`
}

type PhotoDay struct {
	Day    int   `json:"day"`
	Photos []int `json:"photos"`
}

type PhotoInfoResponse struct {
	Bytes    int    `json:"bytes"`
	Filename string `json:"filename"`
}

// Creates a thumbnail of sourceFile (JPG only at the moment)
// and returns the thumbnail path.
func createThumbnail(sourceFile string) (string, error) {
	input, _ := os.Open(sourceFile)
	defer input.Close()

	filename := filepath.Base(sourceFile)
	destFile := "data" + string(os.PathSeparator) + filename + ".thumb.jpg"
	output, _ := os.Create(destFile)
	defer output.Close()

	// Decode the image (from PNG to image.Image):
	src, err := jpeg.Decode(input)
	if err != nil {
		// Likely corrupt file. Flag it to be ignored.
		return "", err
	}

	// Set the expected size that you want:
	//dst := image.NewRGBA(image.Rect(0, 0, src.Bounds().Max.X/2, src.Bounds().Max.Y/2))
	landscape := src.Bounds().Max.X > src.Bounds().Max.Y
	targetWidth := 600
	targetHeight := 600
	scaleRatio := 1.0
	if landscape {
		scaleRatio = float64(targetWidth) / float64(src.Bounds().Max.X)
	} else {
		scaleRatio = float64(targetHeight) / float64(src.Bounds().Max.Y)
	}
	dst := image.NewRGBA(image.Rect(0, 0, int(float64(src.Bounds().Max.X)*scaleRatio), int(float64(src.Bounds().Max.Y)*scaleRatio)))

	// Resize:
	draw.NearestNeighbor.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)

	// Encode to `output`:
	jpeg.Encode(output, dst, nil)

	return destFile, nil
}

func getDateTaken(fname string) time.Time {
	f, err := os.Open(fname)
	if err != nil {
		log.Fatal(err)
	}

	x, err := exif.Decode(f)
	if err != nil {
		return time.Date(1900, 01, 01, 0, 0, 0, 0, time.Local)
	}

	dateTaken, err := x.Get(exif.DateTime)
	if err != nil {
		return time.Date(1900, 01, 01, 0, 0, 0, 0, time.Local)
	}

	dateTakenStr, err := dateTaken.StringVal()
	if err != nil {
		return time.Date(1900, 01, 01, 0, 0, 0, 0, time.Local)
	}

	layout := "2006:01:02 15:04:05"
	myDate, err := time.Parse(layout, dateTakenStr)
	if err != nil {
		return time.Date(1900, 01, 01, 0, 0, 0, 0, time.Local)
	}

	return myDate
}

func PhotoWatch() {

	// Initially, we'll scan on request, those folders that are watched.
	// We need to do a recursive directory listing, and record the names of photos
	// in a database table, generate a thumbnail.
	for 1 == 1 {
		sql := `SELECT rowid,folder FROM photo_folder WHERE state = 'PENDING_SCAN'`
		rows, err := DB.Query(sql)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		var (
			id     string
			folder string
		)

		var folders []string

		for rows.Next() {
			err := rows.Scan(&id, &folder)
			if err != nil {
				log.Fatal(err)
			}
			folders = append(folders, folder)

		}
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}

		for _, folder := range folders {
			err = filepath.Walk(folder,
				func(thePath string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}

					if strings.ToLower(filepath.Ext(thePath)) == ".jpg" {
						// Check if this image already exists in the photo table. If so, ignore it.
						existsSQL := `SELECT COUNT(*) AS photo_count FROM photo WHERE full_path = ?`
						row := DB.QueryRow(existsSQL, thePath)
						photoCount := 0
						row.Scan(&photoCount)

						if photoCount == 0 {
							thumbnailPath, err := createThumbnail(thePath)
							if err == nil {
								dateTaken := getDateTaken(thePath)
								fmt.Println(dateTaken)
								pathElements := strings.Split(thePath, string(os.PathSeparator))
								sql := `INSERT INTO photo(full_path,filename,bytes,parent_folder,date_taken,thumbnail) VALUES(?,?,?,?,?,?)`

								parentFolder := ""
								if len(pathElements) > 2 {
									parentFolder = pathElements[len(pathElements)-2]
								} else {
									parentFolder = ""
								}
								_, err = DB.Exec(sql, thePath, filepath.Base(thePath), info.Size(), parentFolder, dateTaken, thumbnailPath)
								if err != nil {
									log.Println(err)
								}

								fmt.Println(thePath, info.Size())
							}

						}
					}

					return nil
				})
			if err != nil {
				log.Println(err)
			}
			sql := `UPDATE photo_folder SET state = 'SCANNED' WHERE rowid = ?`
			_, err = DB.Exec(sql, id)
		}

		time.Sleep(1 * time.Minute)
	}

}

// GET /photos/{photoID}
func ImageServeHandler(w http.ResponseWriter, r *http.Request) {
	photoID := chi.URLParam(r, "photoID")
	sql := `SELECT full_path FROM photo WHERE photo_id = ?`
	row := DB.QueryRow(sql, photoID)

	fullPath := ""
	_ = row.Scan(&fullPath)
	http.ServeFile(w, r, fullPath)
}

// POST /photos/folder
func NewPhotoFolderHandler(w http.ResponseWriter, r *http.Request) {
	var photoFolderRequest PhotoFolderRequest
	err := json.NewDecoder(r.Body).Decode(&photoFolderRequest)
	if err != nil {
		// If the structure of the body is wrong, return an HTTP error
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println(err)
		return
	}
	AddPhotoFolder(photoFolderRequest.Folder)
}

func AddPhotoFolder(folder string) {
	sql := `INSERT INTO photo_folder(folder, date_added, date_last_scanned, photo_count, state) VALUES (?,CURRENT_TIMESTAMP,NULL,NULL, 'PENDING_SCAN');`
	//sql = strings.Replace(sql, "?", "'"+folder+"'", -1)
	DB.Exec(sql, folder)
	/*preState, err := db.Prepare(sql)
	if err != nil {
		log.Fatal(err)
	}
	defer preState.Close()
	_, err = preState.Query(folder)
	if err != nil {
		log.Fatal(err)
	}*/
}

// GET /photos/thumbnail/{id}
func ThumbnailServeHandler(w http.ResponseWriter, r *http.Request) {
	photoID := chi.URLParam(r, "photoID")

	sql := `SELECT thumbnail FROM photo WHERE photo_id = ?`
	row := DB.QueryRow(sql, photoID)

	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	fullPath := ""
	_ = row.Scan(&fullPath)
	fullPath = currentWorkingDirectory + string(os.PathSeparator) + fullPath
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		// path/to/whatever does not exist
		fmt.Println("Does not exist")
	}
	http.ServeFile(w, r, fullPath)
}

// GET /photos/info/{id}
func PhotoInfoHandler(w http.ResponseWriter, r *http.Request) {
	photoID := chi.URLParam(r, "photoID")

	sql := `SELECT bytes,filename FROM photo WHERE photo_id = ?`
	row := DB.QueryRow(sql, photoID)
	var bytes int
	var filename string
	row.Scan(&bytes, &filename)

	response := PhotoInfoResponse{Bytes: bytes, Filename: filename}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println(string(responseJSON))
	w.Write(responseJSON)
}

// GET /photos
func PhotosHandler(w http.ResponseWriter, r *http.Request) {
	sql := `SELECT photo_id,full_path,filename,bytes,date_taken,parent_folder FROM photo ORDER BY parent_folder DESC,date_taken DESC`
	rows, err := DB.Query(sql)
	if err != nil {
		log.Println(err)
	}

	photoID := 0
	fullPath := ""
	filename := ""
	bytes := 0
	dateTaken := time.Now()
	parentFolder := ""
	lastParentFolder := "!"

	for rows.Next() {
		err := rows.Scan(&photoID, &fullPath, &filename, &bytes, &dateTaken, &parentFolder)
		if err != nil {
			log.Fatal(err)
		}

		w.Write([]byte(`<style>h1 {color:white;} img {
			object-fit: cover;
			width: 320px;
			height: 320px;
			margin: 2px 2px 2px 2px;
		  }</style>`))
		w.Write([]byte("<body bgcolor='black'>"))
		if lastParentFolder != parentFolder {
			w.Write([]byte("<h1>" + parentFolder + "</h1>"))
		}
		w.Write([]byte(fmt.Sprintf("<a href=\"/photos/%d\"><img src=\"/photos/thumbnail/%d\" title=\"%s\"></a>", photoID, photoID, fullPath)))
		w.Write([]byte("</body>"))

		lastParentFolder = parentFolder
	}
}

func monthNumberToName(month int) string {
	return time.Month(month).String()
}

// GET /photos/all
func GetAllPhotosHandler(w http.ResponseWriter, r *http.Request) {
	sql := `SELECT strftime("%Y",date_taken) AS year,strftime("%m",date_taken) AS month, strftime("%d",date_taken) AS day,photo_id FROM photo ORDER BY date_taken DESC`
	rows, err := DB.Query(sql)
	if err != nil {
		log.Println(err)
	}

	row := 0

	photoID := 0
	year := 0
	month := 0
	day := 0

	lastYear := 0
	lastMonth := 0
	lastDay := 0

	var currentPhotoYear *PhotoYear
	photoYears := []*PhotoYear{}
	photos := []int{}

	for rows.Next() {
		err := rows.Scan(&year, &month, &day, &photoID)
		if err != nil {
			log.Fatal(err)
		}

		photos = append(photos, photoID)
		if lastYear != year {
			photoYear := &PhotoYear{Year: year, Months: []*PhotoMonth{}}
			photoYears = append(photoYears, photoYear)
			currentPhotoYear = photoYears[len(photoYears)-1]
		}
		if lastMonth != month || lastYear != year {
			photoMonth := &PhotoMonth{Month: monthNumberToName(month), Days: []*PhotoDay{}}
			currentPhotoYear.Months = append(currentPhotoYear.Months, photoMonth)
		}
		if lastDay != day || lastMonth != month || lastYear != year {
			photoDay := &PhotoDay{Day: day, Photos: photos}
			currentPhotoYear.Months[len(currentPhotoYear.Months)-1].Days = append(currentPhotoYear.Months[len(currentPhotoYear.Months)-1].Days, photoDay)
			photos = []int{}
		}

		lastYear = year
		lastMonth = month
		lastDay = day

		row++
	}
	response := PhotoResponse{PhotoYears: photoYears}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println(string(responseJSON))
	w.Write(responseJSON)
}
