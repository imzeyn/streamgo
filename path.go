package streamgo

func (p *Path) MethodAllowed(method HTTPMethod) bool {
	return p.AllowedMethods[method]
}