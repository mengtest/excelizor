package main

import (
	"fmt"
	"strconv"
	"strings"
)

type Field struct {
	Id          int
	Name        string
	Type        string
	LongType    string
	Data        string
	Count       int
	Size        int
	Layer       int
	Level       int
	Template    *Field
	Fields      []*Field
	ParentField *Field
}

func (f *Field) Init(name string, defination string) (bool, int) {
	if name == "" && defination == "" && f.Name == "" && f.Type == "" {
		return false, -1
	}

	if name != "" {
		f.Name = name
	}
	if defination != "" {
		f.Size = 1
		hasSubField, subFieldDef := f.parseDefination(defination)
		if hasSubField {
			f.Fields = make([]*Field, 0)
		}
		if subFieldDef != "" {
			f.Template = new(Field)
			f.Template.ParentField = f
			if ok, layer := f.Template.Init("", subFieldDef); ok {
				f.Size = f.Template.Size*f.Count + 1
				f.Layer = layer + 1
			}
		}
	}
	return true, f.Layer
}

func (f *Field) ParseSubFieldsDefs(names []string, defs []string) {
	subFieldIndex := 1
	for i := 0; i < len(names); {
		if f.Template == nil {
			f.Template = new(Field)
			f.Template.ParentField = f
			f.Template.Size = 1
		}
		field := f.Template.Copy()
		subFieldName := names[i]
		if f.Type == "list" {
			subFieldName = strconv.Itoa(subFieldIndex)
		}
		if ok, _ := field.Init(subFieldName, defs[i]); ok {
			num := field.Size
			if num > 1 {
				field.ParseSubFieldsDefs(names[i+1:i+num], defs[i+1:i+num])
			}
			f.Fields = append(f.Fields, field)

			i += num
		} else {
			i++
		}
		subFieldIndex++
	}
}

func (f *Field) ParseDatas(id int, datas []string) {
	data := datas[0]
	f.Id = id
	if f.ParentField != nil && f.ParentField.Type == "dict" && strings.TrimSpace(f.Name) == "" {
		nameData := splitName(data)
		f.Name = nameData[0]
		data = nameData[1]
	}

	if f.Count == 0 {
		data = trimData(data)
		subDatas := splitSubData(f.Layer, data)

		f.setSubFieldsData(subDatas)
		//fmt.Println(strconv.Itoa(f.Layer) + " \tname = " + f.Name + " \ttype = " + f.Type + " \tdata = " + f.Data + " \tcount = " + strconv.Itoa(f.Count))
	} else if f.Count == 1 {

		f.Data = data
		f.Data = strings.Replace(f.Data, "\"", "\\\"", -1)
		//fmt.Println(strconv.Itoa(f.Layer) + " \tname = " + f.Name + " \ttyoe = " + f.Type + " \tdata = " + f.Data + " \tcount = " + strconv.Itoa(f.Count))
	} else {
		f.setSubFieldsData(datas)
		//fmt.Println(strconv.Itoa(f.Layer) + " \tname = " + f.Name + " \ttype = " + f.Type + " \tcount = " + strconv.Itoa(f.Count))
	}
}

func (f *Field) setSubFieldsData(data []string) {
	fieldNum := 0
	var offset int
	if f.Count == -1 || f.Count == 0 {
		offset = 0
	} else {
		offset = 1
	}
	for i := offset; i < len(data); {
		if len(f.Fields) <= fieldNum {
			if data[i] == "" {
				i++
				continue
			}
			f.Fields = append(f.Fields, f.Template.Copy())
		}
		size := f.Fields[fieldNum].Size
		subdata := data[i : i+size]
		if f.Type == "list" {
			f.Fields[fieldNum].Name = strconv.Itoa(fieldNum + 1)
		}
		f.Fields[fieldNum].ParseDatas(f.Id, subdata)
		i += size
		fieldNum++
	}
}

func trimData(data string) string {
	before := data
	for {
		data = strings.TrimPrefix(data, "{")
		data = strings.TrimSuffix(data, "}")
		if before == data {
			return data
		}
		before = data
	}

}

func splitName(data string) []string {
	subDatas := make([]string, 0)
	equal := strings.Index(data, "=")

	var n string
	var d string
	if equal == -1 {
		d = data
	} else {
		n = data[:equal]
		d = data[equal+1:]
	}
	subDatas = append(subDatas, n)
	subDatas = append(subDatas, d)
	return subDatas
}

func splitSubData(layer int, data string) []string {
	sept := ""
	for i := 1; i < layer; i++ {
		sept += "}"
	}
	sept += "|"

	subDatas := make([]string, 0)

	for {
		pos := strings.Index(data, sept)
		if pos == -1 {
			subDatas = append(subDatas, data)
			break
		} else {
			subData := data[0 : pos+layer-1]
			data = data[pos+layer:]
			subDatas = append(subDatas, subData)
		}
	}
	return subDatas
}

func (f *Field) parseDefination(def string) (bool, string) {
	first := strings.Index(def, "<")
	last := strings.LastIndex(def, ">:")
	if first != -1 && last != -1 {
		if count, err := strconv.Atoi(def[last+2:]); err == nil {
			f.Type = def[:first]
			f.LongType = def[:last+1]
			f.Count = count
		}
		return true, def[first+1 : last]
	}

	f.Type = def
	f.LongType = def
	f.Count = 1
	f.Size = 1
	if def == "struct" {
		f.Count = -1
		f.Size = -1
		return true, ""
	}

	return false, ""
}

func (f *Field) Copy() *Field {
	field := new(Field)
	field.Id = f.Id
	field.Name = f.Name
	field.Type = f.Type
	field.LongType = f.LongType
	field.Data = f.Data
	field.Count = f.Count
	field.Size = f.Size
	field.Layer = f.Layer
	field.Level = f.Level
	field.ParentField = f.ParentField
	if f.Template != nil {
		field.Template = f.Template.Copy()
	}
	if f.Fields != nil {
		field.Fields = make([]*Field, 0)
		for i := 0; i < len(f.Fields); i++ {
			field.Fields = append(field.Fields, f.Fields[i].Copy())
		}
	}
	return field
}

func (f *Field) SetLevel(level int) {
	f.Level = level
	if f.Fields != nil {
		for j := 0; j < len(f.Fields); j++ {
			f.Fields[j].SetLevel(level + 4)
		}
	}
}

func (f *Field) Print() {
	for i := 0; i < f.Level; i++ {
		fmt.Print("\t")
	}
	fmt.Print(" name = ", f.Name, " | type = ", f.Type, " | long type = ", f.LongType)
	if f.Data != "" {
		fmt.Print(" | data = ", f.Data)
	}
	fmt.Print("\n")
	if f.Fields != nil {
		for j := 0; j < len(f.Fields); j++ {
			f.Fields[j].Print()
		}
	}
}