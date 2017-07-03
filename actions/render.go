package actions

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/bscott/golangflow/models"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/packr"
	"github.com/gobuffalo/plush"
	"github.com/markbates/pop"
	uuid "github.com/satori/go.uuid"
)

var r *render.Engine

func init() {
	r = render.New(render.Options{
		// HTML layout to be used for all HTML requests:
		HTMLLayout: "application.html",

		// Box containing all of the templates:
		TemplatesBox: packr.NewBox("../templates"),

		// Add template helpers here:
		Helpers: render.Helpers{
			"getAvatar": func(id uuid.UUID, help plush.HelperContext) (string, error) {
				tx := help.Value("tx").(*pop.Connection)
				p := models.Post{}
				err := tx.Find(&p, id)
				if err != nil {
					return "", err
				}
				u := models.User{}
				erru := tx.Find(&u, p.UserID)
				if erru != nil {
					return "http://via.placeholder.com/140x100", nil
				}
				return u.GravatarID.String, nil
			},
			"getUser": func(id uuid.UUID, help plush.HelperContext) (string, error) {
				tx := help.Value("tx").(*pop.Connection)
				p := models.Post{}
				err := tx.Find(&p, id)
				if err != nil {
					return " ", err
				}
				u := models.User{}
				erru := tx.Find(&u, p.UserID)
				if erru != nil {
					return "User Not Found", erru
				}
				return u.Name, nil
			},
			"ownsPost": func(post *models.Post, help plush.HelperContext) (template.HTML, error) {
				if cu := help.Value("current_user_id"); cu != nil {
					if post.UserID == cu.(uuid.UUID) && help.HasBlock() {
						s, err := help.Block()
						return template.HTML(s), err
					}
				}
				return "", nil
			},
			"ghProfileURL": func(id uuid.UUID, help plush.HelperContext) (string, error) {
				tx := help.Value("tx").(*pop.Connection)
				p := models.Post{}
				err := tx.Find(&p, id)
				if err != nil {
					return " ", err
				}
				u := models.User{}
				erru := tx.Find(&u, p.UserID)
				if erru != nil {
					return "User Not Found", erru
				}
				un, errg := getGHUser(u.ProviderUserid)

				if errg != nil {
					return " ", errg
				}
				url := "https://github.com/" + un
				return url, nil
			},
		},
	})
}

// GHUser is a Github User struct
type GHUser struct {
	Login             string      `json:"login"`
	ID                int         `json:"id"`
	AvatarURL         string      `json:"avatar_url"`
	GravatarID        string      `json:"gravatar_id"`
	URL               string      `json:"url"`
	HTMLURL           string      `json:"html_url"`
	FollowersURL      string      `json:"followers_url"`
	FollowingURL      string      `json:"following_url"`
	GistsURL          string      `json:"gists_url"`
	StarredURL        string      `json:"starred_url"`
	SubscriptionsURL  string      `json:"subscriptions_url"`
	OrganizationsURL  string      `json:"organizations_url"`
	ReposURL          string      `json:"repos_url"`
	EventsURL         string      `json:"events_url"`
	ReceivedEventsURL string      `json:"received_events_url"`
	Type              string      `json:"type"`
	SiteAdmin         bool        `json:"site_admin"`
	Name              string      `json:"name"`
	Company           string      `json:"company"`
	Blog              string      `json:"blog"`
	Location          string      `json:"location"`
	Email             interface{} `json:"email"`
	Hireable          interface{} `json:"hireable"`
	Bio               string      `json:"bio"`
	PublicRepos       int         `json:"public_repos"`
	PublicGists       int         `json:"public_gists"`
	Followers         int         `json:"followers"`
	Following         int         `json:"following"`
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
}

func getGHUser(s string) (string, error) {

	url := "https://api.github.com/user/" + s
	response, err := http.Get(url)
	if err != nil {
		return " ", err
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return " ", err
	}
	pfl := GHUser{}
	err = json.Unmarshal(body, &pfl)
	if err != nil {
		fmt.Println("whoops:", err)
	}
	return pfl.Login, nil
}
