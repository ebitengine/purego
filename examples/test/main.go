package main

//#cgo CFLAGS: -x objective-c
//#cgo LDFLAGS: -framework AppKit
//#import <AppKit/AppKit.h>
//__attribute__((used))
//static Protocol *__force_protocol_load() {
//   return @protocol(NSAccessibility);
//}
//import "C"

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/ebitengine/purego/objc"
)

type ProtocolImpl struct {
	Name string

	RequiredInstanceMethods []objc.MethodDescription
	RequiredObjectMethods   []objc.MethodDescription
	OptionalInstanceMethods []objc.MethodDescription
	OptionalObjectMethods   []objc.MethodDescription

	AdoptedProtocols []*objc.Protocol

	RequiredInstanceProperties []objc.Property
}

func readProtocol(name string) (imp ProtocolImpl, err error) {
	p := objc.GetProtocol(name)
	if p == nil {
		return ProtocolImpl{}, fmt.Errorf("protocol does not exist")
	}
	imp.Name = name
	imp.RequiredInstanceMethods = p.CopyMethodDescriptionList(true, true)
	imp.RequiredObjectMethods = p.CopyMethodDescriptionList(true, false)
	imp.OptionalInstanceMethods = p.CopyMethodDescriptionList(false, true)
	imp.OptionalObjectMethods = p.CopyMethodDescriptionList(false, false)

	imp.AdoptedProtocols = p.CopyProtocolList()
	//    // Properties
	//    objc_property_t *props = protocol_copyPropertyList(p, &count);
	//    for (unsigned int i = 0; i < count; i++) {
	//        const char *name = property_getName(props[i]);
	//        const char *attrs = property_getAttributes(props[i]);
	//        printf("// Property '%s' with attributes '%s'\n", name, attrs);
	//        // You can parse attributes and call protocol_addProperty if needed
	//    }
	//    free(props);
	//
	//    printf("objc_registerProtocol(proto);\n");
	imp.RequiredInstanceProperties = p.CopyPropertyList(true, true)
	//for _, p := range imp.RequiredInstanceProperties {
	//	fmt.Println(p.Attributes())
	//}
	return imp, nil
}

func main() {
	if p, err := readProtocol("NSAccessibility"); err != nil {
		panic(err)
	} else {
		printProtocol([]ProtocolImpl{p})
	}
}

const templ = `
package main

import "github.com/ebitengine/purego/objc"

func init() {
	var p *objc.Protocol
	{{- range . }}
	{{- $protocolName := .Name }}
	
	// Begin Objective-C protocol definition for: {{$protocolName}}
	p = objc.AllocateProtocol("{{$protocolName}}")
	if p != nil { // only register if it doesn't exist
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

		p.Register()
		// Finished protocol: {{$protocolName}}
	}
	{{ end }}
}
`

func printProtocol(impls []ProtocolImpl) {
	tmpl, err := template.New("protocol.tmpl").Funcs(template.FuncMap{
		"attributeToStructString": attributeToStructString,
	}).Parse(templ)
	if err != nil {
		log.Fatal(err)
	}
	err = tmpl.Execute(os.Stdout, impls)
	if err != nil {
		log.Fatal(err)
	}
}

func attributeToStructString(p objc.Property) string {
	attribs := strings.Split(p.Attributes(), ",")
	var b strings.Builder
	b.WriteString("[]objc.PropertyAttribute{")
	for _, attrib := range attribs {
		b.WriteString(fmt.Sprintf(`{ Name: &[]byte("%s\x00")[0], Value: &[]byte(%s+"\x00")[0] },`, string(attrib[0]), strconv.Quote(attrib[1:])))
	}
	b.WriteString("}")
	return b.String()
}
