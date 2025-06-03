package main

//#cgo CFLAGS: -x objective-c
//#cgo LDFLAGS: -framework AppKit
//#import <AppKit/AppKit.h>
//__attribute__((used))
//static Protocol *__force_protocol_load() {
//  id o = @protocol(NSApplicationDelegate);
//   return @protocol(NSAccessibility);
//}
import "C"

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/ebitengine/purego/objc"
)

var protocols = []string{
	"NSAccessibility",
	"NSApplicationDelegate",
}

type ProtocolImpl struct {
	Name string

	RequiredInstanceMethods []objc.MethodDescription
	RequiredObjectMethods   []objc.MethodDescription
	OptionalInstanceMethods []objc.MethodDescription
	OptionalObjectMethods   []objc.MethodDescription

	AdoptedProtocols []*objc.Protocol

	RequiredInstanceProperties []objc.Property
	RequiredClassProperties    []objc.Property
	OptionalInstanceProperties []objc.Property
	OptionalClassProperties    []objc.Property
}

func readProtocols(names []string) (imps []ProtocolImpl, err error) {
	for _, name := range names {
		p := objc.GetProtocol(name)
		if p == nil {
			return nil, fmt.Errorf("protocol '%s' does not exist", name)
		}
		imp := ProtocolImpl{}
		imp.Name = name
		imp.RequiredInstanceMethods = p.CopyMethodDescriptionList(true, true)
		imp.RequiredObjectMethods = p.CopyMethodDescriptionList(true, false)
		imp.OptionalInstanceMethods = p.CopyMethodDescriptionList(false, true)
		imp.OptionalObjectMethods = p.CopyMethodDescriptionList(false, false)

		imp.AdoptedProtocols = p.CopyProtocolList()

		imp.RequiredInstanceProperties = p.CopyPropertyList(true, true)
		imp.RequiredClassProperties = p.CopyPropertyList(true, false)
		imp.OptionalInstanceProperties = p.CopyPropertyList(false, true)
		imp.OptionalClassProperties = p.CopyPropertyList(false, false)
		imps = append(imps, imp)
	}
	return imps, nil
}

func main() {
	var imps []ProtocolImpl
	var err error
	if imps, err = readProtocols(protocols); err != nil {
		panic(err)
	}
	if err = printProtocols(imps, os.Stdout); err != nil {
		panic(err)
	}
}

const templ = `package main

import (
	"log"

	"github.com/ebitengine/purego/objc"
)

func init() {
	var p *objc.Protocol
	{{- range . }}
	{{- $protocolName := .Name }}

	p = objc.AllocateProtocol("{{$protocolName}}")
	if p != nil {
		{{- range .RequiredInstanceMethods }}
		p.AddMethodDescription(objc.RegisterName("{{ .Name }}"), "{{ .Types }}", true, true)
		{{- end }}
		
		{{- range .RequiredObjectMethods }}
		p.AddMethodDescription(objc.RegisterName("{{ .Name }}"), "{{ .Types }}", true, false)
		{{- end }}
		
		{{- range .OptionalInstanceMethods }}
		p.AddMethodDescription(objc.RegisterName("{{ .Name }}"), "{{ .Types }}", false, true)
		{{- end }}
		
		{{- range .OptionalObjectMethods }}
		p.AddMethodDescription(objc.RegisterName("{{ .Name }}"), "{{ .Types }}", false, false)
		{{- end }}
		var adoptedProtocol *objc.Protocol
		{{- range .AdoptedProtocols }}
		adoptedProtocol = objc.GetProtocol("{{ .Name }}")
		if adoptedProtocol == nil {
			log.Fatalln("protocol '{{ .Name }}' does not exist")
		}
		p.AddProtocol(adoptedProtocol)
		{{- end }}

		{{- range .RequiredInstanceProperties }}
		p.AddProperty("{{ .Name }}", {{ attributeToStructString . }}, true, true)
		{{- end }}


		{{- range .RequiredClassProperties }}
		p.AddProperty("{{ .Name }}", {{ attributeToStructString . }}, true, false)
		{{- end }}

		{{- range .OptionalInstanceProperties }}
		p.AddProperty("{{ .Name }}", {{ attributeToStructString . }}, false, true)
		{{- end }}

		{{- range .OptionalClassProperties }}
		p.AddProperty("{{ .Name }}", {{ attributeToStructString . }}, false, false)
		{{- end }}
		p.Register()
	} // Finished protocol: {{$protocolName}}
	{{- end }}
}
`

func printProtocols(impls []ProtocolImpl, w io.Writer) error {
	tmpl, err := template.New("protocol.tmpl").Funcs(template.FuncMap{
		"attributeToStructString": attributeToStructString,
	}).Parse(templ)
	if err != nil {
		return err
	}
	err = tmpl.Execute(w, impls)
	if err != nil {
		return err
	}
	return nil
}

func attributeToStructString(p objc.Property) string {
	attribs := strings.Split(p.Attributes(), ",")
	var b strings.Builder
	b.WriteString("[]objc.PropertyAttribute{")
	for i, attrib := range attribs {
		b.WriteString(fmt.Sprintf(`{Name: &[]byte("%s\x00")[0], Value: &[]byte(%s)[0]}`, string(attrib[0]), strconv.Quote(attrib[1:]+"\x00")))
		if i != len(attribs)-1 {
			b.WriteString(", ")
		}
	}
	b.WriteString("}")
	return b.String()
}
