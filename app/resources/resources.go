package resources

import (
	"log"
	"net/http"

	"github.com/qor/qor"
	"github.com/qor/qor/admin"
	"github.com/qor/qor/roles"

	. "github.com/qor/qor-example/app/models"
	. "github.com/qor/qor-example/db"
	"github.com/qor/qor/i18n"
	"github.com/qor/qor/i18n/backends/database"
)

var (
	Admin *admin.Admin
	I18n  *i18n.I18n
)

func init() {
	// setting up QOR admin
	Admin = admin.New(&qor.Config{DB: Pub.DraftDB()})
	Admin.AddResource(Pub)
	Admin.SetAuth(&Auth{})

	I18n = i18n.New(database.New(StagingDB))
	Admin.AddResource(I18n)

	roles.Register("admin", func(req *http.Request, currentUser qor.CurrentUser) bool {
		if currentUser == nil {
			return false
		}

		if currentUser.(*User).Role == "admin" {
			return true
		}

		return false
	})

	roles.Register("user", func(req *http.Request, currentUser qor.CurrentUser) bool {
		if currentUser == nil {
			return false
		}

		if currentUser.(*User).Role == "user" {
			return true
		}

		return false
	})

	user := Admin.AddResource(&User{})

	user.Meta(&admin.Meta{
		Name:  "UserRole",
		Label: "Role",
		//Type:  "select_one",
		Collection: func(resource interface{}, context *qor.Context) (results [][]string) {
			return [][]string{
				{"admin", "admin"},
				{"user", "user"},
			}
		},
		Valuer: func(value interface{}, context *qor.Context) interface{} {
			if value == nil {
				return value
			}
			user := value.(*User)
			log.Println("user", user.Role)
			return user.Role
		},
	})

	user.IndexAttrs("ID", "Name", "Role")
	user.NewAttrs("Name", "UserRole")
	user.EditAttrs("Name", "UserRole")

	author := Admin.AddResource(&Author{})

	author.IndexAttrs("ID", "Name")
	author.SearchAttrs("ID", "Name")

	book := Admin.AddResource(&Book{})

	book.Meta(&admin.Meta{
		Name:  "FormattedDate",
		Label: "Release Date",
		Valuer: func(value interface{}, context *qor.Context) interface{} {
			book := value.(*Book)
			return book.ReleaseDate.Format("Jan 2, 2006")
		},
	})

	book.Meta(&admin.Meta{
		Name: "Synopsis",
		Type: "rich_editor",
	})

	// which fields should be displayed in the books list on admin
	book.IndexAttrs("ID", "Title", "Authors", "FormattedDate", "Price")
	// which fields should be editable in the book esit interface
	book.NewAttrs("Title", "Authors", "Synopsis", "ReleaseDate", "Price", "CoverImage")
	book.EditAttrs("Title", "Authors", "Synopsis", "ReleaseDate", "Price", "CoverImage")
	// which fields should be searched when using the item search on the list view
	book.SearchAttrs("ID", "Title")
}

type Auth struct{}

func (Auth) LoginURL(c *admin.Context) string {
	return "/login"
}

func (Auth) LogoutURL(c *admin.Context) string {
	return "/logout"
}

func (Auth) GetCurrentUser(c *admin.Context) qor.CurrentUser {
	if userid, err := c.Request.Cookie("userid"); err == nil {
		var user User
		if !DB.First(&user, "id = ?", userid.Value).RecordNotFound() {
			return &user
		}
	}
	return nil
}
