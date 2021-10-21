package wkhtml

import (
	"log"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/bingoohuang/wkp/pkg/mount"
	"github.com/bingoohuang/wkp/pkg/util"
	"github.com/bingoohuang/wkp/pkg/uuid"
)

var (
	registry *mount.FileRegistry
	mountDir string
	v2pOnce  sync.Once
)

func InitMount() (*mount.FileRegistry, string) {
	mntDir, err := util.TempDir()
	if err != nil {
		log.Fatalf("failed to create temporary dir: %v", err)
	}

	log.Printf("start mount dir: %v", mntDir)

	r := mount.NewFileRegistry()

	go func() {
		if err := mount.Mount(r, mntDir); err != nil {
			log.Fatalf("failed to mount: %v", err)
		}
	}()

	return r, mntDir
}

func (p *ToX) ToPdfV2p(htmlURL, extraArgs string, saveFile bool) (pdf []byte, err error) {
	v2Once.Do(func() { v2Pool = NewV2Pool(p) })
	v2pOnce.Do(func() { registry, mountDir = InitMount() })

	name := uuid.New().String() + ".pdf"
	out := filepath.Join(mountDir, name)
	dataCh, cancelFunc := registry.Register(name)
	defer cancelFunc()

	if err = p.SendArgs(htmlURL, extraArgs, out); err == nil {
		bytes := <-dataCh
		if saveFile {
			return []byte(out), nil
		}

		return bytes, nil
	}

	return nil, err
}

func (p *ToX) SendArgs(htmlURL, extraArgs, out string) error {
	in := p.CacheDirArg() + strconv.Quote(htmlURL) + " " + out + "\n"
	if extraArgs != "" {
		in = extraArgs + " " + in
	}

	wk := v2Pool.borrow()
	result, err := wk.Send(in)
	log.Printf("wk result: %s", result)
	if err == ErrTimeout {
		v2Pool.back(nil)
		if e := wk.Kill("timeout"); e != nil {
			log.Printf("failed to kill, error: %v", e)
		}
	} else {
		v2Pool.back(wk)
	}

	return err
}
