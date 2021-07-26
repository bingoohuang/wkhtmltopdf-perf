package mount

import (
	"context"
	"github.com/bingoohuang/wkp/pkg/util"
	"github.com/seaweedfs/fuse"
	"github.com/seaweedfs/fuse/fs"
	"log"
	"os"
	"sync"
	"syscall"
)

type FS struct {
	fr *FileRegistry
}

var _ = fs.FS(&FS{})

func (f *FS) Root() (fs.Node, error) {
	n := &Dir{
		fr: f.fr,
	}
	return n, nil
}

type FileRegistry struct {
	Registry map[string]chan []byte
	Lock     sync.Mutex
}

func NewFileRegistry() *FileRegistry {
	return &FileRegistry{Registry: map[string]chan []byte{}}
}

func (f *FileRegistry) Register(name string) (chan []byte, func()) {
	f.Lock.Lock()
	defer f.Lock.Unlock()

	c := make(chan []byte, 1)
	f.Registry[name] = c
	return c, func() {
		f.Lock.Lock()
		defer f.Lock.Unlock()

		delete(f.Registry, name)
	}
}

func (f *FileRegistry) Send(name string, data []byte) bool {
	f.Lock.Lock()
	defer f.Lock.Unlock()

	if c, ok := f.Registry[name]; ok {
		c <- data
		return true
	}

	return false
}

func Mount(fr *FileRegistry, mountpoint string) error {
	c, err := fuse.Mount(mountpoint)
	if err != nil {
		return err
	}
	defer c.Close()

	if err := fs.Serve(c, &FS{fr: fr}); err != nil {
		return err
	}

	// check if the mount process has an error to report
	<-c.Ready
	if err := c.MountError; err != nil {
		return err
	}

	return nil
}

type Dir struct {
	fr *FileRegistry
}

var _ = fs.Node(&Dir{})

func (d *Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Mode = os.ModeDir | 0755
	return nil
}

var _ = fs.NodeCreater(&Dir{})

func (d *Dir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	f := &File{
		fr:      d.fr,
		dir:     d,
		name:    req.Name,
		writers: 1,
		// file is empty at Create time, no need to set data
	}
	return f, f, nil
}

type File struct {
	fr   *FileRegistry
	dir  *Dir
	name string

	mu sync.Mutex
	// number of write-capable handles currently open
	writers uint
	// only valid if writers > 0
	data []byte
}

func (f *File) Attr(ctx context.Context, a *fuse.Attr) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	a.Mode = 0644
	a.Size = uint64(len(f.data))
	return nil
}

var _ = fs.NodeOpener(&File{})

func (f *File) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	if req.Flags.IsReadOnly() { // we don't need to track read-only handles
		return f, nil
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	f.writers++
	return f, nil
}

var _ = fs.HandleReleaser(&File{})

func (f *File) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	if req.Flags.IsReadOnly() { // we don't need to track read-only handles
		return nil
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	f.writers--
	if f.writers == 0 {
		f.data = nil
	}
	return nil
}

var _ = fs.HandleWriter(&File{})

const maxInt = int(^uint(0) >> 1)

func (f *File) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// expand the buffer if necessary
	newLen := req.Offset + int64(len(req.Data))
	if newLen > int64(maxInt) {
		return fuse.Errno(syscall.EFBIG)
	}
	if v := int(newLen); v > len(f.data) {
		f.data = append(f.data, make([]byte, v-len(f.data))...)
	}

	n := copy(f.data[req.Offset:], req.Data)
	resp.Size = n
	return nil
}

var _ = fs.HandleFlusher(&File{})

func (f *File) Flush(ctx context.Context, req *fuse.FlushRequest) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.writers == 0 {
		// Read-only handles also get flushes. Make sure we don't
		// overwrite valid file contents with a nil buffer.
		return nil
	}

	t, err := util.TempFile(".pdf")
	if err != nil {
		return err
	}

	f.fr.Send(f.name, f.data)

	if err := os.WriteFile(t, f.data, os.ModePerm); err != nil {
		return err
	}

	log.Printf("write file %s to %s", f.name, t)

	return nil
}
