package main

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/async"
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
	config := dropbox.Config{Token: token}
	if verbose {
		config.LogLevel = dropbox.LogInfo
	}
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
	_, err := dbx.client.CreateFolderV2(arg)
	return err
}

func (dbx *Dbx) Ls(path string, limit int) ([]LsResult, error) {
	arg := &files.ListFolderArg{
		Path:                            path,
		Recursive:                       false,
		IncludeMediaInfo:                true,
		IncludeDeleted:                  false,
		IncludeHasExplicitSharedMembers: false,
		Limit:                           100,
	}

	res, err := dbx.client.ListFolder(arg)
	if err != nil {
		return nil, err
	}
	ls_results := make([]LsResult, 0, limit)

	for {
		for _, entry := range (*res).Entries {
			meta, ok := entry.(*files.FileMetadata)
			if !ok {
				continue
			}
			ls_results = append(ls_results, LsResult{
				path:         meta.PathLower,
				lastModified: meta.ClientModified,
			})
		}
		if len(ls_results) >= limit {
			break
		}
		if res.HasMore {
			longpoll_arg := files.NewListFolderContinueArg(res.Cursor)
			res, err = dbx.client.ListFolderContinue(longpoll_arg)
			if err != nil {
				return nil, err
			}
		} else {
			break
		}
	}

	return ls_results[:limit], nil
}

// only try to launch a batch move file;
// does not check the job status.
func (dbx *Dbx) MvBatch(mv_args []MvArg) (string, error) {
	reloc_paths := [](*files.RelocationPath){}
	for _, arg := range mv_args {
		reloc_paths = append(reloc_paths, files.NewRelocationPath(arg.src, arg.dest))
	}
	dbx_mv_arg := files.NewMoveBatchArg(reloc_paths)
	res, err := dbx.client.MoveBatchV2(dbx_mv_arg)
	if err != nil {
		return "", err
	}
	return res.AsyncJobId, err
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

func (dbx *Dbx) CheckAsyncJobStatus(jobId string) bool {
	res, err := dbx.client.MoveBatchCheckV2(async.NewPollArg(jobId))
	if err != nil {
		return false
	}
	return res.Tag == files.RelocationBatchV2JobStatusComplete
}
