package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

// Contains fields and errors for a new contact form.
type FormData struct {
	Name   string
	Email  string
	Errors []string
}

// Creates a new form data object.
func newFormData() FormData {
	return FormData{
		Name:   "",
		Email:  "",
		Errors: []string{},
	}
}

// Contains information for a single contact.
type Contact struct {
	Name  string
	Email string
	Id    int
}

// Creates a new contact object.
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

// Wrapper type for all related data for the page.
type Page struct {
	Form        FormData
	ContactList ContactList
}

// Creates a new page object.
func newPage() Page {
	return Page{
		Form:        newFormData(),
		ContactList: newData(),
	}
}

// Renders a templ component onto an echo context.
func Render(ctx echo.Context, statusCode int, t templ.Component) error {
	buf := templ.GetBuffer()
	defer templ.ReleaseBuffer(buf)

	if err := t.Render(ctx.Request().Context(), buf); err != nil {
		return err
	}

	return ctx.HTML(statusCode, buf.String())
}

// Entry point for the web server.
func main() {
	e := echo.New()
	e.Static("/images", "images")
	e.Static("/css", "css")

	page := newPage()

	e.GET("/", func(c echo.Context) error {
		return Render(c, http.StatusOK, index(page))
	})

	e.POST("/contacts", func(c echo.Context) error {
		name := c.FormValue("name")
		email := c.FormValue("email")

		if page.ContactList.hasEmail(email) {
			formData := newFormData()
			formData.Name = name
			formData.Email = email
			formData.Errors = append(formData.Errors, "Email already exists")
			return Render(c, http.StatusUnprocessableEntity, form(formData))
		}

		contact := page.ContactList.addContact(name, email)

		err := Render(c, http.StatusOK, form(newFormData()))
		if err != nil {
			return err
		}

		return Render(c, http.StatusOK, oobContact(contact))
	})

	e.DELETE("/contacts/:id", func(c echo.Context) error {
		// Simulate a particularly slow/difficult database call.
		time.Sleep(3 * time.Second)

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid id")
		}

		success := page.ContactList.removeContact(id)
		if !success {
			return c.String(http.StatusNotFound, "Contact not found")
		}

		return c.NoContent(http.StatusOK)
	})

	e.Logger.Fatal((e.Start(":3000")))
}
