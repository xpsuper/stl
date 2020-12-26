package excel

import (
	"reflect"
	"strings"
)

const (
	tagIdentify = "xlsx"
	tagSplit    = ";"

	columnTag  = "column"
	splitTag   = "split"
	defaultTag = "default"
	nilTag     = "nil"
	ignoreTag  = "-"
	reqTag     = "req"
)

type FieldConfig struct {
	// The config equals to tag: column
	ColumnName string
	// The config equals to tag: default
	DefaultValue string
	// The config equals to tag: split
	Split string
	// The config equals to tag: nil
	// if cell.value == NilValue, will skip fc scan
	NilValue string
	// The config equals to tag: req
	// panic if reuqired fc column but not set
	IsRequired bool
	// The config equals to tag: -
	Ignore bool
}

func (this *FieldConfig) froze(fieldIdx int) *fieldConfig {
	return &fieldConfig{
		FieldIndex:   fieldIdx,
		ColumnName:   this.ColumnName,
		DefaultValue: this.DefaultValue,
		Split:        this.Split,
		NilValue:     this.NilValue,
		IsRequired:   this.IsRequired,
	}
}

type ExcelFiledConfiger interface {
	GetXLSXFieldConfigs() map[string]FieldConfig
}

type fieldConfig struct {
	FieldIndex int
	// use ptr in order to know if configed.
	ColumnName   string
	DefaultValue string
	Split        string
	// if cell.value == NilValue, will skip fc scan
	NilValue string
	// panic if reuqired fc column but not set
	IsRequired bool
}

func (fc *fieldConfig) scan(valStr string, fieldValue reflect.Value) error {
	if fc.NilValue == valStr {
		// log.Printf("Got nil,skip")
		return nil
	}
	var err error
	switch fieldValue.Kind() {
	case reflect.Slice, reflect.Array:
		if len(fc.Split) != 0 && len(valStr) > 0 {
			// use split
			elems := strings.Split(valStr, fc.Split)
			fieldValue.Set(reflect.MakeSlice(fieldValue.Type(), 0, len(elems)))
			err = scanSlice(elems, fieldValue.Addr())
		}
	case reflect.Ptr:
		newValue := fieldValue
		if newValue.IsNil() {
			for newValue.Kind() == reflect.Ptr {
				newValue.Set(reflect.New(newValue.Type().Elem()))
				newValue = newValue.Elem()
			}
		}
		err = scan(valStr, newValue.Addr().Interface())
	default:
		err = scan(valStr, fieldValue.Addr().Interface())
	}
	return err
}

func (fc *fieldConfig) ScanDefault(fieldValue reflect.Value) error {
	err := fc.scan(fc.DefaultValue, fieldValue)
	if err != nil && len(fc.DefaultValue) > 0 {
		return err
	}
	return nil
}

type schema struct {
	Type reflect.Type
	// map[FieldIndex]*Field
	Fields []*fieldConfig
}

func newSchema(t reflect.Type) *schema {
	s := &schema{
		Fields: make([]*fieldConfig, 0, t.NumField()),
	}

	// if implement the ExcelFiledConfiger
	var selfDefinedCfgs map[string]FieldConfig
	v := reflect.New(t)
	if v.CanInterface() {
		if i, ok := v.Interface().(ExcelFiledConfiger); ok {
			selfDefinedCfgs = i.GetXLSXFieldConfigs()
		}
	} else if vElem := v.Elem(); vElem.CanInterface() {
		if i, ok := vElem.Interface().(ExcelFiledConfiger); ok {
			selfDefinedCfgs = i.GetXLSXFieldConfigs()
		}
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if selfCfg, ok := selfDefinedCfgs[field.Name]; ok {
			// Use self defiend config first
			if !selfCfg.Ignore {
				frzCfg := selfCfg.froze(i)
				if frzCfg.ColumnName == "" {
					frzCfg.ColumnName = field.Name
				}
				s.Fields = append(s.Fields, frzCfg)
			}
		} else if value, ok := field.Tag.Lookup(tagIdentify); ok {
			// Use tag second
			if value != ignoreTag {
				fieldCnf := praseTagValue(value)
				fieldCnf.FieldIndex = i
				if fieldCnf.ColumnName == "" {
					fieldCnf.ColumnName = field.Name
				}
				s.Fields = append(s.Fields, fieldCnf)
			}
		} else {
			// use default config
			fieldCnf := &fieldConfig{
				FieldIndex: i,
				ColumnName: field.Name,
			}
			s.Fields = append(s.Fields, fieldCnf)
		}
	}
	s.Type = t
	return s
}

func praseTagValue(v string) *fieldConfig {
	c := &fieldConfig{}
	params := strings.Split(v, tagSplit)

	for _, param := range params {
		if param == "" {
			continue
		}
		cnfKey, cnfVal := getTagParam(param)
		fillField(c, cnfKey, cnfVal)
	}
	// with more params
	return c
}

func getTagParam(v string) (key, value string) {
	// expect v = `field_name` or `column(fieldName)` or `default(0)` and so on ...
	start := strings.Index(v, "(")
	end := strings.Index(v, ")")
	if start > 0 && end == len(v)-1 {
		return v[:start], v[start+1 : end]
	}
	// log.Printf("Use column as default?[%s]\n", v)
	return columnTag, v
}

func fillField(c *fieldConfig, k, v string) {
	switch k {
	case columnTag:
		c.ColumnName = v
	case defaultTag:
		c.DefaultValue = v
	case splitTag:
		c.Split = v
	case nilTag:
		c.NilValue = v
	case reqTag:
		c.IsRequired = true
	}
}
