package clover

import (
	"encoding/xml"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/codeclimate/test-reporter/env"
	"github.com/codeclimate/test-reporter/formatters"
	"github.com/pkg/errors"
)

var searchPaths = []string{"build/logs/clover.xml", "clover.xml"}

type Formatter struct {
	Path string
}

func (f *Formatter) Search(paths ...string) (string, error) {
	paths = append(paths, searchPaths...)
	for _, p := range paths {
		logrus.Debugf("checking search path %s for clover formatter", p)
		if _, err := os.Stat(p); err == nil {
			f.Path = p
			return p, nil
		}
	}

	return "", errors.WithStack(errors.Errorf("could not find any files in search paths for clover. search paths were: %s", strings.Join(paths, ", ")))
}

func (r Formatter) Format() (formatters.Report, error) {
	rep, err := formatters.NewReport()
	if err != nil {
		return rep, err
	}

	fx, err := os.Open(r.Path)
	if err != nil {
		return rep, errors.WithStack(err)
	}

	c := &xmlFile{}
	err = xml.NewDecoder(fx).Decode(c)
	if err != nil {
		return rep, errors.WithStack(err)
	}

	gitHead, _ := env.GetHead()
	for _, pp := range c.Packages {
		for _, pf := range pp.Files {
			num := 1
			sf, err := formatters.NewSourceFile(pf.Name, gitHead)
			if err != nil {
				return rep, errors.WithStack(err)
			}
			for _, l := range pf.Lines {
				for num < l.Num {
					sf.Coverage = append(sf.Coverage, formatters.NullInt{})
					num++
				}
				ni := formatters.NewNullInt(l.Count)
				sf.Coverage = append(sf.Coverage, ni)
				num++
			}
			err = rep.AddSourceFile(sf)
			if err != nil {
				return rep, errors.WithStack(err)
			}
		}
	}

	return rep, nil
}
