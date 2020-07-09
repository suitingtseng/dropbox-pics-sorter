package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
)

const (
	CAMERA_UPLOADS = "/Camera Uploads"
	IMAGE_BASE     = "/Camera Uploads"
	VIDEO_BASE     = "/Videos"
)

var (
	token   string
	limit   int
	verbose bool
)

func init() {
	flag.StringVar(&token, "token", "", "Dropbox API token, required")
	flag.IntVar(&limit, "limit", 100, "How many files to create folders for and move in a single execution, default: 100")
	flag.BoolVar(&verbose, "verbose", false, "Verbose, optional, default: false")
	flag.Parse()
	if len(token) == 0 {
		flag.Usage()
		os.Exit(1)
	}
}

func main() {
	dbx := NewDropbox(token, verbose)

	res, err := dbx.Ls(CAMERA_UPLOADS, limit)
	if err != nil {
		log.Fatalf("Erorr from dbx.Ls: %s\n", err.Error())
	}

	chan_mkdirarg := make(chan MkdirArg)
	chan_finish := make(chan int)

	go Mkdir(dbx, chan_mkdirarg, chan_finish)

	mv_args := make([]MvArg, 0)
	for _, ls_result := range res {
		base_dir := ""
		if isImage(ls_result.path) {
			base_dir = IMAGE_BASE
		} else if isVideo(ls_result.path) {
			base_dir = VIDEO_BASE
		} else {
			continue
		}
		_, file := path.Split(ls_result.path)
		date_string := ls_result.lastModified.Format(FORMAT)
		to_path := path.Join(base_dir, date_string, file)
		mv_args = append(mv_args, MvArg{
			src:  ls_result.path,
			dest: to_path,
		})
		// try to create folder
		chan_mkdirarg <- MkdirArg{
			base: base_dir,
			date: ls_result.lastModified,
		}
	}

	// make sure we terminate the goroutine for Mkdir
	close(chan_mkdirarg)
	created_folders := <-chan_finish
	fmt.Printf("Created %d folders...\n", created_folders)

	err = dbx.MvBatch(mv_args)
	if err != nil {
		fmt.Printf("MvBatch error: %s\n", err.Error())
	}
	fmt.Printf("Moved %d files...\n", len(mv_args))
}
