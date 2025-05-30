package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

var (
	nfsDir = "./nfs"
	ssdDir = "./ssd"
	cache  = NewLRUCache(10, 100*1024) // 10 files or 100 KB
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
		log.Printf("❌ Failed to read directory %s: %v", d.realPath, err)
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
		log.Printf("❌ Lookup failed for %s: %v", full, err)
		return nil, fuse.ENOENT
	}
	if fi.IsDir() {
		return &Dir{realPath: full}, nil
	}
	relPath, _ := filepath.Rel(nfsDir, full)
	return &File{virtualPath: relPath}, nil
}

type File struct {
	virtualPath string
}

var _ fs.Node = (*File)(nil)
var _ fs.HandleReader = (*File)(nil)

func hashFile(path string) (string, []byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", nil, err
	}
	defer f.Close()

	h := sha256.New()
	data, err := io.ReadAll(f)
	if err != nil {
		return "", nil, err
	}
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil)), data, nil
}

func (f *File) Attr(ctx context.Context, a *fuse.Attr) error {
	nfsPath := filepath.Join(nfsDir, f.virtualPath)
	if info, err := os.Stat(nfsPath); err == nil {
		a.Mode = 0444
		a.Size = uint64(info.Size())
		return nil
	} else {
		log.Printf("❌ NFS stat failed for %s: %v", nfsPath, err)
	}
	return fuse.ENOENT
}

func (f *File) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	nfsPath := filepath.Join(nfsDir, f.virtualPath)

	hash, content, err := hashFile(nfsPath)
	if err != nil {
		log.Printf("❌ Failed to hash NFS file: %s: %v", nfsPath, err)
		return err
	}
	ssdPath := filepath.Join(ssdDir, hash)

	if _, err := os.Stat(ssdPath); os.IsNotExist(err) {
		log.Printf("📥 Cache miss: %s (hash: %s) → reading from NFS with delay...", f.virtualPath, hash)
		time.Sleep(500 * time.Millisecond)

		if err := os.WriteFile(ssdPath, content, 0644); err != nil {
			log.Printf("❌ Failed to write to SSD: %s: %v", ssdPath, err)
			return err
		}
		cache.Touch(hash, len(content))
		log.Printf("✅ Copied %s to SSD cache (hash: %s)", f.virtualPath, hash)
	} else {
		cache.Touch(hash, len(content))
		log.Printf("⚡ Cache hit: %s (hash: %s)", f.virtualPath, hash)
	}

	file, err := os.Open(ssdPath)
	if err != nil {
		log.Printf("❌ Failed to open SSD file %s: %v", ssdPath, err)
		return err
	}
	defer file.Close()

	buf := make([]byte, req.Size)
	n, err := file.ReadAt(buf, req.Offset)
	if err != nil && err.Error() != "EOF" {
		log.Printf("❌ Failed to read from SSD file %s: %v", ssdPath, err)
		return err
	}
	resp.Data = buf[:n]
	return nil
}

func main() {
	mountpoint := "./mnt/all-projects"

	if err := os.MkdirAll(mountpoint, 0755); err != nil {
		log.Fatalf("❌ Failed to create mountpoint: %v", err)
	}
	if err := os.MkdirAll(ssdDir, 0755); err != nil {
		log.Fatalf("❌ Failed to create SSD cache dir: %v", err)
	}

	c, err := fuse.Mount(mountpoint, fuse.ReadOnly())
	if err != nil {
		log.Fatalf("❌ Failed to mount FUSE: %v", err)
	}
	defer c.Close()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		log.Println("📦 Caught signal — unmounting...")
		if err := fuse.Unmount(mountpoint); err != nil {
			log.Printf("⚠️  Failed to unmount: %v", err)
		}
		os.Exit(0)
	}()

	log.Printf("✅ FUSE filesystem mounted at %s", mountpoint)

	if err := fs.Serve(c, &FS{}); err != nil {
		log.Fatalf("❌ Serve error: %v", err)
	}
}
