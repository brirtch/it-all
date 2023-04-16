package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/rwcarlsen/goexif/exif"
	"golang.org/x/image/draw"
)

// Creates a thumbnail of sourceFile (JPG only at the moment)
// and returns the thumbnail path.
func createThumbnail(sourceFile string) string {
	input, _ := os.Open(sourceFile)
	defer input.Close()

	filename := path.Base(sourceFile)
	destFile := "data" + string(os.PathSeparator) + filename + ".thumb.jpg"
	output, _ := os.Create(destFile)
	defer output.Close()

	// Decode the image (from PNG to image.Image):
	src, _ := jpeg.Decode(input)

	// Set the expected size that you want:
	dst := image.NewRGBA(image.Rect(0, 0, src.Bounds().Max.X/2, src.Bounds().Max.Y/2))

	// Resize:
	draw.NearestNeighbor.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)

	// Encode to `output`:
	jpeg.Encode(output, dst, nil)

	return destFile
}

func getDateTaken(fname string) time.Time {
	f, err := os.Open(fname)
	if err != nil {
		log.Fatal(err)
	}

	x, err := exif.Decode(f)
	if err != nil {
		log.Fatal(err)
	}

	dateTaken, _ := x.Get(exif.DateTime)
	dateTakenStr, _ := dateTaken.StringVal()

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

	sql := `SELECT rowid,folder FROM photo_folder WHERE state = 'PENDING_SCAN'`
	rows, err := db.Query(sql)
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

				if strings.ToLower(path.Ext(thePath)) == ".jpg" {
					createThumbnail(thePath)
					dateTaken := getDateTaken(thePath)
					fmt.Println(dateTaken)
					pathElements := strings.Split(thePath, string(os.PathSeparator))
					sql := `INSERT INTO photo(full_path,filename,bytes,parent_folder,date_taken) VALUES(?,?,?,?,?)`

					parentFolder := ""
					if len(pathElements) > 2 {
						parentFolder = pathElements[len(pathElements)-2]
					} else {
						parentFolder = ""
					}
					_, err = db.Exec(sql, thePath, path.Base(thePath), info.Size(), parentFolder, dateTaken)
					if err != nil {
						log.Println(err)
					}

					fmt.Println(thePath, info.Size())
				}

				return nil
			})
		if err != nil {
			log.Println(err)
		}
		sql := `UPDATE photo_folder SET state = 'SCANNED' WHERE rowid = ?`
		_, err = db.Exec(sql, id)
	}

}
