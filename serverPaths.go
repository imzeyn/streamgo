package streamgo

import (
	"strings"
)

func (sp *serverPaths) getTempPathFullName(key string) string {
	name := ""
	
	if sp.temps[key].Parent != "" {
		name += sp.getTempPathFullName(sp.temps[key].Parent) + "/"
	}

	if !strings.HasPrefix(name, "/"){
		name = "/" + name
	}
	
	return ClearURL(name + sp.temps[key].Path.Name)
}