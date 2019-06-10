package stl

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

type XPFilePathImpl struct {
	absolutePath   string
	exists         bool
	nameWithExt    string
	nameWithOutExt string
	ext            string
	dir            *XPFilePathImpl
	isFile         *bool
}

func NewFilePath(p string) (*XPFilePathImpl, error) {
	path := new(XPFilePathImpl)
	absPath, err := filepath.Abs(p)
	if err != nil {
		return nil, err
	}
	path.absolutePath   = filepath.ToSlash(absPath)
	path.nameWithExt    = filepath.Base(absPath)
	path.ext            = filepath.Ext(path.nameWithExt)
	path.nameWithOutExt = path.nameWithExt[:len(path.nameWithExt)-len(path.ext)]
	path.Update()
	return path, nil
}

func NewFilePathFromCurrentPath() (*XPFilePathImpl, error) {
	path := new(XPFilePathImpl)
	file, _ := exec.LookPath(os.Args[0])
	absPath, err := filepath.Abs(file)
	if err != nil {
		return nil, err
	}
	path.absolutePath   = filepath.ToSlash(absPath)
	path.nameWithExt    = filepath.Base(absPath)
	path.ext            = filepath.Ext(path.nameWithExt)
	path.nameWithOutExt = path.nameWithExt[:len(path.nameWithExt)-len(path.ext)]
	path.Update()
	return path, nil
}

func (p *XPFilePathImpl) Update() {
	stat, err := os.Stat(p.absolutePath)
	if err != nil {
		return
	}
	p.exists = true
	isf, isd := stat.Mode().IsRegular(), stat.Mode().IsDir()
	if !isf && !isd {
		return
	}
	p.isFile = &isf
	if *p.isFile {
		p.dir, _ = NewFilePath(p.absolutePath[:len(p.absolutePath)-len(p.nameWithExt)])
	} else {
		p.dir = p
	}
}

func (p *XPFilePathImpl) Info() string {
	info := fmt.Sprintf(
		"Absolute path: %s\nExists: %t\n",
		p.AbsolutePath(),
		p.Exists(),
	)
	if p.Exists() {
		if !*p.isFile {
			info += fmt.Sprint("Type: Directory\n")
			return info
		}
		info += fmt.Sprintf("Type: File\nExtension: %s\n", p.Extension())
	}
	return info
}

func (p *XPFilePathImpl) AbsolutePath() string {
	return p.absolutePath
}

func (p *XPFilePathImpl) Exists() bool {
	return p.exists
}

func (p *XPFilePathImpl) Directory() *XPFilePathImpl {
	return p.dir
}

func (p *XPFilePathImpl) NameWithExtension() string {
	return p.nameWithExt
}

func (p *XPFilePathImpl) Name() string {
	return p.nameWithExt
}

func (p *XPFilePathImpl) NameWithoutExtension() string {
	return p.nameWithOutExt
}

func (p *XPFilePathImpl) Extension() string {
	return p.ext
}

func (p *XPFilePathImpl) IsFile() *bool {
	return p.isFile
}

func (p *XPFilePathImpl) IsDirectory() *bool {
	if p.isFile == nil {
		return nil
	}
	isDir := !*p.isFile
	return &isDir
}

func (p *XPFilePathImpl) Append(sp string) *XPFilePathImpl {
	result, _ := NewFilePath(filepath.Join(p.absolutePath, sp))
	return result
}

func (p *XPFilePathImpl) ReadFile() (content []byte, err error) {
	if p.Exists() {
		if *p.IsFile() {
			return ioutil.ReadFile(p.AbsolutePath())
		}
		return nil, NewErrors("path object must be of file type")
	}
	return nil, NewErrors("path object refers to non-existing entity")
}
