package template

import (
	"fmt"
	"log"

	"github.com/martinlindhe/feng/value"
	"gopkg.in/yaml.v2"
)

// structure of a evaluated ./templates/ yaml file. see Template for the raw structure corresponding to the yaml file
type DataStructure struct {

	// evaluated constants
	Constants []EvaluatedConstant

	// evaluated file structs
	structs []evaluatedStruct

	Layout []value.DataField

	Magic []Magic

	// default endian
	Endian string

	// if template lacks magic bytes
	NoMagic bool

	// extensions
	Extensions []string

	// lastpath/filename-without-ext, eg "archives/zip"
	BaseName string
}

func UnmarshalTemplateIntoDataStructure(b []byte, basename string) (*DataStructure, error) {
	var template Template
	err := yaml.Unmarshal([]byte(b), &template)
	if err != nil {
		return nil, err
	}

	ds, err := NewDataStructureFrom(&template, basename)
	if err != nil {
		return nil, err
	}

	if len(ds.Layout) == 0 {
		return nil, fmt.Errorf("no layout section found in template")
	}

	return ds, err
}

func NewDataStructureFrom(template *Template, basename string) (*DataStructure, error) {
	if DEBUG {
		log.Println("NewDataStructureFrom", basename)
	}
	constants, err := template.evaluateConstants()
	if err != nil {
		return nil, err
	}
	structs, err := template.evaluateStructs()
	if err != nil {
		return nil, err
	}
	layout, err := template.evaluateLayout()
	if err != nil {
		return nil, err
	}

	return &DataStructure{
		Constants:  constants,
		structs:    structs,
		Layout:     layout,
		Magic:      template.Magic,
		NoMagic:    template.NoMagic,
		Endian:     template.Endian,
		Extensions: template.Extensions,
		BaseName:   basename,
	}, nil
}

// looks up layout name from sections
func (ds *DataStructure) FindStructure(df *value.DataField) (*evaluatedStruct, error) {
	for _, str := range ds.structs {
		if df.Kind == str.Name {
			return &str, nil
		}
	}
	return nil, fmt.Errorf("not found in structs: '%s'", df.Kind)
}

/*
// parses a comma-separated string of constants and integers
func (ds *DataStructure) ParsePattern(in, kind string) ([][]byte, error) {
	res := [][]byte{}
	allIn := strings.Split(in, ",")
	for _, part := range allIn {
		part = strings.TrimSpace(part)

		v, ok := ds.FindConstant(part)

		if ok {
			res = append(res, v)
			continue
		}

		i, err := strconv.ParseUint(part, 10, 64)
		if err != nil {
			return nil, err
		}
		res = append(res, value.U64toBytesBigEndian(i, value.SingleUnitSize(kind)))
	}
	if DEBUG {
		log.Printf("ParsePattern: %s %v => %v", kind, allIn, res)
	}
	return res, nil
}

// returns value and true if found
func (ds *DataStructure) FindConstant(name string) ([]byte, bool) {
	if DEBUG {
		log.Printf("FindConstant: looking for %s", name)
	}
	for _, c := range ds.Constants {
		if c.Field.Label == name {
			if DEBUG {
				log.Printf("FindConstant: found %v", c)
			}
			return c.Value, true
		}
	}
	return nil, false
}
*/
