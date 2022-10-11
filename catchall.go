package mat

func (rg *RouteGroup) CatchAll(handlers ...Handler) *RouteGroup {
	rg.router.catchAll.Insert(rg.prefix, handlers)
	return rg
}
