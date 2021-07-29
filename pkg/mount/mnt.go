package mount

import (
	"context"
	"os"
	"sync"
	"syscall"

	"github.com/seaweedfs/fuse"
	"github.com/seaweedfs/fuse/fs"
)

type FS struct {
	fr *FileRegistry
}

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
	return c.MountError
}

type Dir struct {
	fr *FileRegistry
}

func (d *Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Mode = os.ModeDir | 0o755
	return nil
}

var _ = fs.NodeCreater(&Dir{})

func (d *Dir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	f := &File{
		fr:   d.fr,
		name: req.Name,
	}
	return f, f, nil
}

type File struct {
	fr   *FileRegistry
	name string
	data []byte
}

func (f *File) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Mode = 0o644
	a.Size = uint64(len(f.data))
	return nil
}

const maxInt = int(^uint(0) >> 1)

func (f *File) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
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

func (f *File) Flush(ctx context.Context, req *fuse.FlushRequest) error {
	f.fr.Send(f.name, f.data)
	return nil
}
