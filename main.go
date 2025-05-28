package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

var (
	nfsDir = "./nfs"
)

type FS struct{}

func (f *FS) Root() (fs.Node, error) {
	return &Dir{realPath: nfsDir}, nil
}

type Dir struct {
	realPath string
}

var _ fs.Node = (*Dir)(nil)
var _ fs.HandleReadDirAller = (*Dir)(nil)
var _ fs.NodeStringLookuper = (*Dir)(nil)

func (d *Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Mode = os.ModeDir | 0555
	return nil
}

func (d *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	entries, err := os.ReadDir(d.realPath)
	if err != nil {
		return nil, err
	}
	var dirents []fuse.Dirent
	for _, entry := range entries {
		var dtype fuse.DirentType
		if entry.IsDir() {
			dtype = fuse.DT_Dir
		} else {
			dtype = fuse.DT_File
		}
		dirents = append(dirents, fuse.Dirent{
			Name: entry.Name(),
			Type: dtype,
		})
	}
	return dirents, nil
}

func (d *Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	full := filepath.Join(d.realPath, name)
	fi, err := os.Stat(full)
	if err != nil {
		return nil, fuse.ENOENT
	}
	if fi.IsDir() {
		return &Dir{realPath: full}, nil
	}
	return &File{realPath: full}, nil
}

type File struct {
	realPath string
}

var _ fs.Node = (*File)(nil)
var _ fs.HandleReader = (*File)(nil)

func (f *File) Attr(ctx context.Context, a *fuse.Attr) error {
	info, err := os.Stat(f.realPath)
	if err != nil {
		return err
	}
	a.Mode = 0444
	a.Size = uint64(info.Size())
	return nil
}

func (f *File) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	file, err := os.Open(f.realPath)
	if err != nil {
		return err
	}
	defer file.Close()

	buf := make([]byte, req.Size)
	n, err := file.ReadAt(buf, req.Offset)
	if err != nil && err.Error() != "EOF" {
		return err
	}
	resp.Data = buf[:n]
	return nil
}

func main() {
	mountpoint := "./mnt"

	if err := os.MkdirAll(mountpoint, 0755); err != nil {
		log.Fatal(err)
	}

	c, err := fuse.Mount(
		mountpoint,
		fuse.ReadOnly(),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	err = fs.Serve(c, &FS{})
	if err != nil {
		log.Fatal(err)
	}
}
