package main

import (
	"html/template"
	"io"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Contains html templates.
type Templates struct {
	templates *template.Template
}

// Used by an echo instance to send html responses.
func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

// Creates a new set of templates parsed from html files.
func newTemplate() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}
}

// Contains information for a single contact.
type Contact struct {
	Name  string
	Email string
	Id    int
}

// Creates a new contact.
func newContact(name, email string, id int) Contact {
	return Contact{
		Name:  name,
		Email: email,
		Id:    id,
	}
}

// Wrapper type for a list of contacts.
type ContactList struct {
	AutoIncrementedId int
	Contacts          []Contact
}

// Creates a new data struct with some placeholder contacts.
func newData() ContactList {
	cl := ContactList{
		AutoIncrementedId: 0,
		Contacts:          []Contact{},
	}

	cl.addContact("John", "jd@gmail.com")
	cl.addContact("Clara", "cd@gmail.com")
	return cl
}

// Adds a new contact to the contact list in-place, and returns the new contact.
func (d *ContactList) addContact(name, email string) Contact {
	d.AutoIncrementedId++
	contact := newContact(name, email, d.AutoIncrementedId)
	d.Contacts = append(d.Contacts, contact)
	return contact
}

// Removes a contact from the contact list in-place with the provided id, and returns whether it was found and removed.
func (d *ContactList) removeContact(id int) bool {
	index := d.indexOf(id)
	if index == -1 {
		return false
	}

	// Removes the contact from the list.
	d.Contacts = append(d.Contacts[:index], d.Contacts[index+1:]...)

	return true
}

// Checks if the provided email is in the list of contacts.
func (d *ContactList) hasEmail(email string) bool {
	for _, contact := range d.Contacts {
		if contact.Email == email {
			return true
		}
	}
	return false
}

// Gets the array index of the contact with the provided id, or -1 if not found.
func (d *ContactList) indexOf(id int) int {
	for i, contact := range d.Contacts {
		if contact.Id == id {
			return i
		}
	}
	return -1
}

// Contains key-value maps of form fields and form values, and any associated errors with those fields.
type FormData struct {
	Values map[string]string
	Errors map[string]string
}

// Initializes a new, empty set of form value maps and error maps.
func newFormData() FormData {
	return FormData{
		Values: make(map[string]string),
		Errors: make(map[string]string),
	}
}

// Wrapper type for all related data for the page.
type Page struct {
	Data ContactList
	Form FormData
}

// Creates a new page object.
func newPage() Page {
	return Page{
		Data: newData(),
		Form: newFormData(),
	}
}

// Entry point for the web app.
func main() {
	e := echo.New()
	e.Renderer = newTemplate()
	e.Use(middleware.Logger())
	e.Static("/images", "images")
	e.Static("/css", "css")

	page := newPage()

	e.GET("/", func(c echo.Context) error {
		return c.Render(200, "index", page)
	})

	e.POST("/contacts", func(c echo.Context) error {
		name := c.FormValue("name")
		email := c.FormValue("email")

		if page.Data.hasEmail(email) {
			formData := newFormData()
			formData.Values["name"] = name
			formData.Values["email"] = email
			formData.Errors["email"] = "Email already exists"
			return c.Render(422, "form", formData)
		}

		contact := page.Data.addContact(name, email)

		err := c.Render(200, "form", newFormData())
		if err != nil {
			return err
		}

		return c.Render(200, "oob-contact", contact)
	})

	e.DELETE("/contacts/:id", func(c echo.Context) error {
		// Simulate a particularly slow/difficult database call.
		time.Sleep(3 * time.Second)

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.String(400, "Invalid id")
		}

		success := page.Data.removeContact(id)
		if !success {
			return c.String(404, "Contact not found")
		}

		return c.NoContent(200)
	})

	e.Logger.Fatal((e.Start(":3000")))
}
