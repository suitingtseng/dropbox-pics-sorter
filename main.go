package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/files"
)

const (
	CAMERA_UPLOADS = "/Camera Uploads"
)

var (
	token   string
	verbose bool
)

func init() {
	flag.StringVar(&token, "token", "", "Dropbox API token, required")
	flag.BoolVar(&verbose, "verbose", false, "Verbose, optional, default: false")
	flag.Parse()
	if len(token) == 0 {
		flag.Usage()
		os.Exit(1)
	}
}

func main() {
	dbx := NewDropbox(token, verbose)

	res, err := dbx.Ls(CAMERA_UPLOADS)
	if err != nil {
		os.Exit(2)
	}

	chan_time := make(chan time.Time)
	chan_finish := make(chan int)

	go Mkdir(dbx, chan_time, chan_finish)

	mv_args := make([]MvArg, 0)
	for _, entry := range (*res).Entries {
		meta, ok := entry.(*files.FileMetadata)
		if !ok || !isImage(meta.PathLower) {
			continue
		}
		dir, file := path.Split(meta.PathLower)
		date_string := meta.ClientModified.Format(FORMAT)
		to_path := path.Join(dir, date_string, file)
		mv_args = append(mv_args, MvArg{
			src:  meta.PathLower,
			dest: to_path,
		})
		// try to create folder
		chan_time <- meta.ClientModified
	}

	// make sure we terminate the goroutine for Mkdir
	close(chan_time)
	created_folders := <-chan_finish
	fmt.Printf("Created %d folders...\n", created_folders)

	err = dbx.MvBatch(mv_args)
	if err != nil {
		fmt.Printf("MvBatch error: %s\n", err.Error())
	}
	fmt.Printf("Moved %d files...\n", len(mv_args))
}
