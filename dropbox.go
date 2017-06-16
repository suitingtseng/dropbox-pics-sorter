package main

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/files"
)

type Dbx struct {
	client files.Client
}

type MvArg struct {
	src  string
	dest string
}

type MkdirArg struct {
	base string
	date time.Time
}

type LsResult struct {
	path         string
	lastModified time.Time
}

func NewDropbox(token string, verbose bool) *Dbx {
	config := dropbox.Config{Token: token, Verbose: verbose}
	dbx_files := files.New(config)
	return &Dbx{
		client: dbx_files,
	}
}

func (dbx *Dbx) Mkdir(path string) error {
	arg := &files.CreateFolderArg{
		Path:       path,
		Autorename: false,
	}
	_, err := dbx.client.CreateFolder(arg)
	return err
}

func (dbx *Dbx) Ls(path string) ([]LsResult, error) {
	arg := &files.ListFolderArg{
		Path:                            path,
		Recursive:                       false,
		IncludeMediaInfo:                true,
		IncludeDeleted:                  false,
		IncludeHasExplicitSharedMembers: false,
	}

	res, err := dbx.client.ListFolder(arg)
	if err != nil {
		return nil, err
	}

	ls_results := make([]LsResult, 0)
	for _, entry := range (*res).Entries {
		meta, ok := entry.(*files.FileMetadata)
		if !ok || !isImage(meta.PathLower) {
			continue
		}
		ls_results = append(ls_results, LsResult{
			path:         meta.PathLower,
			lastModified: meta.ClientModified,
		})
	}
	return ls_results, nil
}

// only try to launch a batch move file;
// does not check the job status.
func (dbx *Dbx) MvBatch(mv_args []MvArg) error {
	dbx_mv_arg := &files.RelocationBatchArg{
		Entries:           [](*files.RelocationPath){},
		AllowSharedFolder: false,
		Autorename:        false,
	}
	for _, arg := range mv_args {
		entry := &files.RelocationPath{
			FromPath: arg.src,
			ToPath:   arg.dest,
		}
		dbx_mv_arg.Entries = append(dbx_mv_arg.Entries, entry)
	}
	_, err := dbx.client.MoveBatch(dbx_mv_arg)
	return err
}

func Mkdir(dbx *Dbx, c chan MkdirArg, finish chan int) {
	dir_set := make(DirSet)
	created := 0
	for mkdir_arg := range c {
		t := mkdir_arg.date
		base_dir := mkdir_arg.base
		if dir_set.Contains(mkdir_arg) {
			continue
		}
		dir_set.Add(mkdir_arg)
		err := dbx.Mkdir(path.Join(base_dir, t.Format(FORMAT)))
		if err != nil {
			if !strings.Contains(err.Error(), "conflict") {
				fmt.Printf("Mkdir error: %s\n", err.Error())
			}
		} else {
			created += 1
		}
	}
	finish <- created
}
