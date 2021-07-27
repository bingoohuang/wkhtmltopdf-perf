package wkhtml

import (
	"github.com/bingoohuang/wkp/pkg/mount"
	"github.com/bingoohuang/wkp/pkg/util"
	"github.com/bingoohuang/wkp/pkg/uuid"
	"log"
	"path/filepath"
	"strconv"
	"sync"
)

var registry *mount.FileRegistry
var mountDir string
var v2pOnce sync.Once

func InitMount() (*mount.FileRegistry, string) {
	mntDir, err := util.TempDir()
	if err != nil {
		log.Fatalf("failed to create temporary dir: %v", err)
	}

	log.Printf("start mount dir: %v", mntDir)

	registry := mount.NewFileRegistry()

	go func() {
		if err := mount.Mount(registry, mntDir); err != nil {
			log.Fatalf("failed to mount: %v", err)
		}
	}()

	return registry, mntDir
}

func (p *ToX) ToPdfV2p(htmlURL, extraArgs string) (pdf []byte, err error) {
	v2Once.Do(func() { v2Pool = NewV2Pool(p.MaxPoolSize) })
	v2pOnce.Do(func() { registry, mountDir = InitMount() })

	name := uuid.New().String() + ".pdf"
	out := filepath.Join(mountDir, name)
	dataCh, cancelFunc := registry.Register(name)
	defer cancelFunc()

	if err = p.SendArgs(htmlURL, extraArgs, out); err == nil {
		return <-dataCh, nil
	}

	return nil, err
}

func (p *ToX) SendArgs(htmlURL, extraArgs, out string) error {
	in := strconv.Quote(htmlURL) + " " + out + "\n"
	if extraArgs != "" {
		in = extraArgs + " " + in
	}

	wk := v2Pool.borrow()
	result, err := wk.Send(in, "Done", "Error:")
	log.Printf("wk result: %s", result)
	if err == ErrTimeout {
		v2Pool.reportKill()
		if err := wk.Kill("timeout"); err != nil {
			log.Printf("failed to kill, error: %v", err)
		}
	} else {
		v2Pool.back(wk)
	}

	return err
}
