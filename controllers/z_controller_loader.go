package controllers

// LoadControllers add the controllers in this method
func (z *Hub) LoadControllers() {
	z.Controllers = make([]interface{}, 0)
	z.Controllers = append(z.Controllers, &PageController{})
}
