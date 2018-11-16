package gluster

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gluster/gogfapi/gfapi"
	"gopkg.in/src-d/go-billy.v4"
)

var _ billy.Basic = new(GlusterFS)

type GlusterFS struct {
	v *gfapi.Volume
}

func New(host, volume string) (*GlusterFS, error) {
	vol := new(gfapi.Volume)
	err := vol.Init(volume, host)
	if err != nil {
		return nil, err
	}

	err = vol.Mount()
	if err != nil {
		return nil, err
	}

	g := &GlusterFS{v: vol}
	return g, nil
}

func (g *GlusterFS) Close() error {
	return g.v.Unmount()
}

const (
	strNoFileOrDir = "no such file or directory"

	defaultDirectoryMode = 0755
	defaultCreateMode    = 0666
)

func (g *GlusterFS) createDir(fullpath string) error {
	dir := filepath.Dir(fullpath)
	if dir != "." {
		if err := g.MkdirAll(dir, defaultDirectoryMode); err != nil {
			return err
		}
	}

	return nil
}

func (g *GlusterFS) Create(filename string) (billy.File, error) {
	if err := g.createDir(filename); err != nil {
		return nil, err
	}

	f, err := g.v.Create(filename)
	if err != nil {
		return nil, err
	}

	return NewFile(filename, f, os.O_RDWR), nil
}

func (g *GlusterFS) Open(filename string) (billy.File, error) {
	return g.OpenFile(filename, os.O_RDONLY, 0)
}

func (g *GlusterFS) OpenFile(filename string, flag int, perm os.FileMode) (billy.File, error) {
	if flag&os.O_CREATE != 0 {
		if err := g.createDir(filename); err != nil {
			return nil, err
		}

		// O_CREATE does not create the file. Here Create is used if we can
		// not find it. This could be done in a more efficient way by reusing
		// the created file descriptor instead of reopening with the specific
		// flags in some cases.
		_, err := g.Stat(filename)
		if err != nil {
			c, err := g.v.Create(filename)
			if err != nil {
				return nil, err
			}
			if err = c.Close(); err != nil {
				return nil, err
			}

			// Setting permissions in OpenFile is not supported. Change it
			// manually with Chmod.
			err = g.v.Chmod(filename, perm)
			if err != nil {
				return nil, err
			}
		}
	}

	f, err := g.v.OpenFile(filename, flag, perm)
	if err != nil {
		return nil, err
	}

	return NewFile(filename, f, flag), nil
}

func (g *GlusterFS) Stat(filename string) (os.FileInfo, error) {
	return g.v.Stat(filename)
}

func (g *GlusterFS) Rename(oldpath string, newpath string) error {
	return g.v.Rename(oldpath, newpath)
}

func (g *GlusterFS) Remove(filename string) error {
	return g.v.Unlink(filename)
}

func (g *GlusterFS) Join(elem ...string) string {
	return filepath.Join(elem...)
}

func (g *GlusterFS) ReadDir(path string) ([]os.FileInfo, error) {
	return nil, fmt.Errorf("ReadDir not implemented")
}

func (g *GlusterFS) MkdirAll(filename string, perm os.FileMode) error {
	return g.v.MkdirAll(filename, perm)
}